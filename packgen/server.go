package packgen

import (
	"reflect"
)

type MethodDoc struct {
	Name     string
	SVC      int32
	Wrappers []string
}

type server struct {
	g      *Generator
	s      *structNode
	ms     []reflect.Type
	descrs []MethodDoc
}

func newServer(base *nodeBase, t reflect.Type, ftDescr ...interface{}) *server {
	serv := &server{
		g: base.g,
		s: newStructNode(base, t),
	}

	for i := 0; i < len(ftDescr); i += 2 {
		ft, descr := ftDescr[i].(reflect.Type), ftDescr[i+1].(MethodDoc)
		serv.ms = append(serv.ms, ft)
		serv.descrs = append(serv.descrs, descr)
	}

	return serv
}

func (s *server) genHandle(t reflect.Type, doc MethodDoc) {
	s.g.WriteStringfn("case %d:", doc.SVC)
	s.g.WriteStringfn("r := &%s{}", t.In(1).Name())
	s.g.WriteStringfn("err := r.UnmarshalAyproto(bytes.NewReader(p.Data))")
	s.g.WriteStringfn("fmt.Println(err)")
	s.g.WriteStringfn("if err != nil {")
	s.g.WriteStringfn("resBt = ayproto.GenericErrorResp")
	s.g.WriteStringfn("break")
	s.g.WriteStringfn("}")
	resN := t.NumOut()
	for i := 1; i <= resN; i++ {
		s.g.WriteStringf("r%d", i)
		if i != resN {
			s.g.WriteStringf(", ")
		}
	}
	s.g.WriteStringfn(" := s.%s(ctx, *r)", doc.Name)
	for i := resN; i > 0; i-- {
		s.g.WriteStringfn("if r%d != nil {", i)
		s.g.WriteStringfn("resBt, _ = r%d.MarshalAyproto()", i)
		s.g.WriteStringf("}")
		if i != 1 {
			s.g.WriteStringf(" else ")
		} else {
			s.g.WriteStringfn(" else {")
			s.g.WriteStringfn("resBt = ayproto.GenericErrorResp")
			s.g.WriteStringfn("}")
		}
	}
}

func (s *server) gen() {
	s.g.WriteStringfn("func (s *%s) ServeAYProto(ctx context.Context, c ayproto.Conn, p ayproto.Packet) {", s.s.getType())
	s.g.WriteStringfn("var resBt []byte")
	s.g.WriteStringfn("switch p.Header.Msg {")
	for i := 0; i < len(s.descrs); i++ {
		s.genHandle(s.ms[i], s.descrs[i])
	}
	s.g.WriteStringfn("default:")
	s.g.WriteStringfn("resBt = ayproto.GenericErrorResp")
	s.g.WriteStringfn("}")
	s.g.WriteStringfn("err := c.Send(ctx, ayproto.ResponseTo(p, resBt))")
	s.g.WriteStringfn("if err != nil {")
	s.g.WriteStringfn("fmt.Println(err)")
	s.g.WriteStringfn("}")
	s.g.WriteStringfn("}")
}
