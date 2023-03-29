package main

import (
	"fmt"
	"net"
	"os"

	"github.com/alecthomas/kong"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/sirupsen/logrus"
	"github.com/synfinatic/go-spoof-raw-ip/spoofsrcip"
)

type CLI struct {
	Interface string `kong:"short='i',help='Interface to send packets out of'"`
	SrcIP     string `kong:"short='s',help='Source IP'"`
	DstIP     string `kong:"short='d',help='Destination IP'"`
	SrcPort   uint16 `kong:"short='S',help='UDP source port'"`
	DstPort   uint16 `kong:"short='D',help='UDP destination port'"`
	Payload   string `kong:"short='p',help='UDP payload'"`
}

var log *logrus.Logger

func main() {
	log = logrus.New()
	cli := CLI{}
	parser := kong.Must(
		&cli,
		kong.Name("udp-spoof"),
		kong.Description("Sample program showing how to use spoofrawip"),
		kong.UsageOnError(),
		kong.Vars{},
	)
	_, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	iface, err := net.InterfaceByName(cli.Interface)
	if err != nil {
		log.WithError(err).Fatalf("Unable to lookup %s", cli.Interface)
	}

	srip, err := spoofsrcip.NewSpoofSrcIP(iface, "ip4:17") // hard code ipv4/UDP
	if err != nil {
		log.WithError(err).Fatalf("Unable to init NewSpoofSrcIP")
	}

	srcIP := net.ParseIP(cli.SrcIP)
	dstIP := net.ParseIP(cli.DstIP)
	fmt.Printf("SrcIP: %v\t\tDstIP: %v\n", srcIP, dstIP)

	if err := srip.Open(); err != nil {
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
		Protocol:   layers.IPProtocolUDP,
		Length:     0,
		Id:         0x1234,
		TTL:        2,
	}
	udp.SetNetworkLayerForChecksum(ip4)

	len, err := srip.SendTo4(srcIP, dstIP, cli.DstPort, opts, ip4, udp, payload)
	if err != nil {
		log.WithError(err).Fatalf("Unable to send packet")
	}
	log.Infof("Sent %d bytes to %s:%d", len, cli.DstIP, cli.DstPort)
}
