package util

import (
	"encoding/binary"
	"math"
	"net"
)

type WrapperConn struct {
	net.Conn
}

func (e *WrapperConn) ReadData() (data []byte, err error) {
	// if err := e.SetDeadline(time.Now().Add(time.Second)); err != nil {
	// 	return nil, err
	// }
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("%v", r)
	// 	}
	// 	if err := e.SetDeadline(time.Time{}); err != nil {
	// 		panic(err)
	// 	}
	// }()

	buf := make([]byte, math.MaxInt16)
	n, err := e.Read(buf)
	if n < 8 || err != nil {
		return nil, err
	}
	size := int(binary.LittleEndian.Uint64(buf[:8]))

	return buf[8 : size+8], nil
}

func (e *WrapperConn) WriteData(data []byte) (n int, err error) {
	// if err := e.SetDeadline(time.Now().Add(time.Second)); err != nil {
	// 	return -1, err
	// }
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("%v", r)
	// 	}
	// 	if err := e.SetDeadline(time.Time{}); err != nil {
	// 		panic(err)
	// 	}
	// }()

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(len(data)))
	bytes = append(bytes, data...)

	return e.Write(bytes)
}
