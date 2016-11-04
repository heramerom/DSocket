package main

import (
	"flag"
	"net"
	"io"
	"encoding/binary"
	"fmt"
	"errors"
	"log"
	"bytes"
	"strings"
	"strconv"
)

var (
	_conn_pool = make(map[string]Session)
	_unpack_rules = make(map[uint16][]string)
)

type Session struct {
	conn *net.TCPConn
	die  chan struct{}
}

func connect(cmd string, args []string) error {
	var host, name string
	var port int
	f := flag.NewFlagSet(cmd, flag.ContinueOnError)
	f.StringVar(&host, "h", "127.0.0.1", "socket host")
	f.StringVar(&host, "host", "127.0.0.1", "socket port")
	f.StringVar(&name, "n", "default", "server name")
	f.StringVar(&name, "name", "default", "server name")
	f.IntVar(&port, "p", 8888, "socket port")
	f.IntVar(&port, "port", 8888, "socketp port")

	err := f.Parse(args)
	if err != nil {
		return err
	}
	return connect_tcp(name, host, port)
}

func addSession(name string, sess Session) {
	if s, ok := _conn_pool[name]; ok {
		close(s.die)
	}
	_conn_pool[name] = sess
}

func connect_tcp(name, host string, port int) error {
	host_port := fmt.Sprintf("%s:%d", host, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host_port)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}

	addr, p, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return errors.New("error got remote ip address")
	}
	log.Println("success connect server ", addr, ":", p)

	var sess Session
	sess.conn = conn
	sess.die = make(chan struct{})
	addSession(name, sess)

	in := make(chan []byte)

	go recv(&sess, in)

	go func() {
		defer close(sess.die)
		defer close(in)
		defer delete(_conn_pool, name)
		header := make([]byte, 2)
		for {
			_, err := io.ReadFull(conn, header)
			if err != nil {
				log.Println("Error", err)
				log.Println("[ERROR]:", "lost conn with name", name)
				return
			}
			size := binary.BigEndian.Uint16(header)
			playload := make([]byte, size)
			_, err = io.ReadFull(conn, playload)
			if err != nil {
				log.Println("[ERROR]:", err)
				log.Println("[ERROR]:", "lost conn with name", name)
				return
			}
			//in <- playload
			select {
			case in <- playload:

			case <-sess.die:
				return
			}
		}
	}()
	return nil
}

func recv(sess *Session, in chan []byte) {
	for {
		select {
		case msg, ok := <-in:
			if ok {
				unpack(msg)
			}
		case <-sess.die:
			return
		}
	}
}

func unpack(msg []byte) {
	reader := Reader(msg)
	code, err := reader.ReadU16()
	if err != nil {
		log.Println("\tError: ", err.Error())
	}

	rule := _unpack_rules[code]

	sb := bytes.NewBufferString("")

	for _, value := range rule {
		switch value {
		case "int8", "i8":
			v, err := reader.ReadS8()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%d", v))
			}
		case "int16", "i16":
			v, err := reader.ReadS16()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%d", v))
			}
		case "int24", "i24":
			v, err := reader.ReadS24()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%d", v))
			}
		case "int32", "int", "i", "i32":
			v, err := reader.ReadS32()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%d", v))
			}
		case "int64", "i64":
			v, err := reader.ReadS64()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%d", v))
			}
		case "uint16", "u16":
			v, err := reader.ReadU16()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%d", v))
			}
		case "uint32", "u32":
			v, err := reader.ReadU32()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%d", v))
			}
		case "uint64", "u64":
			v, err := reader.ReadU64()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%d", v))
			}
		case "float32", "f", "f32", "float":
			v, err := reader.ReadFloat32()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%f", v))
			}
		case "float64", "f64":
			v, err := reader.ReadFloat64()
			if checkErr(err) {
				sb.WriteString(fmt.Sprintf("%f", v))
			}
		case "string", "s":
			v, err := reader.ReadString()
			if checkErr(err) {
				sb.WriteString(v)
			}
		default:
			log.Println("[ERROR]: unsupport type ", value)
		}
		sb.WriteString(" | ")
	}

	log.Println("[RECV]: ", sb.String())
}

