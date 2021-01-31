package packgen

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type strBuilder struct {
	strings.Builder
}

func (b *strBuilder) WriteStringf(format string, a ...interface{}) {
	b.WriteString(fmt.Sprintf(format, a...))
}
func (b *strBuilder) WriteStringfn(format string, a ...interface{}) {
	b.WriteString(fmt.Sprintf(format, a...))
	b.WriteByte('\n')
}

type strVersionaizer struct {
	mp map[string]int
}

func newStrVersionaizer() strVersionaizer {
	return strVersionaizer{
		mp: make(map[string]int),
	}
}
func (v *strVersionaizer) version(str string) string {
	if _, ok := v.mp[str]; !ok {
		v.mp[str] = 0
	}
	v.mp[str] = v.mp[str] + 1
	return fmt.Sprintf("%s%d", str, v.mp[str])
}

type Generator struct {
	*strBuilder
	strVersionaizer
	targetpkg  string
	targetpath string
	pkgs       map[string]string
	structs    map[string]*rootNode
	servers    []*server
	errors     []StructErr
}

func NewGenerator(targetpkg string, targetpath string) *Generator {
	return &Generator{
		strBuilder:      &strBuilder{},
		strVersionaizer: newStrVersionaizer(),
		targetpkg:       targetpkg,
		targetpath:      targetpath,
		pkgs:            make(map[string]string),
		structs:         make(map[string]*rootNode),
		errors:          make([]StructErr, 0),
	}
}

func (g *Generator) addErr(err StructErr) {
	g.errors = append(g.errors, err)
}

func (g *Generator) Errors() []StructErr {
	return g.errors
}

func (g *Generator) addPkg(path string) {
	if path == g.targetpath {
		return
	}
	if _, ok := g.pkgs[path]; !ok {
		g.pkgs[path] = "pkg" + strconv.Itoa(len(g.pkgs))
	}
	return
}

func (g *Generator) getPrefix(path string) string {
	if path == g.targetpath {
		return ""
	}
	return g.pkgs[path] + "."
}

func (g *Generator) Gen() {
	g.genHeader()
	for _, v := range g.structs {
		v.genMarsh("")
		v.genUnmarsh("")
	}
	for _, v := range g.servers {
		v.gen()
	}
}

func (g *Generator) genHeader() {
	g.WriteStringfn("package %s", g.targetpkg)
	g.WriteStringfn("import (")
	g.WriteStringfn(`"bytes"`)
	g.WriteStringfn(`"context"`)
	g.WriteStringfn(`"fmt"`)
	for path, name := range g.pkgs {
		g.WriteStringfn(`%s "%s"`, name, path)
	}
	g.WriteStringfn(`"github.com/Apakhov/ayproto"`)
	g.WriteStringfn(")")
}

func (g *Generator) WriteFiles(dir string) error {
	tmpF, err := os.Create(dir + "/packgen_ayproto.go.temp")
	if err != nil {
		return errors.Wrap(err, "cant open temp file")
	}
	defer tmpF.Close()

	_, err = fmt.Fprintln(tmpF, g.String())
	if err != nil {
		return errors.Wrap(err, "cant write to temp file")
	}

	return nil
}
