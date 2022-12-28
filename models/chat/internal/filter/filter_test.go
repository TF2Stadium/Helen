package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageFilter(t *testing.T) {
	cases := []struct {
		filteredWords   []string
		shouldFilter    []string
		shouldNotFilter []string
	}{{[]string{"foo", "bar"},
		[]string{"sentence with foo", "sentence_with_foo",
			"spaces f o o", "bar foo", "f\to\to"},
		[]string{"baz"}}}

	for _, tc := range cases {
		for _, filtered := range tc.shouldFilter {
			assert.Truef(t,
				FilteredMessage(filtered, tc.filteredWords),
				"%s should be filtered", filtered)
		}
		for _, filtered := range tc.shouldNotFilter {
			assert.Falsef(t,
				FilteredMessage(filtered, tc.filteredWords),
				"%s should not be filtered", filtered)
		}
	}
}
