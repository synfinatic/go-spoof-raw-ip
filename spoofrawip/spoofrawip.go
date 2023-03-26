package spoofsrcip

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type SpoofSrcIP struct {
	SrcIP net.IP
	conn  *net.IPConn
	iface net.Interface
}

type Layer interface {
	SerializeTo(gopacket.SerializeBuffer, gopacket.SerializeOptions) error
}

func NewSpoofSrcIP(iface net.Interface, proto string) (*SpoofSrcIP, error) {
	conn, err := net.DialIP(proto, nil, nil)
	if err != nil {
		return nil, err
	}

	sysConn, err := conn.SyscallConn()
	if err != nil {
		return nil, err
	}

	var opErr error
	err = sysConn.Control(func(fd uintptr) {
		opErr = setupLocalSocket(fd, iface)
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, err
	}

	ssip := &SpoofSrcIP{
		conn:  conn,
		iface: iface,
	}

	return ssip, nil
}

func (ssip *SpoofSrcIP) buildTop(layer4, payload Layer, checksum bool) (gopacket.SerializeBuffer, error) {
	opts := gopacket.SerializeOptions{
		FixLengths:       false,
		ComputeChecksums: checksum,
	}
	buffer := gopacket.NewSerializeBuffer()
	// payload
	if err := payload.SerializeTo(buffer, opts); err != nil {
		return buffer, fmt.Errorf("can't serialize payload: %s", err.Error())
	}

	// layer4
	if err := layer4.SerializeTo(buffer, opts); err != nil {
		return buffer, fmt.Errorf("can't serialize layer4: %s", err.Error())
	}
	return buffer, nil
}

func (ssip *SpoofSrcIP) SendTo4(srcIP, dstIP net.IP, dstPort int, ip layers.IPv4, layer4, payload Layer, checksum bool) (int, error) {
	buffer, err := ssip.buildTop(layer4, payload, checksum)
	if err != nil {
		return 0, err
	}

	opts := gopacket.SerializeOptions{
		FixLengths:       false,
		ComputeChecksums: checksum,
	}

	// ip header
	ip.SrcIP = srcIP
	ip.DstIP = dstIP
	if err := ip.SerializeTo(buffer, opts); err != nil {
		return 0, fmt.Errorf("can't serialize ip header: %s", err.Error())
	}

	return ssip.conn.WriteToIP(buffer.Bytes(), &net.IPAddr{IP: dstIP})
}

func (ssip *SpoofSrcIP) SendTo6(srcIP, dstIP net.IP, dstPort int, ip layers.IPv6, layer4, payload Layer, checksum bool) (int, error) {
	buffer, err := ssip.buildTop(layer4, payload, checksum)
	if err != nil {
		return 0, err
	}

	opts := gopacket.SerializeOptions{
		FixLengths:       false,
		ComputeChecksums: checksum,
	}

	// ip header
	ip.SrcIP = srcIP
	ip.DstIP = dstIP
	if err := ip.SerializeTo(buffer, opts); err != nil {
		return 0, fmt.Errorf("can't serialize ip header: %s", err.Error())
	}

	return ssip.conn.WriteToIP(buffer.Bytes(), &net.IPAddr{IP: dstIP})
}
