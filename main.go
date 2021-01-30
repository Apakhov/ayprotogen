package main

import (
	"fmt"
	"os"

	"github.com/Apakhov/ayprotogen/bootstrap"
)

func fatal(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	succ := false
	dir := os.Args[1]
	defer func() {
		err := bootstrap.CleadUp(dir, succ)
		fatal(err)
	}()
	name, packets, servers, err := bootstrap.ParseDir(dir)
	fatal(err)
	err = bootstrap.GenBootstrap(dir, name, packets, servers)
	fatal(err)
	err = bootstrap.RunBootstrap(dir)
	fatal(err)
	succ = true
}
