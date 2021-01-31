package packgen

import (
	"context"
	"reflect"
)

func (g *Generator) AddStruct(t reflect.Type) {
	if t.Kind() != reflect.Struct {
		g.addErr(StructErr{
			Err:  ErrNotStruct,
			Name: t.Name(),
		})
		return
	}

	root := &rootNode{newNodeBase(g)}

	err := makeChildNode(root, t, opts{})
	if err != nil {
		g.addErr(StructErr{
			Err:  err,
			Name: t.Name(),
		})
		return
	}
	g.structs[root.childs[0].getType()] = root
}

func (g *Generator) AddServer(t reflect.Type, ftDescr ...interface{}) {
	for i := 0; i < len(ftDescr); i += 2 {
		ft := ftDescr[i].(reflect.Type)
		g.AddMethod(ft)
	}

	if t.Kind() != reflect.Struct {
		g.addErr(StructErr{
			Err:  ErrNotStruct,
			Name: t.Name(),
		})
		return
	}

	g.servers = append(g.servers, newServer(newNodeBase(g), t, ftDescr...))
}

func checkMethodSiganure(t reflect.Type) bool {
	if t.NumIn() != 2 || // context, in msg
		!t.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) ||
		t.In(1).Kind() != reflect.Struct ||
		t.NumOut() < 1 { // out msgs
		return false
	}

	for i := 0; i < t.NumOut(); i++ {
		if t.Out(i).Kind() != reflect.Ptr || t.Out(i).Elem().Kind() != reflect.Struct {
			return false
		}
	}

	return true
}

func (g *Generator) AddMethod(t reflect.Type) {
	if t.Kind() != reflect.Func {
		g.addErr(StructErr{
			Err:  ErrNotFunc,
			Name: t.Name(),
		})
		return
	}

	if !checkMethodSiganure(t) {
		g.addErr(StructErr{
			Err:  ErrFuncSiganure,
			Name: t.Name(),
		})
		return
	}

	g.AddStruct(t.In(1))
	for i := 0; i < t.NumOut(); i++ {
		g.AddStruct(t.Out(i).Elem())
	}
}
