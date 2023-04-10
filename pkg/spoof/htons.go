package spoof

import (
	"encoding/binary"
	"net"
	"unsafe"
)

func htons(i uint16) uint16 {
	var nativeEndian binary.ByteOrder
	tbuf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&tbuf[0])) = uint16(0xABCD)

	switch tbuf {
	case [2]byte{0xCD, 0xAB}:
		nativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		nativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}

	buf := make([]byte, 4)
	nativeEndian.PutUint16(buf, i)
	return binary.BigEndian.Uint16(buf)
}

func ip4ToByteSlice(i net.IP) [4]byte {
	ip := i.To4()
	return [4]byte{ip[0], ip[1], ip[2], ip[3]}
}

func ip6ToByteSlice(i net.IP) [16]byte {
	ip := i.To16()
	return [16]byte{ip[0], ip[1], ip[2], ip[3], ip[4], ip[5],
		ip[6], ip[7], ip[8], ip[9], ip[10], ip[11],
		ip[12], ip[13], ip[14], ip[15],
	}
}
