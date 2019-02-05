package creds

import "github.com/forj-oss/goforjj"

type YamlValue struct {
	Value    *goforjj.ValueStruct
	Resource map[string]string
	Source   string
}
