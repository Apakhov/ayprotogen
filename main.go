package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Apakhov/ayprotogen/bootstrap"
)

//go:generate go run genfiles/main.go packgen

func fatal(err error, msg string) {
	if err != nil {
		fmt.Println(msg, ":", err)
		os.Exit(1)
	}
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	name, trg, packets, servers, err := bootstrap.ParseDir(dir)
	fatal(err, "parsing dir")
	err = bootstrap.GenBootstrap(dir, trg, name, packets, servers, gfiles)
	fatal(err, "generating bootstrap")
	err = bootstrap.RunBootstrap(dir)
	fatal(err, "runing bootstrap")
	err = bootstrap.MvTmp(dir)
	fatal(err, "writing new file")
	if len(os.Args) < 2 || os.Args[1] != "leave-temps" {
		bootstrap.CleadUp(dir)
	}
}
