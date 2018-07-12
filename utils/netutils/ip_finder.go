package netutils

import (
	"log"
	"net"
)

func MustFindInterface(interfaceName string) net.Interface {
	ifaces, err := net.Interfaces()

	if err != nil {
		panic(err)
	}

	for _, i := range ifaces {
		if i.Name == interfaceName {
			return i
		}
	}

	panic("interface not found.")
}

func MustFindInterfaceGlobalUnicastIp(interfaceName string) net.IP {
	var ip net.IP

	i := MustFindInterface(interfaceName)

	addrs, err := i.Addrs()
	if err != nil {
		log.Fatal(err)
		return ip
	}

	// handle err
	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			if v.IP.IsGlobalUnicast() {
				return v.IP
			}
		case *net.IPAddr:
			if v.IP.IsGlobalUnicast() {
				return v.IP
			}
		}
	}
	return ip
}
