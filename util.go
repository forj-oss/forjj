package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/user"
	"regexp"
	"strconv"
	"strings"
)


// Simple function to convert a dynamic type to bool
// it returns false by default except if the internal type is:
// - bool. value as is
// - string: call https://golang.org/pkg/strconv/#ParseBool
//
func to_bool(v interface{}) bool {
	switch v.(type) {
	case bool:
		return v.(bool)
	case string:
		s := v.(string)
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
		return false
	}
	return false
}

// simply extract string from the dynamic type
// otherwise the returned string is empty.
func to_string(v interface{}) (result string) {
	switch v.(type) {
	case string:
		return v.(string)
	}
	return
}

// Function to read a document from a url like github raw or directly from a local path
func read_document_from(s *url.URL) (yaml_data []byte, err error) {
	if s.Scheme == "" {
		// File to read locally
		return read_document_from_fs(s.Path)
	}
	// File to read from an url. Usually, a raw from github.
	return read_document_from_url(s.String())
}

// Read from the filesystem. If the path start with ~, replaced by the user homedir. In some context, this won't work well, like in container.
func read_document_from_fs(source string) (yaml_data []byte, err error) {
	// File to read locally
	if source[:1] == "~" {
		cur_user := &user.User{}
		if cur_user, err = user.Current(); err != nil {
			err = fmt.Errorf("Unable to get your user. %s. Consider to replace ~ by $HOME\n", err)
			return
		}
		source = string(regexp.MustCompile("^~").ReplaceAll([]byte(source), []byte(cur_user.HomeDir)))
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

func MapBoolKeys(m map[string]bool) (a []string) {
	a = make([]string, 0, len(m))
	for key := range m {
		a = append(a, key)
	}
	return a
}

func arrayStringDelete(a []string, element string) []string {
	for index, value := range a {
		if value == element {
			return append(a[:index], a[index+1:]...)
		}
	}
	return a
}

func inStringList(element string, elements ...string) string {
	for _, value := range elements {
		if element == value {
			return value
		}
	}
	return ""
}
