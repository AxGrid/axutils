package net

import (
	"encoding/binary"
	"github.com/go-errors/errors"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (14.03.2024)
*/

var ErrUnsupportedType = errors.New("unsupported type")

func GetBytesFromInt(data int, count int) []byte {
	bs := make([]byte, count)
	for i := 0; i < count; i++ {
		bs[i] = byte(data >> uint(i*8))
	}
	return bs
}

func GetIntFromBytes(data []byte) int {
	var res int
	for i, b := range data {
		res |= int(b) << uint(i*8)
	}
	return res
}

func PackUint(order binary.ByteOrder, v interface{}) (data []byte) {
	switch val := v.(type) {
	case uint64:
		data = make([]byte, 8)
		order.PutUint64(data, val)
	case uint32:
		data = make([]byte, 4)
		order.PutUint32(data, val)
	case uint16:
		data = make([]byte, 2)
		order.PutUint16(data, val)
	case uint8:
		data = []byte{byte(val)}
	default:
		panic("unsupported type")
	}
	return data
}

func UnpackUint(order binary.ByteOrder, data []byte, v interface{}) error {
	switch val := v.(type) {
	case *uint64:
		*val = order.Uint64(data)
	case *uint32:
		*val = order.Uint32(data)
	case *uint16:
		*val = order.Uint16(data)
	case *uint8:
		*val = data[0]
	default:
		return ErrUnsupportedType
	}
	return nil
}

func GetBytesFromUInt32(len uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, len)
	return bs
}

func GetUInt32FromBytes(lens []byte) uint32 {
	return binary.LittleEndian.Uint32(lens)
}

func GetBytesFromUInt16(len uint16) []byte {
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, len)
	return bs
}

func GetUInt16FromBytes(lens []byte) uint16 {
	return binary.LittleEndian.Uint16(lens)
}

// AddSize32 - add size(4 byte) to data
func AddSize32(data []byte) []byte {
	ld := uint32(len(data))
	return append(GetBytesFromUInt32(ld), data...)
}

// AddSize16 - add size(2 byte) to data
func AddSize16(data []byte) []byte {
	ld := uint16(len(data))
	return append(GetBytesFromUInt16(ld), data...)
}

// Chunk - split slice to chunks
func Chunk(data []byte, size int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

func ByteToBoolArray(data []byte, count ...int) []bool {
	var res []bool
	for _, b := range data {
		for i := 0; i < 8; i++ {
			res = append(res, b&(1<<uint(i)) != 0)
		}
	}
	if len(count) > 0 {
		return res[:count[0]]
	}
	return res
}

func BoolArrayToByte(data []bool) []byte {
	var res []byte
	for i := 0; i < len(data); i += 8 {
		var b byte
		for j := 0; j < 8; j++ {
			if i+j >= len(data) {
				break
			}
			if data[i+j] {
				b |= 1 << uint(j)
			}
		}
		res = append(res, b)
	}
	return res
}

func CreateAllTrueBoolArray(size int) []bool {
	arr := make([]bool, size)
	for i := 0; i < size; i++ {
		arr[i] = true
	}
	return arr
}

func CopyBytes(data []byte, l int) []byte {
	res := make([]byte, l)
	copy(res, data[:l])
	return res
}
