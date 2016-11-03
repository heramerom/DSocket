package main

import (
	"testing"
	"fmt"
	"strings"
)

func TestSplitN(t *testing.T) {
	fmt.Println("test...")

	s := "string hello world|int 10|float 3.2"
	args := strings.Split(s, "|")

	for _, value := range args {
		a := strings.SplitN(value, " ", 2)
		fmt.Println(a)
		for _, b := range a {
			fmt.Println(b)
		}
	}
}