package util

import (
	"regexp"
	"bytes"
)

// Regexp definitions
var keyMatchRegex = regexp.MustCompile(`\"(\w+)\":`)
var wordBarrierRegex = regexp.MustCompile(`(\w)([A-Z])`)

func ConvertJSONKeyCase(rawJSON []byte) []byte {
	convertedJSON := keyMatchRegex.ReplaceAllFunc(
		rawJSON,
		func(match []byte) []byte {
			return bytes.ToLower(wordBarrierRegex.ReplaceAll(
				match,
				[]byte(`${1}_${2}`),
			))
		},
	)

	return convertedJSON
}
