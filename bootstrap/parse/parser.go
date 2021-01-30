// Package parse contains a simple parser that extracts RPC methods from a
// golang source file.
package parse

import (
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

// Parser extracts method list from a go source file.
type Parser struct {
	Package      string
	ImportPath   string
	Dependencies string
}

// ParseFile runs the parser on a single file.
func (p *Parser) ParseFile(filename string) error {
	pkgdir := filepath.Dir(filename)
	return p.parseFiles(pkgdir, []string{filename})
}

// Parse runs the parser on all files of the package.
func (p *Parser) Parse(pkgdir string) error {
	pkgFiles, err := ioutil.ReadDir(pkgdir)
	if err != nil {
		return err
	}
	files := make([]string, 0, len(pkgFiles))
	for _, f := range pkgFiles {
		if f.IsDir() {
			continue
		}
		fname := f.Name()
		if filepath.Ext(fname) != ".go" {
			continue
		}
		if strings.HasSuffix(fname, "_ayproto.go") {
			continue
		}
		filename := path.Join(pkgdir, fname)
		files = append(files, filename)
	}
	if len(files) == 0 {
		return nil
	}
	return p.parseFiles(pkgdir, files)
}

func (p *Parser) parseFiles(pkgdir string, files []string) error {
	fset := token.NewFileSet()

	var pkgs []string
	for _, filename := range files {
		f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		if p.Package == "" {
			p.Package = f.Name.String()
		}

		pkgs = append(pkgs, f.Name.String())
	}

	p.Dependencies = findDependencies(pkgs...).String()

	dir, err := filepath.Abs(pkgdir)
	if err != nil {
		return fmt.Errorf("cannot get absolute path for %s: %v", pkgdir, err)
	}
	pkg, err := build.ImportDir(dir, 0)
	if err != nil {
		return fmt.Errorf("cannot process directory %s: %s", pkgdir, err)
	}

	isModule, err := isGoMod(dir)
	if err != nil {
		return err
	}

	if !isModule {
		p.ImportPath = pkg.ImportPath
		return nil
	}

	modPath, err := goModPath(dir, true)
	if err != nil {
		return err
	}

	p.ImportPath, err = getPkgPathFromGoMod(dir, true, modPath)
	if err != nil {
		return err
	}

	return nil
}
