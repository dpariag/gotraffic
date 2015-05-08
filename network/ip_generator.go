package network

import (
	"net"
)

type IPGenerator interface {
	GenerateIP() net.IP
	GenerateIPs(count uint64) []net.IP
}

type sequential struct {
	seed net.IP
}

func (s sequential) GenerateIP() net.IP {
	for i := len(s.seed) - 1; i >= 0; i-- {
		s.seed[i]++
		if s.seed[i] != 0 {
			break
		}
	}
	newIP := make(net.IP, len(s.seed))
	copy(newIP, s.seed)
	return newIP
}

func (s sequential) GenerateIPs(count uint64) []net.IP {
	ips := make([]net.IP, count)
	for i := uint64(0); i < count; i++ {
		ips[i] = s.GenerateIP()
	}
	return ips
}

func NewSequentialIPGenerator(seed net.IP) IPGenerator {
	return sequential{seed: seed.To16()}
}
