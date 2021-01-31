package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Apakhov/ayprotogen/bootstrap"
)

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	dir := os.Args[1]
	infos, err := ioutil.ReadDir(dir)
	fatal(err)

	gfiles := []bootstrap.GFile{}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		file, err := os.Open(dir + "/" + info.Name())
		fatal(err)
		bt, err := ioutil.ReadAll(file)
		fatal(err)
		gfiles = append(gfiles, bootstrap.GFile{
			Name:    file.Name(),
			Content: string(bt),
		})
		file.Close()
	}

	file, err := os.Create("files.go")
	fatal(err)
	_, err = file.WriteString(fmt.Sprintf(`package main

import "github.com/Apakhov/ayprotogen/bootstrap"

var gfiles = %#v`, gfiles))
	fatal(err)
	file.Close()
}
