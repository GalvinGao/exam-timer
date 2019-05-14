package main

import "strings"

func sanitize(s string) string {
	for _, str := range badPaths {
		s = strings.ReplaceAll(s, str, "-")
	}
	return s
}
