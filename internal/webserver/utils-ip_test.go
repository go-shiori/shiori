package webserver

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivateIP(t *testing.T) {
	assert.True(t, IsPrivateIP(net.ParseIP("127.0.0.1")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("192.168.254.254")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("10.255.0.3")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("172.16.255.255")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("172.31.255.255")), "should be private")
	assert.True(t, !IsPrivateIP(net.ParseIP("128.0.0.1")), "should be private")
	assert.True(t, !IsPrivateIP(net.ParseIP("192.169.255.255")), "should not be private")
	assert.True(t, !IsPrivateIP(net.ParseIP("9.255.0.255")), "should not be private")
	assert.True(t, !IsPrivateIP(net.ParseIP("172.32.255.255")), "should not be private")

	assert.True(t, IsPrivateIP(net.ParseIP("::0")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("::1")), "should be private")
	assert.True(t, !IsPrivateIP(net.ParseIP("::2")), "should not be private")

	assert.True(t, IsPrivateIP(net.ParseIP("fe80::1")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("febf::1")), "should be private")
	assert.True(t, !IsPrivateIP(net.ParseIP("fec0::1")), "should not be private")
	assert.True(t, !IsPrivateIP(net.ParseIP("feff::1")), "should not be private")

	assert.True(t, IsPrivateIP(net.ParseIP("ff00::1")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("ff10::1")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")), "should be private")

	assert.True(t, IsPrivateIP(net.ParseIP("2002::")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("2002:ffff:ffff:ffff:ffff:ffff:ffff:ffff")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("0100::")), "should be private")
	assert.True(t, IsPrivateIP(net.ParseIP("0100::0000:ffff:ffff:ffff:ffff")), "should be private")
	assert.True(t, !IsPrivateIP(net.ParseIP("0100::0001:0000:0000:0000:0000")), "should be private")
}

func TestIsIpValidAndPublic(t *testing.T) {
	// test empty address
	assert.False(t, isIpValidAndPublic(""))
	// test public address
	assert.True(t, isIpValidAndPublic("31.41.244.124"))
	assert.True(t, isIpValidAndPublic("62.233.50.248"))
	// trim head or tail space
	assert.True(t, isIpValidAndPublic(" 62.233.50.249"))
	assert.True(t, isIpValidAndPublic(" 62.233.50.250 "))
	assert.True(t, isIpValidAndPublic("62.233.50.251 "))
	// test private address
	assert.False(t, isIpValidAndPublic("10.1.123.52"))
	assert.False(t, isIpValidAndPublic("192.168.123.24"))
	assert.False(t, isIpValidAndPublic("172.17.0.1"))
}
