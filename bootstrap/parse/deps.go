package parse

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"go/build"
	"io"
	"os"
	"path"
	"sort"
)

type dependencies map[string][]string // package import path -> file list

func findDependencies(pkg ...string) dependencies {
	ctx := build.Context{
		GOPATH:   os.Getenv("GOPATH"),
		GOROOT:   "", // Make sure we don't find standard libs.
		Compiler: "gc",
	}

	deps := make(dependencies)

	toProcess := pkg
	for len(toProcess) > 0 {
		processing := toProcess
		toProcess = []string{}

		for _, p := range processing {
			wd, _ := os.Getwd()

			pkg, err := ctx.Import(p, wd, 0)
			if err != nil {
				// Only list what we can find. Skip standard libs (due to empty GOROOT).
				continue
			}

			for _, f := range pkg.GoFiles {
				deps[p] = append(deps[p], path.Join(pkg.SrcRoot, p, f))
			}
			for _, path := range pkg.Imports {
				if _, ok := deps[path]; ok {
					continue
				}
				toProcess = append(toProcess, path)
				deps[path] = []string{}
			}
		}
	}

	return deps
}

func (deps dependencies) String() string {
	ret := map[string][]string{}

	var pkgs []string
	for pkg := range deps {
		pkgs = append(pkgs, pkg)
	}

	sort.Strings(pkgs)

	for pkg, pkgDeps := range deps {
		for _, p := range pkgDeps {
			f, err := os.Open(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: unable to open dependency %v: %v\n", f, err)
				continue
			}

			h := sha1.New()
			_, err = io.Copy(h, f)

			f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: unable to read dependency %v: %v\n", f, err)
				continue
			}

			ret[pkg] = append(ret[pkg], fmt.Sprintf("%s:%x", path.Base(p), h.Sum(nil)))
		}
		sort.Strings(ret[pkg])
	}

	marshalled, _ := json.Marshal(ret)
	bytes := sha1.Sum(marshalled)
	return fmt.Sprintf("%x", bytes[:])
}