func checkErr(err error) bool {
	if err != nil {
		log.Println("[ERROR]: ", err.Error())
		return false
	}
	return true
}

func addRule(cmd string, args []string) error {
	var rule string
	var code uint
	f := flag.NewFlagSet(cmd, flag.ContinueOnError)
	f.StringVar(&rule, "r", "", "unpack rule")
	f.StringVar(&rule, "rule", "", "unpack rule")
	f.UintVar(&code, "c", 0, "unpack code")
	f.UintVar(&code, "code", 0, "unpack code")
	err := f.Parse(args)
	if err != nil {
		return err
	}
	c := uint16(code)
	_unpack_rules[c] = strings.Split(rule, "|")
	return nil
}

func send(cmd string, args []string) error {

	var data, name string

	f := flag.NewFlagSet(cmd, flag.ContinueOnError)
	f.StringVar(&data, "d", "", "send data")
	f.StringVar(&data, "data", "", "send data")
	f.StringVar(&name, "n", "default", "server name")
	f.StringVar(&name, "name", "default", "server name")

	err := f.Parse(args)
	if err != nil {
		return err
	}

	conn, ok := _conn_pool[name]
	if !ok {
		return errors.New("unknow conn " + name)
	}

	d, err := pack(data)

	if err != nil {
		return err
	}

	n, err := conn.conn.Write(d)

	if err != nil {
		return err
	}
	log.Printf("[SUCCESS]: success to send %d\n", n)
	return nil
}

func pack(data string) ([]byte, error) {

	args := strings.Split(data, "|")

	writer := Writer()

	for _, value := range args {
		kv := strings.SplitN(value, " ", 2)
		if len(kv) == 1 {
			writer.WriteString(kv[0])
			continue
		}
		if len(kv) > 2 {
			return nil, errors.New("error parse")
		}
		switch kv[0] {
		case "int8", "i8":
			v, err := strconv.ParseInt(kv[1], 10, 8)
			if err != nil {
				return nil, err
			}
			writer.WriteS8(int8(v))
		case "int16", "i16":
			v, err := strconv.ParseInt(kv[1], 10, 16)
			if err != nil {
				return nil, err
			}
			writer.WriteS16(int16(v))
		case "int32", "int", "i", "i32":
			v, err := strconv.ParseInt(kv[1], 10, 32)
			if err != nil {
				return nil, err
			}
			writer.WriteS32(int32(v))
		case "int64", "i64":
			v, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return nil, err
			}
			writer.WriteS64(int64(v))

		case "uint16", "u16":
			v, err := strconv.ParseUint(kv[1], 10, 16)
			if err != nil {
				return nil, err
			}
			writer.WriteU16(uint16(v))

		case "uint32", "uint", "u32":
			v, err := strconv.ParseUint(kv[1], 10, 32)
			if err != nil {
				return nil, err
			}
			writer.WriteU32(uint32(v))

		case "uint64", "u64":
			v, err := strconv.ParseUint(kv[1], 10, 64)
			if err != nil {
				return nil, err
			}
			writer.WriteU64(uint64(v))

		case "float", "float32", "f32":
			v, err := strconv.ParseFloat(kv[1], 32)
			if err != nil {
				return nil, err
			}
			writer.WriteFloat32(float32(v))

		case "float64", "f64":
			v, err := strconv.ParseFloat(kv[1], 64)
			if err != nil {
				return nil, err
			}
			writer.WriteFloat64(float64(v))

		case "string", "s":
			writer.WriteString(kv[1])

		default:
			return nil, errors.New("unsupport type " + kv[0])
		}
	}
	w := Writer()
	w.WriteBytes(writer.Data())
	return w.Data(), nil
}