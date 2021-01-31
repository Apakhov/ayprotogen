package packgen

import (
	"fmt"
	"log"
	"reflect"
)

type genNode interface {
	base() *nodeBase
	getType() string
	genMarsh(acc string)
	genUnmarsh(acc string)
}

type nodeBase struct {
	g      *Generator
	parent genNode
	childs []genNode
}

func (nb *nodeBase) base() *nodeBase {
	return nb
}

func makeChildNode(node genNode, t reflect.Type) error {
	newBase := newNodeBase(node.base().g)
	var newNode genNode
	switch t.Kind() {
	case reflect.Struct:
		newNode = newStructNode(newBase, t)
		for i := 0; i < t.NumField(); i++ {
			err := makeChildNode(newNode, t.Field(i).Type)
			if err != nil {
				return errorWrap(err, fmt.Sprint("generating ", t))
			}
		}
	case reflect.Slice:
		newNode = newSliceNode(newBase)
		err := makeChildNode(newNode, t.Elem())
		if err != nil {
			return errorWrap(err, fmt.Sprint("generating ", t))
		}
	default:
		var err error
		newNode, err = newSimpleNode(newBase, t)
		if err != nil {
			return errorWrap(err, fmt.Sprint("generating ", t))
		}
	}
	node.base().childs = append(node.base().childs, newNode)
	newNode.base().parent = node
	return nil
}

func newNodeBase(g *Generator) *nodeBase {
	return &nodeBase{
		g:      g,
		childs: make([]genNode, 0),
	}
}

type rootNode struct{ *nodeBase }

func (n *rootNode) getType() string {
	return ""
}

func (n *rootNode) genMarsh(acc string) {
	g := n.g
	g.WriteStringfn("func (v *%s) MarshalAyproto() ([]byte, error) {", n.childs[0].getType())
	g.WriteStringfn("res := make([]byte, 0)")
	n.childs[0].genMarsh("v")
	g.WriteStringfn("return res, nil")
	g.WriteStringfn("}")
	return
}

func (n *rootNode) genUnmarsh(acc string) {
	g := n.g
	g.WriteStringfn("func (v *%s) UnmarshalAyproto(r *bytes.Reader) error {", n.childs[0].getType())
	g.WriteStringfn("var err error")
	g.WriteStringfn("var l uint32") // slice len
	g.WriteStringfn("_,_ = l,err")
	g.WriteStringfn("var int8conv uint8")
	g.WriteStringfn("var int16conv uint16")
	g.WriteStringfn("var int32conv uint32")
	g.WriteStringfn("var int64conv uint64")
	g.WriteStringfn("var intconv uint32")
	g.WriteStringfn("var uintconv uint32")
	g.WriteStringfn("_,_,_,_,_,_ = int8conv, int16conv, int32conv, int64conv, intconv, uintconv")
	n.childs[0].genUnmarsh("v")
	g.WriteStringfn("return nil")
	g.WriteStringfn("}")
	return
}

type structNode struct {
	*nodeBase
	name string
	t    reflect.Type
}

func newStructNode(base *nodeBase, t reflect.Type) *structNode {
	base.g.addPkg(t.PkgPath())
	return &structNode{
		nodeBase: base,
		t:        t,
	}
}

func (n *structNode) getType() string {
	name := n.t.Name()
	if name == "" { // anon struct
		log.Println("cant use anonimus struct")
		panic("anon struct")
	}
	return n.g.getPrefix(n.t.PkgPath()) + name
}

func (n *structNode) genMarsh(acc string) {
	for i, child := range n.childs {
		child.genMarsh(acc + "." + n.t.Field(i).Name)
	}
	return
}

func (n *structNode) genUnmarsh(acc string) {
	for i, child := range n.childs {
		child.genUnmarsh(acc + "." + n.t.Field(i).Name)
	}
	return
}

type sliceNode struct {
	*nodeBase
	name string
}

func newSliceNode(base *nodeBase) *sliceNode {
	return &sliceNode{
		nodeBase: base,
	}
}

func (n *sliceNode) getType() string {
	return "[]" + n.childs[0].getType()
}

func (n *sliceNode) genMarsh(acc string) {
	n.g.WriteStringfn("res = ayproto.PackUint32(res, uint32(len(%s)), 0)", acc)
	n.g.WriteStringfn("for _, v := range %s {", acc)
	n.childs[0].genMarsh("v")
	n.g.WriteStringfn("}")
	return
}

