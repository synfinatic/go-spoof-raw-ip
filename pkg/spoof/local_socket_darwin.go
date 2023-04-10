//go:build darwin
// +build darwin

package spoof

import (
	"fmt"
	"net"
	"syscall"
)

func setupLocalSocket(fd uintptr, iface net.Interface) error {
	var err error
	// bind to our interface in a different way than *BSD
	if err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BOUND_IF, iface.Index); err != nil {
		return fmt.Errorf("setsockopt IP_BOUND_IF: %s", err.Error())
	}

	// we will fill out the source IP and other auto-calculated fields.
	// not necessary on Linux, but no harm in setting it again!
	if err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
		return fmt.Errorf("setsockopt IP_HDRINCL: %s", err.Error())
	}
	/*

		// no need to receive anything, LOL... Linux uses 0 instead of 1
		if err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, 1); err != nil {
			return fmt.Errorf("setsockopt SO_RCVBUF: %s", err.Error())
		}
	*/

	return nil
}
