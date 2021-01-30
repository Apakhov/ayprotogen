package ayproto

import (
	"bytes"
	"context"
	"io"
	"net"
)

type Server struct {
	h Handler
}

type Handler interface {
	ServeAYProto(ctx context.Context, c Conn, p Packet)
}

func ListenAndServe(ctx context.Context, network, addr string, h Handler) error {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	s := Server{
		h: h,
	}
	return s.Serve(ctx, ln)
}

func readHeader(c net.Conn) (p Packet, err error) {
	header := make([]byte, 12)
	_, err = io.ReadFull(c, header)
	if err != nil {
		return
	}
	rd := bytes.NewReader(header)
	var h Header
	UnpackUint32(rd, &h.Msg, 0)
	UnpackUint32(rd, &h.Len, 0)
	UnpackUint32(rd, &h.Sync, 0)
	data := make([]byte, h.Len)
	_, err = io.ReadFull(c, data)
	if err != nil {
		return
	}
	return Packet{
		Header: h,
		Data:   data,
	}, nil
}

func (s Server) Serve(ctx context.Context, ln net.Listener) (err error) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			defer c.Close()

			p, err := readHeader(c)
			if err != nil {
				return
			}
			s.h.ServeAYProto(context.Background(), Conn{c}, p)
		}(conn)
	}
}
