package tuntap

/*
#include <string.h>
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <stdlib.h>
#include <linux/if.h>
#include <linux/if_tun.h>

char *tun_ioctl(int fd, char *name) {
  struct ifreq req;
  memset(&req, 0, sizeof(req));
  req.ifr_flags = IFF_TUN;
  strncpy(req.ifr_name, name, IFNAMSIZ);
  if (ioctl(fd, TUNSETIFF, (void*)&req)) {
    return NULL;
  }

  char *ret = malloc(strlen(req.ifr_name)+1);
  strcpy(ret, req.ifr_name);
  return ret;
}
*/
import "C"

import (
	"os"
	"unsafe"
)

const FlagTruncated = C.TUN_PKT_STRIP

func createInterface(file *os.File, ifPattern string, kind DevKind) (string, error) {
	cIfPattern := C.CString(ifPattern)
	defer C.free(unsafe.Pointer(cIfPattern))
	cIfName, err := C.tun_ioctl(C.int(file.Fd()), cIfPattern)
	defer C.free(unsafe.Pointer(cIfName))
	if cIfName == nil {
		return "", err
	}
	return C.GoString(cIfName), nil
}
