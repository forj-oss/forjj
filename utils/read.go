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

// PluginTag to identify variable component name in a Document path.
const RepoTag = "<repo>"

// ReadDocumentFrom is used to read a document from a url like github raw or directly from a local path
// It supports file or url stored in url.URL structure
// each urls can be defined with a plugin tag "<plugin>" which will be replaced by the document name(document)
// the file name and the extension is added at the end of the string.
func ReadDocumentFrom(urls []*url.URL, repos, subPaths []string, document, contentType string) ([]byte, error) {
	if urls == nil {
		return nil, fmt.Errorf("url parameter is nil")
	}
	if contentType == "" {
		contentType = "text/plain"
	}
	for _, s := range urls {
		for i, repo := range repos {
			var fileName string
			if s.Scheme == "" {
				// File to read locally
				if !strings.Contains(s.Path, RepoTag) {
					s.Path = path.Join(s.Path, RepoTag)
				}
				fileName = BuildURLPath(s.Path, repo, subPaths[i], document)
				gotrace.Trace("Searching file document '%s'", fileName)

				if found, data, err := readDocumentFromFS(fileName); found {
					return data, err
				}
				continue
			}
			// File to read from an url. Usually, a raw from github.
			urlData := ""
			if u, err := url.PathUnescape(s.String()); err != nil {
				return nil, fmt.Errorf("Url path issue: %s", err)
			} else {
				urlData = u
			}
			fileName = BuildURLPath(urlData, repo, subPaths[i], document)
			gotrace.Trace("Searching file document from url '%s'", fileName)
			if found, data, err := readDocumentFromURL(fileName, contentType); found {
				return data, err
			}
		}
	}
	return nil, fmt.Errorf("Document not found from URLs given")
}

// BuildURLPath build the path logic introducing the pluginTag to replace.
//
func BuildURLPath(aPath, repo, subpath, document string) (fullFileName string) {

	if strings.Contains(aPath, RepoTag) {
		aPath = strings.Replace(aPath, RepoTag, repo, -1)
	}
	// path.Join is not usable on a url as it replaces // by /
	fullFileName = aPath + "/" + path.Join(subpath, document)
	if document[len(document)-1] == '/' {
		fullFileName += "/"
	}

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
	if data, err = ioutil.ReadFile(source); err == nil {
		found = true
		gotrace.Trace("Loaded file definition at '%s'", source)
	}
	return
}

// readDocumentFromUrl Read from the URL string. Data is returned is content type is of text/plain
func readDocumentFromURL(source, contentType string) (found bool, yamlData []byte, err error) {
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
	if strings.Contains(http.DetectContentType(d), contentType) {
		yamlData = d
		gotrace.Trace("Loaded file definition at '%s'", source)
	}
	return
}
