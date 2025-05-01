package torznab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryEncode(t *testing.T) {
	for _, row := range []struct {
		left, right Query
	}{
		{Query{}, Query{}},
		{Query{Type: "search", Q: "the llama show"}, Query{Q: "the llama show"}},
	} {
		assert.Equal(t, row.left.Encode(), row.right.Encode())
	}
}
