// +build ignore

package tuntap

/*
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <linux/if.h>
#include <linux/if_tun.h>

#define IFREQ_SIZE sizeof(struct ifreq)
*/
import "C"

const (
	flagTruncated = C.TUN_PKT_STRIP

	iffTun = C.IFF_TUN
	iffTap = C.IFF_TAP
	iffOneQueue = C.IFF_ONE_QUEUE
)

type ifReq struct {
	Name [C.IFNAMSIZ]byte
	Flags uint16
	pad [C.IFREQ_SIZE-C.IFNAMSIZ-2]byte
}
