package main

import (
	"fmt"
	"os"

	"github.com/Apakhov/ayprotogen/bootstrap"
	_ "github.com/Apakhov/ayprotogen/packgen"
)

//go:generate go run genfiles/main.go packgen

func exit1OnFatal() {
	if s, ok := recover().(string); ok && s == "" {
		os.Exit(1)
	}
}

func fatal(err error, msg string) {
	if err != nil {
		fmt.Println(msg, ":", err)
		panic("")
	}
}

func main() {
	defer exit1OnFatal()
	dir, err := os.Getwd()
	fatal(err, "opening dir")
	defer func() {
		if len(os.Args) < 2 || os.Args[1] != "leave-temps" {
			bootstrap.CleadUp(dir)
		}
	}()
	name, trg, packets, servers, err := bootstrap.ParseDir(dir)
	fatal(err, "parsing dir")
	err = bootstrap.GenBootstrap(dir, trg, name, packets, servers, gfiles)
	fatal(err, "generating bootstrap")
	err = bootstrap.RunBootstrap(dir)
	fatal(err, "runing bootstrap")
	err = bootstrap.MvTmp(dir)
	fatal(err, "writing new file")
}
