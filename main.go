package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Apakhov/ayprotogen/bootstrap"
)

//go:generate go run genfiles/main.go packgen

func fatal(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	succ := false
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if len(os.Args) >= 2 && os.Args[1] == "leave-temps" {
			succ = false
		}
		err := bootstrap.CleadUp(dir, succ)
		fatal(err)
	}()
	name, trg, packets, servers, err := bootstrap.ParseDir(dir)
	fatal(err)
	err = bootstrap.GenBootstrap(dir, trg, name, packets, servers, gfiles)
	fatal(err)
	err = bootstrap.RunBootstrap(dir)
	fatal(err)
	succ = true
}
