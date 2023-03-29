package spoofsrcip

import (
	"fmt"
	"net"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type SpoofSrcIP struct {
	IPConn  *net.IPConn
	iface   *net.Interface
	ifaceIP net.IP
	network string
}

// NewSpoofSrcIP opens a socket on the specified iface for the given
// network in the same format as https://pkg.go.dev/net#Dial
func NewSpoofSrcIP(iface *net.Interface, network string) (*SpoofSrcIP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err

	}
	var ip net.IP
	for _, addr := range addrs {
		ips := strings.Split(addr.String(), "/")
		ip = net.ParseIP(ips[0])
		if ip.To4() != nil {
			break
		}
	}

	ssip := &SpoofSrcIP{
		network: network,
		iface:   iface,
		ifaceIP: ip,
	}

	return ssip, nil
}

func (ssip *SpoofSrcIP) Open() error {
	var err error
	ssip.IPConn, err = net.ListenIP(ssip.network, nil)
	if err != nil {
		return err
	}

	/*
		sysConn, err := ssip.IPConn.SyscallConn()
		if err != nil {
			return err
		}

			var opErr error
			err = sysConn.Control(func(fd uintptr) {
				opErr = setupLocalSocket(fd, *(ssip.iface))
			})
			if err != nil {
				return err
			}
			if opErr != nil {
				return err
			}
	*/
	return nil
}

func (ssip *SpoofSrcIP) Close() {
	ssip.IPConn.Close()
}

// SendTo4 sends an IPv4 packet using the specified srcIP and dstIP.
// opts specifies the options passed to gopacket.SerializeBuffer for
// the ip header and other layers.
func (ssip *SpoofSrcIP) SendTo4(srcIP, dstIP net.IP, dstPort uint16, opts gopacket.SerializeOptions,
	ip *layers.IPv4, layers ...gopacket.SerializableLayer) (int, error) {
	buffer := gopacket.NewSerializeBuffer()
	ip.SrcIP = srcIP
	ip.DstIP = dstIP
	l := []gopacket.SerializableLayer{ip}
	for _, layer := range layers {
		l = append(l, layer)
	}
	fmt.Printf("headers: %s", spew.Sdump(l))
	gopacket.SerializeLayers(buffer, opts, l...)
	fmt.Printf("buffer: %s", spew.Sdump(buffer))
	fmt.Printf("bytes: %s", spew.Sdump(buffer.Bytes()))

	return ssip.IPConn.WriteToIP(buffer.Bytes(), &net.IPAddr{IP: dstIP})
	//return ssip.IPConn.Write(buffer.Bytes())
}

// SendTo6 sends an IPv6 packet using the specified srcIP and dstIP.
// opts specifies the options passed to gopacket.SerializeBuffer for
// the ip header and other layers.
func (ssip *SpoofSrcIP) SendTo6(srcIP, dstIP net.IP, dstPort int, opts gopacket.SerializeOptions,
	ip *layers.IPv6, layers ...gopacket.SerializableLayer) (int, error) {
	buffer := gopacket.NewSerializeBuffer()
	ip.SrcIP = srcIP
	ip.DstIP = dstIP

	fmt.Printf("header: %s", spew.Sdump(ip))

	l := []gopacket.SerializableLayer{ip}
	for _, layer := range layers {
		l = append(l, layer)
	}
	gopacket.SerializeLayers(buffer, opts, l...)

	return ssip.IPConn.WriteToIP(buffer.Bytes(), &net.IPAddr{IP: dstIP})
	// return ssip.IPConn.Write(buffer.Bytes())
}
