# spoofrawip

A pure Go Source IP Spoofing library

### Status

This project is a work in progress.  Not yet working :(

### Overview

It's pretty trival to use [gopacket/pcap](https://pkg.go.dev/github.com/google/gopacket/pcap)
to inject packets with arbitrary IP header values, but this has two major downsides:

 1. Now you are using CGO to link libpcap
 1. Packets can not be seen by the local host

This library allows you to easily spoof the source IP address for arbitrary
packets without using libpcap.

### Reading

Linux and *BSD/macOS behave differently which causes some issues with spoofrawip:

 1. Linux allows reading from RAW sockes for TCP and UDP packets
 2. *BSD/macOS never passes TCP or UDP packets to a RAW socket

[More details here](https://sock-raw.org/papers/sock_raw).