package main

import (
	"net/url"
	"regexp"
)

var (
	BookRe         = regexp.MustCompile(`^/books/*$`)
	BookReWithID   = regexp.MustCompile(`^/books/([a-z0-9]+(?:-[a-z0-9]+)+)$`)
	AuthorRe       = regexp.MustCompile(`^/authors/*$`)
	AuthorReWithID = regexp.MustCompile(`^/authors/([a-z0-9]+(?:-[a-z0-9]+)+)$`)
	UserRe         = regexp.MustCompile(`^/users/*$`)
	UserReWithID   = regexp.MustCompile(`^/users/([a-z0-9]+(?:-[a-z0-9]+)+)$`)
	RegisterPath   = "/auth/register"
	LoginPath      = "/auth/login"
	LogoutPath     = "/auth/logout"
)

var params = map[string]map[string]bool{
	"author": {"book_name": true, "author_name": true, "genre": true},
	"book":   {"book_name": true, "genre": true, "publication_date": true, "author_name": true},
	"user":   {"name": true, "mail": true, "role": true},
}

func ToMap(values url.Values) map[string]string {
	res := make(map[string]string)
	for k := range values {
		res[k] = values.Get(k)
	}
	return res
}

func ValidParams(api string, m map[string]string) bool {
	var available_params = params[api]
	for k := range m {
		if _, ok := available_params[k]; !ok {
			return false
		}
	}

	return true
}
