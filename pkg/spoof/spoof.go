package spoof

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	// "github.com/davecgh/go-spew/spew"
)

type SpoofType int

const (
	SpoofDialIPConn SpoofType = iota
	SpoofListenIPConn
	SpoofRAW
)

type Spoof struct {
	stype     SpoofType
	network   string
	iface     *net.Interface
	ifaceIP   net.IP
	Raw       int
	IPConn    *net.IPConn
	sysConn   syscall.RawConn
	connected bool
}

// NewSpoof creates a Spoof handle for either an net.IPConn or RAW_SOCK
// network in the same format as https://pkg.go.dev/net#Dial
func NewSpoof(stype SpoofType, iface *net.Interface, network string) (*Spoof, error) {
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

	s := &Spoof{
		stype:   stype,
		network: network,
		iface:   iface,
		ifaceIP: ip,
	}
	fmt.Printf("interface %s IP = %s\n", iface.Name, ip.To4().String())

	return s, nil
}

// Open calls Open() on our underlying socket
func (s *Spoof) Open(srcIP, dstIP net.IP) error {
	switch s.stype {
	case SpoofDialIPConn:
		return s.dialIPConn(srcIP, dstIP)

	case SpoofListenIPConn:
		return s.listenIPConn(srcIP)

	case SpoofRAW:
		return s.openRaw()

	default:
		return errors.New(fmt.Sprintf("Unsupported socket type: %d", s.stype))
	}
}

func (s *Spoof) openRaw() error {
	var err error
	// proto := syscall.IPPROTO_IP
	proto := syscall.IPPROTO_RAW
	if s.Raw, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, proto); err != nil {
		return err
	}
	if err = setupLocalSocket(uintptr(s.Raw), *(s.iface)); err != nil {
		return err
	}
	return nil
}

func (s *Spoof) dialIPConn(srcIP, dstIP net.IP) error {
	var err error

	laddr := &net.IPAddr{
		IP: srcIP,
	}

	raddr := &net.IPAddr{
		IP: dstIP,
	}
	if s.IPConn, err = net.DialIP(s.network, laddr, raddr); err != nil {
		return err
	}
	return s.finishIPConnSetup()
}

func (s *Spoof) finishIPConnSetup() error {
	var err error
	s.sysConn, err = s.IPConn.SyscallConn()
	if err != nil {
		return err
	}

	var opErr error
	err = s.sysConn.Control(func(fd uintptr) {
		opErr = setupLocalSocket(fd, *(s.iface))
	})
	if err != nil {
		return err
	}
	if opErr != nil {
		return err
	}
	return nil
}

func (s *Spoof) listenIPConn(srcIP net.IP) error {
	var err error

	laddr := &net.IPAddr{
		IP: srcIP,
	}

	if s.IPConn, err = net.ListenIP(s.network, laddr); err != nil {
		return err
	}

	return s.finishIPConnSetup()
}

func (s *Spoof) Close() error {
	switch s.stype {
	case SpoofDialIPConn, SpoofListenIPConn:
		return s.IPConn.Close()
	case SpoofRAW:
		return syscall.Close(s.Raw)
	default:
		return errors.New(fmt.Sprintf("Unsupported socket type: %d", s.stype))
	}
}

func (s *Spoof) SendTo4(srcIP, dstIP net.IP, dstPort uint16, opts gopacket.SerializeOptions,
	ip *layers.IPv4, layers ...gopacket.SerializableLayer) (int, error) {
	buffer := gopacket.NewSerializeBuffer()

	// update IPv6 header
	ip.SrcIP = srcIP
	ip.DstIP = dstIP
	l := []gopacket.SerializableLayer{
		ip,
	}

	for _, layer := range layers {
		l = append(l, layer)
	}

	// fmt.Printf("headers: %s", spew.Sdump(l))
	gopacket.SerializeLayers(buffer, opts, l...)
	// fmt.Printf("buffer: %s", spew.Sdump(buffer))
	// fmt.Printf("bytes: %s", spew.Sdump(buffer.Bytes()))

	switch s.stype {
	case SpoofListenIPConn:
		return s.IPConn.WriteToIP(buffer.Bytes(), &net.IPAddr{IP: dstIP})

	case SpoofDialIPConn:
		return s.IPConn.Write(buffer.Bytes()) // Only works on a connected socket?

	case SpoofRAW:
		addr := syscall.SockaddrInet4{
			// Port: int(htons(syscall.IPPROTO_UDP)),
			Addr: ip4ToByteSlice(dstIP),
		}
		flags := syscall.MSG_DONTROUTE
		// return len(buffer.Bytes()), syscall.Sendto(s.Raw, buffer.Bytes(), flags, &addr)

		return syscall.SendmsgN(s.Raw, buffer.Bytes(), []byte{}, &addr, flags)

	default:
		return -1, errors.New(fmt.Sprintf("Unsupported socket type: %d", s.stype))
	}
	// return syscall.Write(s.fd, buffer.Bytes()) // needs connect() first
	// return ssip.IPConn.WriteToIP(buffer.Bytes(), &net.IPAddr{IP: dstIP})
	// return ssip.IPConn.Write(buffer.Bytes())
}

/*
// SendTo6 sends an IPv6 packet using the specified srcIP and dstIP.
// opts specifies the options passed to gopacket.SerializeBuffer for
// the ip header and other layers.
func (s *Spoof) SendTo6(srcIP, dstIP net.IP, dstPort int, opts gopacket.SerializeOptions,
	ip *layers.IPv6, layers ...gopacket.SerializableLayer) (int, error) {
	buffer := gopacket.NewSerializeBuffer()

	// update IPv6 header
	ip.SrcIP = srcIP
	ip.DstIP = dstIP

	l := []gopacket.SerializableLayer{
		ip,
	}
	for _, layer := range layers {
		l = append(l, layer)
	}
	gopacket.SerializeLayers(buffer, opts, l...)

	switch s.stype {
	case SpoofListenIPConn:
		return s.IPConn.Write(buffer.Bytes()) // , &net.IPAddr{IP: dstIP})

	case SpoofDialIPConn:
		return s.IPConn.WriteToIP(buffer.Bytes(), &net.IPAddr{IP: dstIP})

	case SpoofRAW:
		addr := syscall.SockaddrInet6{
			Addr: ip6ToByteSlice(dstIP),
		}
		return len(buffer.Bytes()), syscall.Sendto(s.Raw, buffer.Bytes(), 0, &addr)

	default:
		return -1, errors.New(fmt.Sprintf("Unsupported socket type: %d", s.stype))
	}
}
*/
