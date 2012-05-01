// Package tuntap provides a portable interface to create and use
// TUN/TAP virtual network interfaces.
//
// Note that while this package lets you create the interface and pass
// packets to/from it, it does not provide an API to configure the
// interface. Interface configuration is a very large topic and should
// be dealt with separately.
package tuntap

import (
	"encoding/binary"
	"io"
	"os"
	"unsafe"
)

type DevKind int

const (
	// Receive/send layer 3 packets (IP, IPv6, OSPF...)
	DevTun DevKind = iota
	// Receive/send Ethernet II frames.
	DevTap
)

type Packet struct {
	// The Ethernet type of the packet. Commonly seen values are
	// 0x8000 for IPv4 and 0x86dd for IPv6.
	Protocol int
	// True if the packet was too large to be read completely.
	Truncated bool
	// The raw bytes of the Ethernet payload (for DevTun) or the full
	// Ethernet frame (for DevTap).
	Packet []byte
}

type TunTap struct {
	// The name of the interface. May be different from the name given
	// to Open(), if the latter was a pattern.
	DevName string
	// Channel of packets coming from the kernel.
	In      <-chan *Packet
	// Channel of packets going to the kernel.
	Out     chan<- *Packet

	file     *os.File
	shutdown chan interface{}
}

// Disconnect from the tun/tap interface.
//
// If the interface isn't configured to be persistent, it is
// immediately destroyed by the kernel.
func (t *TunTap) Close() error {
	close(t.shutdown)
	return t.file.Close()
}

// Open connects to the specified tun/tap interface.
//
// If the specified device has been configured as persistent, this
// simply looks like a "cable connected" event to observers of the
// interface. Otherwise, the interface is created out of thin air.
//
// ifPattern can be an exact interface name, e.g. "tun42", or a
// pattern containing one %d format specifier, e.g. "tun%d". In the
// latter case, the kernel will select an available interface name and
// create it.
//
// Returns a TunTap object with channels to send/receive packets, or
// nil and an error if connecting to the interface failed.
func Open(ifPattern string, kind DevKind) (*TunTap, error) {
	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	ifName, err := createInterface(file, ifPattern, kind)
	if err != nil {
		file.Close()
		return nil, err
	}

	in := make(chan *Packet)
	out := make(chan *Packet)
	shutdown := make(chan interface{})
	go reader(file, in, shutdown)
	go writer(file, out, shutdown)
	ret := &TunTap{
		DevName:  ifName,
		In:       in,
		Out:      out,
		file:     file,
		shutdown: shutdown,
	}
	return ret, nil
}

func reader(file io.Reader, ch chan *Packet, shutdown chan interface{}) {
	// Enough to read a jumbo ethernet frame. If that's not enough,
	// truncated packets will show up.
	buf := make([]byte, 10000)
	for {
		n, err := file.Read(buf)
		if err != nil {
			return
		}
		pkt := &Packet{Packet: buf[4:n]}
		pkt.Protocol = int(binary.BigEndian.Uint16(buf[2:4]))
		flags := *(*uint16)(unsafe.Pointer(&buf[0]))
		if flags&flagTruncated != 0 {
			pkt.Truncated = true
		}
		select {
		case ch <- pkt:
		case <-shutdown:
			return
		}
	}
}

func writer(file io.Writer, ch chan *Packet, shutdown chan interface{}) {
	for {
		select {
		case <-shutdown:
			return
		case pkt := <-ch:
			buf := make([]byte, len(pkt.Packet)+4)
			binary.BigEndian.PutUint16(buf[2:4], uint16(pkt.Protocol))
			copy(buf[4:], pkt.Packet)
			if n, err := file.Write(buf); err != nil || n != len(buf) {
				// tuntap should not be giving us short writes, since
				// we need to atomically pass one packet at a time.
				return
			}
		}
	}
}
