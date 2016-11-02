package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"strings"
	"os"
	"errors"
)

const (
	Version = "0.0.1"
)

func main() {

	fmt.Println("DSocket is a tools to debug socket, ", Version)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: Color(">>> ", Cyan),
		HistoryFile: "/tmp/dsockert.history",
	})

	if err != nil {
		panic(err)
	}

	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		ins := strings.TrimSpace(line)

		fmt.Println("\t", ins)

		if ins == "exit" || ins == "quit" {
			os.Exit(1)
		}

		if ins == "" {
			continue
		}

		cmds, err := ParseInputArgs(ins)

		if err != nil {
			fmt.Println(Color("\tError", Red), err.Error())
		}

		var e error
		switch cmds[0] {
		case "conn", "connect":
			e = connect(cmds[0], cmds[1:])
		case "pack", "packet":
		case "unpack", "unpacket":
		case "send":
		default:
			e = errors.New("unsupport" + cmds[0])
		}
		if e != nil {
			fmt.Println("\tError", e.Error())
		}

	}
}

func ParseInputArgs(line string) (args []string, err error) {
	as := make([]string, 0)
	rs := []rune(line)
	res := make([][]rune, 0)
	sub := make([]rune, 0)
	quote := false
	tm := false
	for _, v := range rs {
		if v == '\\' {
			if tm {
				sub = append(sub, v)
				tm = false
			} else {
				tm = true
			}
		} else if v == '"' {
			if tm {
				sub = append(sub, v)
				tm = false
			} else {
				quote = !quote
				if !quote {
					res = append(res, sub)
					sub = make([]rune, 0)
				}
			}
		} else if v == ' ' {
			if tm {
				sub = append(sub, v)
				tm = false
			} else if quote {
				sub = append(sub, v)
			} else {
				if len(sub) != 0 {
					res = append(res, sub)
					sub = make([]rune, 0)
				}
			}
		} else {
			if quote {
				if tm {
					if v == 't' {
						sub = append(sub, []rune{' ', ' ', ' ', ' '}...)
					} else if v == 'n' {
						sub = append(sub, '\n')
					} else {
						sub = append(sub, v)
					}
					tm = false
				} else {
					sub = append(sub, v)
				}
			} else {
				sub = append(sub, v)
			}
		}
	}
	if len(sub) != 0 {
		res = append(res, sub)
	}

	for _, v := range res {
		as = append(as, string(v))
	}
	args = as
	return
}