func (n *sliceNode) genUnmarsh(acc string) {
	n.g.WriteStringfn(mpFieldToUnpackSchema[reflect.Uint32], "l")
	n.g.WriteStringfn("if int64(l) > r.Size() {")
	n.g.WriteStringfn(`return fmt.Errorf("cant unpack array - invalid array length %%d in packet of length %%d", l, r.Size())`)
	n.g.WriteStringfn("}")
	n.g.WriteStringfn("%s = make([]%s, l)", acc, n.childs[0].getType())
	iv := n.g.version("i")
	n.g.WriteStringfn("for %s := range %s {", iv, acc)
	n.childs[0].genUnmarsh(fmt.Sprintf("%s[%s]", acc, iv))
	n.g.WriteStringfn("}")
	return
}

type simpleNode struct {
	*nodeBase
	name         string
	t            string
	packSchema   string
	unpackSchema string
}

func newSimpleNode(base *nodeBase, t reflect.Type) (*simpleNode, error) {
	nt, err := nameFromReflect(t.Kind())
	return &simpleNode{
		nodeBase:     base,
		t:            nt,
		packSchema:   mpFieldToPackSchema[t.Kind()],
		unpackSchema: mpFieldToUnpackSchema[t.Kind()],
	}, errorWrap(err, "generating field")
}

func (n *simpleNode) getType() string {
	return n.t
}

func (n *simpleNode) genMarsh(acc string) {
	n.g.WriteStringfn(n.packSchema, acc)
	return
}

func (n *simpleNode) genUnmarsh(acc string) {
	n.g.WriteStringfn(n.unpackSchema, acc)
	return
}

var mpFieldToPackSchema = map[reflect.Kind]string{
	reflect.Int8:   "res = ayproto.PackUint8( res, uint8(%s),  0)",
	reflect.Int16:  "res = ayproto.PackUint16(res, uint16(%s), 0)",
	reflect.Int32:  "res = ayproto.PackUint32(res, uint32(%s), 0)",
	reflect.Int64:  "res = ayproto.PackUint64(res, uint64(%s), 0)",
	reflect.Int:    "res = ayproto.PackUint32(res, uint32(%s), 0)",
	reflect.Uint8:  "res = ayproto.PackUint8( res, %s,         0)",
	reflect.Uint16: "res = ayproto.PackUint16(res, %s,         0)",
	reflect.Uint32: "res = ayproto.PackUint32(res, %s,         0)",
	reflect.Uint64: "res = ayproto.PackUint64(res, %s,         0)",
	reflect.Uint:   "res = ayproto.PackUint32(res, uint32(%s), 0)",
	reflect.String: "res = ayproto.PackString(res, %s,         0)",
}

func simpleUnpack(f string) string {
	return fmt.Sprintf(""+
		"err = %s(r, &(%%[1]s), 0)\n"+
		"if err != nil {return fmt.Errorf(\"cant unpack %%[1]s: %%%%w\", err)}", f)
}

func typedUnpack(f, convT, convF string) string {
	return fmt.Sprintf(""+
		"err = %[2]s(r, &%[3]sconv, 0)\n"+
		"if err != nil {return fmt.Errorf(\"cant unpack %%[1]s: %%%%w\", err)}\n"+
		"%%[1]s = %[3]s(%[3]sconv)", convF, f, convT)
}

var mpFieldToUnpackSchema = map[reflect.Kind]string{
	reflect.Int8:   typedUnpack("ayproto.UnpackUint8", "int8", "uint8"),
	reflect.Int16:  typedUnpack("ayproto.UnpackUint16", "int16", "uint16"),
	reflect.Int32:  typedUnpack("ayproto.UnpackUint32", "int32", "uint32"),
	reflect.Int64:  typedUnpack("ayproto.UnpackUint64", "int64", "uint64"),
	reflect.Int:    typedUnpack("ayproto.UnpackUint32", "int", "uint32"),
	reflect.Uint8:  simpleUnpack("ayproto.UnpackUint8"),
	reflect.Uint16: simpleUnpack("ayproto.UnpackUint16"),
	reflect.Uint32: simpleUnpack("ayproto.UnpackUint32"),
	reflect.Uint64: simpleUnpack("ayproto.UnpackUint64"),
	reflect.Uint:   typedUnpack("ayproto.UnpackUint32", "uint", "uint32"),
	reflect.String: simpleUnpack("ayproto.UnpackString"),
}
