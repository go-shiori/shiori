package core

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryEncodeWithoutEmptyValues(t *testing.T) {
	t.Run("Encodes single key-value pair", func(t *testing.T) {
		assert.Equal(
			t, // t

			"q=shiori", // expected
			queryEncodeWithoutEmptyValues(url.Values{"q": {"shiori"}})) // actual
	})

	t.Run("Omits empty values", func(t *testing.T) {
		assert.Equal(
			t, // t

			"q=shiori&utm", // expected
			queryEncodeWithoutEmptyValues(
				url.Values{
					"q":   {"shiori"},
					"utm": {""},
				},
			), // actual
		)
	})

	t.Run("Handles multiple keys in order", func(t *testing.T) {
		assert.Equal(
			t, // t

			"a=first&d&z=last", // expected
			queryEncodeWithoutEmptyValues(
				url.Values{
					"z": {"last"},
					"a": {"first"},
					"d": {""},
				},
			), // actual
		)
	})

	t.Run("Nil values map returns empty string", func(t *testing.T) {
		assert.Equal(
			t, // t

			"", // expected
			queryEncodeWithoutEmptyValues(url.Values{}), // actual
		)
	})
}

func TestUrlSchemeOk(t *testing.T) {
	var cases []struct {
		reqUrl   string
		expected bool
	}

	/*** * * ***/

	cases = []struct {
		reqUrl   string
		expected bool
	}{
		// Explicit http(s) schemes
		{"https://example.com", true},
		{"http://example.com", true},
		{"https://example", true},
		{"http://example", true},

		// Other schemes with ://
		{"ftp://example.com", true},
		{"custom+scheme://resource", true},
		{"git+ssh://github.com/user/repo", true},

		// Reject exactly "://"
		{"://", false},

		// Reject those *starting* exactly with "://"
		{"://example.com", false},
		{"://example", false},

		// Reject those *starting* exactly with ":/" (one /)
		{":/example.com", false},
		{":/example", false},

		// No scheme, no : before //, nothing(!)
		{"example.com", false},
		{"example", false},
		{"//example.com", false},
		{"//example", false},
		{"", false},

		// Edge cases: missing slash in http(s)
		{"https:/example.com", false},
		{"https:/example", false},
		{"http:/example.com", false},
		{"http:/example", false},

		// Accept random with scheme before ://
		{"a://", true},
		{"1://", true},
		{"!://", true},
		{"abcdefg://", true},
		{"12345://", true},
		{"!@#$%://", true},

		// Reject random without scheme
		{"a", false},
		{"1", false},
		{"!", false},
		{"abcdefg", false},
		{"12345", false},
		{"!@#$%", false},
	}

	/*** * * ***/

	for _, c := range cases {
		t.Run(c.reqUrl, func(t *testing.T) {
			assert.Equal(
				t, // t

				c.expected,            // expected
				urlSchemeOk(c.reqUrl), // actual
			)
		})
	}
}

func TestParse(t *testing.T) {
	t.Run("Parses full URL with scheme", func(t *testing.T) {
		var resUrl *url.URL
		var resErr error

		/*** * * ***/

		resUrl, resErr = Parse("http://example.com/page")

		/*** * * ***/

		assert.NoError(t, resErr)

		assert.Equal(
			t, // t

			"http",        // expected
			resUrl.Scheme, // actual
		)

		assert.Equal(
			t, // t

			"example.com", // expected
			resUrl.Host,   // actual
		)
	})

	t.Run("Adds https scheme if missing", func(t *testing.T) {
		var resUrl *url.URL
		var resErr error

		/*** * * ***/

		resUrl, resErr = Parse("example.com/page")

		/*** * * ***/

		assert.NoError(t, resErr)

		assert.Equal(
			t, // t

			"https",       // expected
			resUrl.Scheme, // actual
		)

		assert.Equal(
			t, // t

			"example.com", // expected
			resUrl.Host,   // actual
		)
	})
}

func TestRemoveUTMParams(t *testing.T) {
	t.Run("Removes UTM parameters", func(t *testing.T) {
		var resStr string
		var resErr error

		/*** * * ***/

		resStr, resErr = RemoveUTMParams("https://example.com/article?utm_source=newsletter&utm_medium=email&q=go")

		/*** * * ***/

		assert.NoError(t, resErr)

		assert.Equal(
			t, // t

			"https://example.com/article?q=go", // expected
			resStr,                             // actual
		)
	})

	t.Run("Returns original URL on parse error", func(t *testing.T) {
		const REQ_URL_INVALID string = "http://[::1]:namedport"

		/*** * * ***/

		var resStr string
		var resErr error

		/*** * * ***/

		resStr, resErr = RemoveUTMParams(REQ_URL_INVALID)

		/*** * * ***/

		assert.Error(t, resErr)

		assert.Equal(
			t, // t

			REQ_URL_INVALID, // expected
			resStr,          // actual
		)
	})

	t.Run("Preserves URL with no utm_* params", func(t *testing.T) {
		const REQ_URL string = "https://example.com/path?q=test"

		/*** * * ***/

		var resUrl string
		var resErr error

		/*** * * ***/

		resUrl, resErr = RemoveUTMParams(REQ_URL)

		/*** * * ***/

		assert.NoError(t, resErr)

		assert.Equal(
			t, // t

			REQ_URL, // expected
			resUrl,  // actual
		)
	})
}
