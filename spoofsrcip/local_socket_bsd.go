//go:build freebsd || dragonfly || netbsd || openbsd
// +build freebsd dragonfly netbsd openbsd

package spoofsrcip

import (
	"fmt"
	"net"
	"syscall"
)

func setupLocalSocket(fd uintptr, iface net.Interface) error {
	var err error

	// we will fill out the source IP and other auto-calculated fields
	if err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
		return fmt.Errorf("setsockopt IP_HDRINCL: %s", err.Error())
	}

	// allow bind(2) on this socket, but not actually bind
	if err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BINDANY, 1); err != nil {
		return fmt.Errorf("setsockopt IP_BINDANY: %s", err.Error())
	}

	var addr net.IP
	addrs, err := iface.Addrs()
	if err != nil {
		return fmt.Errorf("getting Addrs(): %s", err.Error())
	}
	for _, a := range addrs {
		ip := a.String()
		addr = net.ParseIP(a.String())
	}

	sa = syscall.SockaddrInet4{
		Port: 0,
		Addr: addr.To4(),
	}
	if err = syscall.Bind(int(fd), sa); err != nil {
		return fmt.Errorf("bind: %s", err.Error())
	}

	// no need to receive anything.
	if err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, 1); err != nil {
		return fmt.Errorf("setsockopt SO_RCVBUF: %s", err.Error())
	}

	return nil
}
