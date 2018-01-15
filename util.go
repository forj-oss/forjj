package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"forjj/utils"
)

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
