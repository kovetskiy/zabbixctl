package main

import (
	"regexp"
	"strings"
)

func parseSearchQuery(targets []string) (words []string, pattern string) {
	var (
		search bool
		query  []string
	)

	for _, target := range targets {
		if strings.HasPrefix(target, "/") {
			query = append(
				query,
				strings.TrimPrefix(target, "/"),
			)

			search = true

			continue
		}

		if search {
			query = append(query, target)
		} else {
			words = append(words, target)
		}
	}

	return words, getSearchPattern(query)
}

func getSearchPattern(query []string) string {
	letters := strings.Split(
		strings.Replace(
			strings.Join(query, ""),
			" ", "", -1,
		),
		"",
	)
	for i, letter := range letters {
		letters[i] = regexp.QuoteMeta(letter)
	}

	pattern := strings.Join(letters, ".*")

	return pattern
}

func matchPattern(pattern, target string) bool {
	match, err := regexp.MatchString(strings.ToLower(pattern),
		strings.ToLower(target))
	if err != nil {
		debugf("Error: %+v", err)
	}
	return match
}
