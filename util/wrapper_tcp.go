package util

import (
	"encoding/binary"
	"math"
	"net"
)

type WrapperConn struct {
	net.Conn
}

func (e *WrapperConn) ReadData() ([]byte, error) {
	buf := make([]byte, math.MaxInt16)
	n, err := e.Read(buf)
	if n < 8 || err != nil {
		return nil, err
	}
	size := int(binary.LittleEndian.Uint64(buf[:8]))

	return buf[8 : size+8], nil
}

func (e *WrapperConn) WriteData(data []byte) (int, error) {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(len(data)))
	bytes = append(bytes, data...)

	return e.Write(bytes)
}
