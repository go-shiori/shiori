package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractSingleFileMetadata(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		wantURL   string
		wantTitle string
	}{
		{
			name: "saved from url comment",
			html: `<!DOCTYPE html>
<!-- saved from url=(0042)https://www.example.com/article?id=1 -->
<html><head><title>Example Article</title></head><body></body></html>`,
			wantURL:   "https://www.example.com/article?id=1",
			wantTitle: "Example Article",
		},
		{
			name: "block comment url line",
			html: `<!DOCTYPE html>
<!--
 Page saved with SingleFile
 url: https://www.example.com/page
 saved date: Mon, 01 Jan 2024
-->
<html><head><title>My Page</title></head><body></body></html>`,
			wantURL:   "https://www.example.com/page",
			wantTitle: "My Page",
		},
		{
			name: "base href fallback",
			html: `<html><head>
<base href="https://www.example.com/base">
<title>Base Test</title>
</head><body></body></html>`,
			wantURL:   "https://www.example.com/base",
			wantTitle: "Base Test",
		},
		{
			name: "canonical link fallback",
			html: `<html><head>
<link rel="canonical" href="https://www.example.com/canonical">
<title>Canonical Test</title>
</head><body></body></html>`,
			wantURL:   "https://www.example.com/canonical",
			wantTitle: "Canonical Test",
		},
		{
			name: "og:url fallback",
			html: `<html><head>
<meta property="og:url" content="https://www.example.com/og">
<meta property="og:title" content="OG Title">
</head><body></body></html>`,
			wantURL:   "https://www.example.com/og",
			wantTitle: "OG Title",
		},
		{
			name: "twitter:url fallback",
			html: `<html><head>
<meta name="twitter:url" content="https://www.example.com/twitter">
<title>Twitter Test</title>
</head><body></body></html>`,
			wantURL:   "https://www.example.com/twitter",
			wantTitle: "Twitter Test",
		},
		{
			name: "title falls back to og:title when no title tag",
			html: `<html><head>
<!-- saved from url=(0030)https://www.example.com/ -->
<meta property="og:title" content="OG Only Title">
</head><body></body></html>`,
			wantURL:   "https://www.example.com/",
			wantTitle: "OG Only Title",
		},
		{
			name:      "no url found returns empty",
			html:      `<html><head><title>No URL</title></head><body></body></html>`,
			wantURL:   "",
			wantTitle: "No URL",
		},
		{
			name:      "comment marker takes priority over base href",
			html: `<!DOCTYPE html>
<!-- saved from url=(0040)https://www.example.com/primary -->
<html><head>
<base href="https://www.example.com/base">
<title>Priority Test</title>
</head><body></body></html>`,
			wantURL:   "https://www.example.com/primary",
			wantTitle: "Priority Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSingleFileMetadata([]byte(tt.html))
			assert.Equal(t, tt.wantURL, got.URL)
			assert.Equal(t, tt.wantTitle, got.Title)
		})
	}
}
