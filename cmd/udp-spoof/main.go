package main

import (
	"fmt"
	"net"
	"os"

	"github.com/alecthomas/kong"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/sirupsen/logrus"
	"github.com/synfinatic/go-spoof-raw-ip/pkg/spoof"
)

type CLI struct {
	Interface string `kong:"short='i',help='Interface to send packets out of',required"`
	SrcIP     string `kong:"short='s',help='Source IP',default='10.5.5.5'"`
	DstIP     string `kong:"short='d',help='Destination IP',default='172.16.1.100'"`
	SrcPort   uint16 `kong:"short='S',help='UDP source port',default='5555'"`
	DstPort   uint16 `kong:"short='D',help='UDP destination port',default='6666'"`
	Payload   string `kong:"short='p',help='UDP payload',default='this is my packet data'"`
	Type      string `kong:"short='t',help='Socket type [raw|dial|listen]',enum='raw,dial,listen',default='raw'"`
}

var log *logrus.Logger

func main() {
	var err error
	log = logrus.New()
	cli := CLI{}
	parser := kong.Must(
		&cli,
		kong.Name("udp-spoof"),
		kong.Description("Sample program showing how to use spoofrawip"),
		kong.UsageOnError(),
		kong.Vars{},
	)
	_, err = parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	iface, err := net.InterfaceByName(cli.Interface)
	if err != nil {
		log.WithError(err).Fatalf("Unable to lookup %s", cli.Interface)
	}

	srcIP := net.ParseIP(cli.SrcIP)
	dstIP := net.ParseIP(cli.DstIP)
	fmt.Printf("SrcIP: %v\t\tDstIP: %v\n", srcIP, dstIP)

	var s *spoof.Spoof
	switch cli.Type {
	case "raw":
		if s, err = spoof.NewSpoof(spoof.SpoofRAW, iface, "ip4:udp"); err != nil {
			log.WithError(err).Fatalf("Unable to init NewSpoofSrcIP")
		}
	case "listen":
		if s, err = spoof.NewSpoof(spoof.SpoofListenIPConn, iface, "ip4:udp"); err != nil {
			log.WithError(err).Fatalf("Unable to init NewSpoofSrcIP")
		}
	case "dial":
		if s, err = spoof.NewSpoof(spoof.SpoofDialIPConn, iface, "ip4:udp"); err != nil {
			log.WithError(err).Fatalf("Unable to init NewSpoofSrcIP")
		}
	}

	if err = s.Open(srcIP, dstIP); err != nil {
		log.WithError(err).Fatalf("Unable to open socket")
	}

	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	payload := gopacket.Payload([]byte(cli.Payload))
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(cli.SrcPort),
		DstPort: layers.UDPPort(cli.DstPort),
	}
	ip4 := &layers.IPv4{
		Version:    4,
		IHL:        5,
		TOS:        0,
		FragOffset: 0,
		SrcIP:      srcIP,
		DstIP:      dstIP,
		Flags:      layers.IPv4DontFragment,
		Protocol:   layers.IPProtocolUDP,
		Length:     0, // uint16(20 + len(cli.Payload)),
		Id:         0x1234,
		TTL:        16,
	}
	udp.SetNetworkLayerForChecksum(ip4)

	len, err := s.SendTo4(srcIP, dstIP, cli.DstPort, opts, ip4, udp, payload)
	if err != nil {
		log.WithError(err).Fatalf("Unable to send packet")
	}
	log.Infof("Sent %d bytes to %s:%d", len, cli.DstIP, cli.DstPort)
}
