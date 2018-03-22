package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
)

const PluginTag = "<plugin>"

// ReadDocumentFrom is used to read a document from a url like github raw or directly from a local path
// It supports file or url stored in url.URL structure
// each urls can be defined with a plugin tag "<plugin>" which will be replaced by the document name(document)
// the file name and the extension is added at the end of the string.
func ReadDocumentFrom(urls []*url.URL, extension, document, docType string) ([]byte, error) {
	if urls == nil {
		return nil, fmt.Errorf("url parameter is nil")
	}
	for _, s := range urls {
		var fileName string
		if s.Scheme == "" {
			// File to read locally
			fileName = BuildURLPath(s.Path, docType, document, extension)

			if found, data, err := readDocumentFromFS(fileName); err != nil || found {
				return data, err
			}
			continue
		}
		// File to read from an url. Usually, a raw from github.
		fileName = BuildURLPath(s.String(), docType, document, extension)
		if found, data, err := readDocumentFromURL(s.String()); err != nil || found {
			return data, err
		}
	}
	return nil, fmt.Errorf("Document not found from URLs given")
}

// BuildURLPath build the path logic introducing the pluginTag to replace.
//
func BuildURLPath(aPath, docType, document, extension string) (fullFileName string) {
	fileName := document + extension

	if strings.Contains(aPath, PluginTag) {
		aPath = strings.Replace(aPath, PluginTag, document, -1)
	} else {
		if docType != "" {
			aPath = path.Join(aPath, docType, document)
		}
	}
	fullFileName = path.Join(aPath, fileName)
	return
}

// readDocumentFromFS read from the filesystem. If the path start with ~, replaced by the user homedir. In some context, this won't work well, like in container.
func readDocumentFromFS(source string) (found bool, data []byte, err error) {
	if source, err = Abs(source); err != nil {
		return
	}
	if _, err = os.Stat(source); err != nil {
		return
	}
	gotrace.Trace("Load file definition at '%s'", source)
	if data, err = ioutil.ReadFile(source); err == nil {
		found = true
	}
	return
}

// readDocumentFromUrl Read from the URL string. Data is returned is content type is of text/plain
func readDocumentFromURL(source string) (found bool, yamlData []byte, err error) {
	gotrace.Trace("Load file definition at '%s'", source)

	var resp *http.Response
	if resp, err = http.Get(source); err != nil {
		err = fmt.Errorf("Unable to read '%s'. %s", source, err)
		return
	}
	found = true
	defer resp.Body.Close()

	var d []byte
	if d, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	if strings.Contains(http.DetectContentType(d), "text/plain") {
		yamlData = d
	}
	return
}
