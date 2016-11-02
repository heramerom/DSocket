package main

import (
	"flag"
	"net"
	"io"
	"encoding/binary"
	"fmt"
)

var (
	_conn_pool = make(map[string]Session)
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
	go connect_tcp(name, host, port)
	return nil
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

	var sess Session
	sess.conn = conn
	sess.die = make(chan struct{})
	_conn_pool[name] = sess

	go func() {
		header := make([]byte, 2)
		for {
			n, err := io.ReadFull(conn, header)
			if err != nil {
				//TODO
			}
			size := binary.BigEndian.Uint16(n)
			playload := make([]byte, size)
			n, err = io.ReadFull(conn, playload)
			if err != nil {
				//todo
			}
		}
	}()
	return nil
}