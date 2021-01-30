package ayproto

import (
	"context"
	"net"
)

type Conn struct {
	net.Conn
}

func (c *Conn) Send(ctx context.Context, p Packet) error {
	buff := PackUint32([]byte{}, p.Header.Msg, 0)
	buff = PackUint32(buff, p.Header.Len, 0)
	buff = PackUint32(buff, p.Header.Sync, 0)
	buff = append(buff, p.Data...)
	_, err := c.Conn.Write(buff)
	return err
}
