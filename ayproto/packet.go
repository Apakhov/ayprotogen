package ayproto

type Header struct {
	Msg  uint32
	Len  uint32
	Sync uint32
}

type Packet struct {
	Header Header
	Data   []byte
}

func ResponseTo(p Packet, bt []byte) Packet {
	return Packet{
		Header{
			Msg:  p.Header.Msg,
			Len:  uint32(len(bt)),
			Sync: p.Header.Sync,
		},
		bt,
	}
}
