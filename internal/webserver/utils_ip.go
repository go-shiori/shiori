package webserver

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

var (
	userRealIpHeaderCandidates = [...]string{"X-Real-Ip", "X-Forwarded-For"}
	// From: https://github.com/letsencrypt/boulder/blob/main/bdns/dns.go#L30-L146
	// Private CIDRs to ignore
	privateNetworks = []net.IPNet{
		// RFC1918
		// 10.0.0.0/8
		{
			IP:   []byte{10, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		// 172.16.0.0/12
		{
			IP:   []byte{172, 16, 0, 0},
			Mask: []byte{255, 240, 0, 0},
		},
		// 192.168.0.0/16
		{
			IP:   []byte{192, 168, 0, 0},
			Mask: []byte{255, 255, 0, 0},
		},
		// RFC5735
		// 127.0.0.0/8
		{
			IP:   []byte{127, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		// RFC1122 Section 3.2.1.3
		// 0.0.0.0/8
		{
			IP:   []byte{0, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		// RFC3927
		// 169.254.0.0/16
		{
			IP:   []byte{169, 254, 0, 0},
			Mask: []byte{255, 255, 0, 0},
		},
		// RFC 5736
		// 192.0.0.0/24
		{
			IP:   []byte{192, 0, 0, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// RFC 5737
		// 192.0.2.0/24
		{
			IP:   []byte{192, 0, 2, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// 198.51.100.0/24
		{
			IP:   []byte{198, 51, 100, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// 203.0.113.0/24
		{
			IP:   []byte{203, 0, 113, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// RFC 3068
		// 192.88.99.0/24
		{
			IP:   []byte{192, 88, 99, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// RFC 2544, Errata 423
		// 198.18.0.0/15
		{
			IP:   []byte{198, 18, 0, 0},
			Mask: []byte{255, 254, 0, 0},
		},
		// RFC 3171
		// 224.0.0.0/4
		{
			IP:   []byte{224, 0, 0, 0},
			Mask: []byte{240, 0, 0, 0},
		},
		// RFC 1112
		// 240.0.0.0/4
		{
			IP:   []byte{240, 0, 0, 0},
			Mask: []byte{240, 0, 0, 0},
		},
		// RFC 919 Section 7
		// 255.255.255.255/32
		{
			IP:   []byte{255, 255, 255, 255},
			Mask: []byte{255, 255, 255, 255},
		},
		// RFC 6598
		// 100.64.0.0/10
		{
			IP:   []byte{100, 64, 0, 0},
			Mask: []byte{255, 192, 0, 0},
		},
	}
	// Sourced from https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
	// where Global, Source, or Destination is False
	privateV6Networks = []net.IPNet{
		parseCIDR("::/128", "RFC 4291: Unspecified Address"),
		parseCIDR("::1/128", "RFC 4291: Loopback Address"),
		parseCIDR("::ffff:0:0/96", "RFC 4291: IPv4-mapped Address"),
		parseCIDR("100::/64", "RFC 6666: Discard Address Block"),
		parseCIDR("2001::/23", "RFC 2928: IETF Protocol Assignments"),
		parseCIDR("2001:2::/48", "RFC 5180: Benchmarking"),
		parseCIDR("2001:db8::/32", "RFC 3849: Documentation"),
		parseCIDR("2001::/32", "RFC 4380: TEREDO"),
		parseCIDR("fc00::/7", "RFC 4193: Unique-Local"),
		parseCIDR("fe80::/10", "RFC 4291: Section 2.5.6 Link-Scoped Unicast"),
		parseCIDR("ff00::/8", "RFC 4291: Section 2.7"),
		// We disable validations to IPs under the 6to4 anycase prefix because
		// there's too much risk of a malicious actor advertising the prefix and
		// answering validations for a 6to4 host they do not control.
		// https://community.letsencrypt.org/t/problems-validating-ipv6-against-host-running-6to4/18312/9
		parseCIDR("2002::/16", "RFC 7526: 6to4 anycast prefix deprecated"),
	}
)

// parseCIDR parses the predefined CIDR to `net.IPNet` that consisting of IP and IPMask.
func parseCIDR(network string, comment string) net.IPNet {
	_, subNet, err := net.ParseCIDR(network)
	if err != nil {
		panic(fmt.Sprintf("error parsing %s (%s): %s", network, comment, err))
	}
	return *subNet
}

// isPrivateV4 checks whether an `ip` is private based on whether the IP is in the private CIDR range.
func isPrivateV4(ip net.IP) bool {
	for _, subNet := range privateNetworks {
		if subNet.Contains(ip) {
			return true
		}
	}
	return false
}

// isPrivateV6 checks whether an `ip` is private based on whether the IP is in the private CIDR range.
func isPrivateV6(ip net.IP) bool {
	for _, subNet := range privateV6Networks {
		if subNet.Contains(ip) {
			return true
		}
	}
	return false
}

// IsPrivateIP check IPv4 or IPv6 address according to the length of byte array
func IsPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return isPrivateV4(ip4)
	}
	return ip.To16() != nil && isPrivateV6(ip)
}

// IsIPValidAndPublic is a helper function check if an IP address is valid and public.
func IsIPValidAndPublic(ipAddr string) bool {
	if ipAddr == "" {
		return false
	}
	ipAddr = strings.TrimSpace(ipAddr)
	ip := net.ParseIP(ipAddr)
	// remote address within public address range
	if ip != nil && !IsPrivateIP(ip) {
		return true
	}
	return false
}

// GetUserRealIP Get User Real IP from headers of request `r`
//  1. First, determine whether the remote addr of request is a private address.
//     If it is a public network address, return it directly;
//  2. Otherwise, get and check the real IP from X-REAL-IP and X-Forwarded-For headers in turn.
//     if the header value contains multiple IP addresses separated by commas, that is,
//     the request may pass through multiple reverse proxies, we just keep the first one,
//     which imply it is the user connecting IP.
//     then we check the value is a valid public IP address using the `IsIPValidAndPublic` function.
//     If it is, the function returns the value as the client's real IP address.
//  3. Finally, If the above headers do not exist or are invalid, the remote addr is returned as is.
func GetUserRealIP(r *http.Request) string {
	fallbackAddr := r.RemoteAddr
	connectAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return fallbackAddr
	}
	if IsIPValidAndPublic(connectAddr) {
		return connectAddr
	}
	// in case that remote address is private(container or internal)
	for _, hd := range userRealIpHeaderCandidates {
		val := r.Header.Get(hd)
		if val == "" {
			continue
		}
		// remove leading or tailing comma, tab, space
		ipAddr := strings.Trim(val, ",\t ")
		if idxFirstIP := strings.Index(ipAddr, ","); idxFirstIP >= 0 {
			ipAddr = ipAddr[:idxFirstIP]
		}
		if IsIPValidAndPublic(ipAddr) {
			return ipAddr
		}
	}
	return fallbackAddr
}
