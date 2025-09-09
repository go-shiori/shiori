package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceDifference(t *testing.T) {
	t.Run("empty_slices", func(t *testing.T) {
		result := SliceDifference([]int{}, []int{})
		assert.Empty(t, result, "Difference of empty slices should be empty")
	})

	t.Run("empty_haystack", func(t *testing.T) {
		result := SliceDifference([]int{}, []int{1, 2, 3})
		assert.Empty(t, result, "Difference with empty haystack should be empty")
	})

	t.Run("empty_needle", func(t *testing.T) {
		result := SliceDifference([]int{1, 2, 3}, []int{})
		assert.Equal(t, []int{1, 2, 3}, result, "Difference with empty needle should be the haystack")
	})

	t.Run("no_difference", func(t *testing.T) {
		result := SliceDifference([]int{1, 2, 3}, []int{1, 2, 3})
		assert.Empty(t, result, "Difference of identical slices should be empty")
	})

	t.Run("partial_difference", func(t *testing.T) {
		result := SliceDifference([]int{1, 2, 3, 4}, []int{2, 4})
		assert.Equal(t, []int{1, 3}, result, "Should return elements in haystack but not in needle")
	})

	t.Run("complete_difference", func(t *testing.T) {
		result := SliceDifference([]int{1, 2, 3}, []int{4, 5, 6})
		assert.Equal(t, []int{1, 2, 3}, result, "Should return all elements from haystack when needle has no common elements")
	})

	t.Run("with_duplicates", func(t *testing.T) {
		result := SliceDifference([]int{1, 2, 2, 3, 3, 3}, []int{2, 3})
		assert.Equal(t, []int{1}, result, "Should handle duplicates correctly")
	})

	t.Run("string_type", func(t *testing.T) {
		result := SliceDifference([]string{"a", "b", "c"}, []string{"b"})
		assert.Equal(t, []string{"a", "c"}, result, "Should work with string type")
	})
}
