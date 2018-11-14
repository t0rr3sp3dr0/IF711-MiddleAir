package util

import (
	"encoding/binary"
	"math"
	"net"
)

type WrapperPacketConn struct {
	net.PacketConn
}

func (e *WrapperPacketConn) ReadData() ([]byte, net.Addr, error) {
	buf := make([]byte, math.MaxInt16)
	n, addr, err := e.ReadFrom(buf)
	if n < 8 || err != nil {
		return nil, nil, err
	}
	size := int(binary.LittleEndian.Uint64(buf[:8]))

	return buf[8 : size+8], addr, nil
}

func (e *WrapperPacketConn) WriteData(addr net.Addr, data []byte) (int, error) {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(len(data)))
	bytes = append(bytes, data...)

	return e.WriteTo(bytes, addr)
}
