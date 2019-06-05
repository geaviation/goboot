package util

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
)

//https://github.com/phayes/freeport/blob/master/freeport.go
func FreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func RunBash(script string, env ...string) {
	cmd := exec.Command("/bin/bash", script)
	cmd.Env = append(os.Environ(), env...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("bash$ %s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	err = cmd.Wait()
	if err != nil {
		panic(err)
	}
}
