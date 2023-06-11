package slice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlice_chunk(t *testing.T) {
	r := require.New(t)

	s := []int{1, 2, 3, 4, 5}
	r.Equal([][]int{{1, 2, 3}, {4, 5}}, Chunk(s, 3))
}
