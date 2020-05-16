package processor

import (
	"io/ioutil"
)

// ProcessGeneralFile process files that not HTML, JS or CSS.
func ProcessGeneralFile(req Request) (Resource, error) {
	// Read content from request input
	content, err := ioutil.ReadAll(req.Reader)
	if err != nil {
		return Resource{}, err
	}

	return createResource(content, req.URL, nil)
}
