package utils

import (
	"net/url"
	"io/ioutil"
	"net/http"
	"fmt"
	"strings"
	"github.com/forj-oss/forjj-modules/trace"
)


// Function to read a document from a url like github raw or directly from a local path
func ReadDocumentFrom(s *url.URL) ([]byte, error) {
	if s.Scheme == "" {
		// File to read locally
		return read_document_from_fs(s.Path)
	}
	// File to read from an url. Usually, a raw from github.
	return read_document_from_url(s.String())
}

// Read from the filesystem. If the path start with ~, replaced by the user homedir. In some context, this won't work well, like in container.
func read_document_from_fs(source string) (_ []byte, err error) {
	if source, err = Abs(source) ; err != nil {
		return
	}
	gotrace.Trace("Load file definition at '%s'", source)
	return ioutil.ReadFile(source)
}

// Read from the URL string. Data is returned is content type is of text/plain
func read_document_from_url(source string) (yaml_data []byte, err error) {
	gotrace.Trace("Load file definition at '%s'", source)

	var resp *http.Response
	if resp, err = http.Get(source); err != nil {
		err = fmt.Errorf("Unable to read '%s'. %s\n", source, err)
		return
	}
	defer resp.Body.Close()

	var d []byte
	if d, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	if strings.Contains(http.DetectContentType(d), "text/plain") {
		yaml_data = d
	}
	return
}
