// Simple use of the tuntap package that prints packets received by the interface.
package main

import (
	"fmt"
	"os"

	"code.google.com/p/tuntap"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("syntax:", os.Args[0], "tun|tap", "<device name>")
		return
	}

	var typ tuntap.DevKind
	switch os.Args[1] {
	case "tun":
		typ = tuntap.DevTun
	case "tap":
		typ = tuntap.DevTap
	default:
		fmt.Println("Unknown device type", os.Args[1])
		return
	}

	tun, err := tuntap.Open(os.Args[2], typ)
	if err != nil {
		fmt.Println("Error opening tun/tap device:", err)
		return
	}

	fmt.Println("Listening on", tun.Name())
	for {
		pkt, err := tun.ReadPacket()
		if err != nil {
			fmt.Println("Read error:", err)
		} else {
			if pkt.Truncated {
				fmt.Printf("!")
			} else {
				fmt.Printf(" ")
			}
			fmt.Printf("%x %x\n", pkt.Protocol, pkt.Packet)
		}
	}
}
