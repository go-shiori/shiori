package webserver

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCidr(t *testing.T) {
	res := parseCIDR("192.168.0.0/16", "internal 192.168.x.x")
	assert.Equal(t, res.IP, net.IP([]byte{192, 168, 0, 0}))
	assert.Equal(t, res.Mask, net.IPMask([]byte{255, 255, 0, 0}))
}

func TestParseCidrInvalidAddr(t *testing.T) {
	assert.Panics(t, func() { parseCIDR("192.168.0.0/34", "internal 192.168.x.x") })
}

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
	assert.False(t, IsIPValidAndPublic(""))
	// test public address
	assert.True(t, IsIPValidAndPublic("31.41.244.124"))
	assert.True(t, IsIPValidAndPublic("62.233.50.248"))
	// trim head or tail space
	assert.True(t, IsIPValidAndPublic(" 62.233.50.249"))
	assert.True(t, IsIPValidAndPublic(" 62.233.50.250 "))
	assert.True(t, IsIPValidAndPublic("62.233.50.251 "))
	// test private address
	assert.False(t, IsIPValidAndPublic("10.1.123.52"))
	assert.False(t, IsIPValidAndPublic("192.168.123.24"))
	assert.False(t, IsIPValidAndPublic("172.17.0.1"))
}

func BenchmarkIsPrivateIPv4(b *testing.B) {
	// range: 2-254
	n1 := 2 + rand.Intn(252)
	n2 := 2 + rand.Intn(252)
	for i := 0; i < b.N; i++ {
		IsPrivateIP(net.ParseIP(fmt.Sprintf("192.168.%d.%d", n1, n2)))
	}
}

func BenchmarkIsPrivateIPv6(b *testing.B) {
	n1 := 2 + rand.Intn(252)
	for i := 0; i < b.N; i++ {
		IsPrivateIP(net.ParseIP(fmt.Sprintf("2002::%d", n1)))
	}
}

func testIsPublicHttpRequestAddressHelper(
	t *testing.T, wantIP string, headers map[string]string, isPublic bool,
) {
	testIsPublicHttpRequestAddressHelperWrapped(t, nil, wantIP, headers, isPublic)
}

func testIsPublicHttpRequestAddressHelperWrapped(
	t *testing.T, r *http.Request, wantIP string, headers map[string]string, isPublic bool,
) {
	var (
		err    error
		userIP string
	)
	if r == nil {
		r = httptest.NewRequest("GET", "/", nil)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}

	origVal := GetUserRealIP(r)
	if strings.Contains(origVal, ":") {
		userIP, _, err = net.SplitHostPort(origVal)
		if err != nil {
			t.Error(err)
		}
	} else {
		userIP = origVal
	}

	if isPublic {
		// should equal first ip in list
		assert.Equal(t, wantIP, userIP)
		assert.True(t, IsIPValidAndPublic(userIP))
	} else {
		assert.Equal(t, origVal, r.RemoteAddr)
		assert.False(t, IsIPValidAndPublic(userIP))
	}
}

func TestGetUserRealIPWithSetRemoteAddr(t *testing.T) {
	// Test Public RemoteAddr
	testIsPublicHttpRequestAddressHelper(t, "", nil, false)

	r := httptest.NewRequest("GET", "/", nil)
	wantIP := "34.23.123.122"
	r.RemoteAddr = fmt.Sprintf("%s:1234", wantIP)
	testIsPublicHttpRequestAddressHelperWrapped(t, r, wantIP, nil, true)
}

func TestGetUserRealIPWithInvalidRemoteAddr(t *testing.T) {
	// Test Public RemoteAddr
	testIsPublicHttpRequestAddressHelper(t, "", nil, false)

	r := httptest.NewRequest("GET", "/", nil)
	wantIP := "34.23.123.122"
	// without port
	r.RemoteAddr = wantIP
	testIsPublicHttpRequestAddressHelperWrapped(t, r, wantIP, nil, true)
}

func TestGetUserRealIPWithEmptyHeader(t *testing.T) {
	// Test Empty X-Real-IP
	testIsPublicHttpRequestAddressHelper(t, "", nil, false)
}

func TestGetUserRealIPWithInvalidHeaderValue(t *testing.T) {
	for _, name := range userRealIpHeaderCandidates {
		// invalid ip
		m := map[string]string{
			name: "31.41.24a.12",
		}
		testIsPublicHttpRequestAddressHelper(t, "", m, false)
	}
}

func TestGetUserRealIPWithXRealIpHeader(t *testing.T) {
	// Test public Real IP
	for _, name := range userRealIpHeaderCandidates {
		wantIP := "31.41.242.12"
		m := map[string]string{
			name: wantIP,
		}
		testIsPublicHttpRequestAddressHelper(t, wantIP, m, true)
	}
}

func TestGetUserRealIPWithPrivateXRealIpHeader(t *testing.T) {
	for _, name := range userRealIpHeaderCandidates {
		wantIP := "192.168.123.123"
		// test private ip in header
		m := map[string]string{
			name: wantIP,
		}
		testIsPublicHttpRequestAddressHelper(t, wantIP, m, false)
	}
}

func TestGetUserRealIPWithXRealIpListHeader(t *testing.T) {
	// Test Real IP List
	for _, name := range userRealIpHeaderCandidates {
		ipList := []string{"34.23.123.122", "34.23.123.123"}
		// should equal first ip in list
		wantIP := ipList[0]
		// test private ip in header
		m := map[string]string{
			name: strings.Join(ipList, ", "),
		}
		testIsPublicHttpRequestAddressHelper(t, wantIP, m, true)
	}
}

func TestGetUserRealIPWithXRealIpHeaderIgnoreComma(t *testing.T) {
	// Test Real IP List with leading or tailing comma
	wantIP := "34.23.123.124"
	ipVariants := []string{
		",34.23.123.124", " ,34.23.123.124", "\t,34.23.123.124",
		",34.23.123.124,", " ,34.23.123.124, ", "\t,34.23.123.124,\t",
		"34.23.123.124,", "34.23.123.124, ", "34.23.123.124,\t"}
	for _, variant := range ipVariants {
		for _, name := range userRealIpHeaderCandidates {
			m := map[string]string{name: variant}
			testIsPublicHttpRequestAddressHelper(t, wantIP, m, true)
		}
	}
}

func TestGetUserRealIPWithDifferentHeaderOrder(t *testing.T) {
	var m map[string]string
	wantIP := "34.23.123.124"
	m = map[string]string{
		"X-Real-Ip":       "192.168.123.122",
		"X-Forwarded-For": wantIP,
	}

	testIsPublicHttpRequestAddressHelper(t, wantIP, m, true)
	m = map[string]string{
		"X-Real-Ip":       wantIP,
		"X-Forwarded-For": "192.168.123.122",
	}
	testIsPublicHttpRequestAddressHelper(t, wantIP, m, true)
}
