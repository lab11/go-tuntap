Tun / Tap Bindings for Go
=========================


Install
-------

    go get github.com/lab11/go-tuntap/tuntap


Usage
-----

    import "github.com/lab11/go-tuntap/tuntap"

    func main () {
        var tund *tuntap.Interface
        var erro error
        var inpkt *tuntap.Packet

        tund, err = tuntap.Open("tun0", tuntap.DevTun)
        inpkt = tund.ReadPacket()
    }


Thanks
------

Thank you to @danderson for writing the original version at
http://code.google.com/p/tuntap/ .
