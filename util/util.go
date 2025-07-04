package util

import (
	"fmt"
	"github.com/rabobank/hzmon/conf"
	"net"
	"os"
)

// GetIP Gets the IPv4 address of the localhost (skipping loopback address and interfaces that are down)
func GetIP() string {
	UnKownIp := "UNKNOWN IP"
	if CF_IP := os.Getenv("CF_INSTANCE_IP"); CF_IP != "" {
		return CF_IP
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return UnKownIp
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return UnKownIp
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			LogDebug(fmt.Sprintf("found local IP %s", ip.String()))
			return ip.String()
		}
	}
	return UnKownIp
}

func LogDebug(msg string) {
	if conf.Debug {
		fmt.Println(msg)
	}
}
