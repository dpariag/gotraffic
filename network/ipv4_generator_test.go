package network

import (
	"net"
	"testing"
)

func CheckExpectedIP(t *testing.T, ip net.IP, expected string) {
	expectedIP := net.ParseIP(expected)
	if !ip.Equal(expectedIP) {
		t.Errorf("IP Generation failed. Expected %v. Got %v\n", expected, ip.String())
	}
}

func TestSingleSequential(t *testing.T) {
	gen := NewSequentialIPGenerator(net.ParseIP("10.0.0.0"))
	ip := gen.GenerateIP()
	CheckExpectedIP(t, ip, "10.0.0.1")
}

func TestMultipleSequential(t *testing.T) {
	gen := NewSequentialIPGenerator(net.ParseIP("1.2.3.4"))
	ip := gen.GenerateIP()
	ip = gen.GenerateIP()
	CheckExpectedIP(t, ip, "1.2.3.6")
}

func TestLastOctetBoundary(t *testing.T) {
	gen := NewSequentialIPGenerator(net.ParseIP("10.0.1.255"))
	ip := gen.GenerateIP()
	CheckExpectedIP(t, ip, "10.0.2.0")

}

func TestTwoOctetBoundary2(t *testing.T) {
	gen := NewSequentialIPGenerator(net.ParseIP("192.168.255.255"))
	ip := gen.GenerateIP()
	CheckExpectedIP(t, ip, "192.169.0.0")
}

func TestSliceGeneration(t *testing.T) {
	seed := net.ParseIP("1.255.255.255")
	gen := NewSequentialIPGenerator(seed)
	ips := gen.GenerateIPs(3)

	expected := []net.IP{net.ParseIP("2.0.0.0"), net.ParseIP("2.0.0.1"), net.ParseIP("2.0.0.2")}

	if len(ips) != len(expected) {
		t.Errorf("Incorrect number of IPs generated in slice")
	}

	for i := 0; i < len(expected); i++ {
		if !ips[i].Equal(expected[i]) {
			t.Errorf("Slice of IPs generation failed. Expected %v. Got %v\n",
				expected[i], ips[i])
		}
	}
}
