package webserver

import (
	"net"
	"strings"
)

type CIDRs struct {
	elements []*net.IPNet
}

func newCIDRs(addresses []string) (*CIDRs, error) {
	cidr := make([]*net.IPNet, 0, len(addresses))
	for _, addr := range addresses {
		if !strings.Contains(addr, "/") {
			ip := net.ParseIP(addr)
			if ip == nil {
				return nil, &net.ParseError{Type: "IP address", Text: addr}
			}
			if ip.To4() == nil {
				addr += "/128"
			} else {
				addr += "/32"
			}
		}
		_, cidrNet, err := net.ParseCIDR(addr)
		if err != nil {
			return nil, err
		}
		cidr = append(cidr, cidrNet)
	}
	return &CIDRs{elements: cidr}, nil
}

func (c *CIDRs) Len() int {
	return len(c.elements)
}

func (c *CIDRs) ContainIP(ip net.IP) bool {
	for _, el := range c.elements {
		if el.Contains(ip) {
			return true
		}
	}

	return false
}

func (c *CIDRs) ContainStringIP(ip string) bool {
	return c.ContainIP(net.ParseIP(ip))
}
