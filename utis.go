package main

import (
	"net/url"
	"regexp"
)

var (
	BookRe       = regexp.MustCompile(`^/books/*$`)
	BookReWithID = regexp.MustCompile(`^/books/([a-z0-9]+(?:-[a-z0-9]+)+)$`)
)

var params = map[string]bool{"name": true, "genre": true, "publication_date": true}

func ToMap(values url.Values) map[string]string {
	res := make(map[string]string)
	for k := range values {
		res[k] = values.Get(k)
	}
	return res
}

func ValidParams(m map[string]string) bool {
	for k := range m {
		if _, ok := params[k]; !ok {
			return false
		}
	}

	return true
}
