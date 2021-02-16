package main

import (
	"fmt"
	"github.com/Shea11012/interpreter_in_go/repl"
	"os"
	user2 "os/user"
)

func main() {
	user, err := user2.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
	fmt.Printf("feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
