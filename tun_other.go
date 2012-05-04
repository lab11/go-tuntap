// +build !linux

package tuntap

import (
	"os"
)

flagTruncated = 0

createInterface(f *os.File, ifPattern string, kind DevKind) (string, error) {
	panic("Not implemented on this platform")
}
