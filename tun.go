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
	// Receive/send IP frames.
	DevTun DevKind = iota
	// Receive/send Ethernet frames.
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
	DevName string
	In      <-chan *Packet
	Out     chan<- *Packet

	file     *os.File
	shutdown chan interface{}
}

func (t *TunTap) Close() error {
	close(t.shutdown)
	return t.file.Close()
}

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
