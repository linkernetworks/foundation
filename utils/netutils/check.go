package netutils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func CheckNetworkConnectivity(ip string, port int, protocol string, timeout int) error {
	host := net.JoinHostPort(ip, strconv.Itoa(port))

	find := false
	for i := 0; i < timeout; i++ {
		conn, err := net.DialTimeout(strings.ToLower(protocol), host, time.Duration(1)*time.Second)
		if err == nil {
			find = true
			defer conn.Close()
			break
		}
		time.Sleep(time.Duration(1) * time.Second)
	}

	if !find {
		return fmt.Errorf("Connect to %s@%s:%d fails", protocol, ip, port)
	}
	return nil
}
