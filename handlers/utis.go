package handlers

import "regexp"

var (
	BookRe       = regexp.MustCompile(`^/books/*$`)
	BookReWithID = regexp.MustCompile(`^/books/([a-z0-9]+(?:-[a-z0-9]+)+)$`)
)
