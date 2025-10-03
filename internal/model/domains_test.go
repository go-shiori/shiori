package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBookmarkSearchOrderMethod_Values(t *testing.T) {
	tests := []struct {
		name     string
		order    BookmarkSearchOrderMethod
		expected int
	}{
		{
			name:     "DefaultSearchOrder should be 0",
			order:    DefaultSearchOrder,
			expected: 0,
		},
		{
			name:     "ByLastAddedSearchOrder should be 1",
			order:    ByLastAddedSearchOrder,
			expected: 1,
		},
		{
			name:     "ByLastModifiedSearchOrder should be 2",
			order:    ByLastModifiedSearchOrder,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, int(tt.order))
		})
	}
}

func TestBookmarksSearchOptions_ToDBGetBookmarksOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    BookmarksSearchOptions
		expected DBGetBookmarksOptions
	}{
		{
			name: "Complete conversion with all fields",
			input: BookmarksSearchOptions{
				IDs:          []int{1, 2, 3},
				Tags:         []string{"tag1", "tag2"},
				ExcludedTags: []string{"exclude1"},
				Keyword:      "test keyword",
				WithContent:  true,
				OrderMethod:  ByLastAddedSearchOrder,
				Limit:        100,
				Offset:       50,
			},
			expected: DBGetBookmarksOptions{
				IDs:          []int{1, 2, 3},
				Tags:         []string{"tag1", "tag2"},
				ExcludedTags: []string{"exclude1"},
				Keyword:      "test keyword",
				WithContent:  true,
				OrderMethod:  ByLastAdded,
				Limit:        100,
				Offset:       50,
			},
		},
		{
			name: "Minimal conversion with zero values",
			input: BookmarksSearchOptions{
				Keyword:     "minimal",
				OrderMethod: DefaultSearchOrder,
			},
			expected: DBGetBookmarksOptions{
				Keyword:     "minimal",
				OrderMethod: DefaultOrder,
				Limit:       0,
				Offset:      0,
			},
		},
		{
			name: "Conversion with all order methods",
			input: BookmarksSearchOptions{
				Keyword:     "order test",
				OrderMethod: ByLastModifiedSearchOrder,
			},
			expected: DBGetBookmarksOptions{
				Keyword:     "order test",
				OrderMethod: ByLastModified,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToDBGetBookmarksOptions()

			assert.Equal(t, tt.expected.IDs, result.IDs)
			assert.Equal(t, tt.expected.Tags, result.Tags)
			assert.Equal(t, tt.expected.ExcludedTags, result.ExcludedTags)
			assert.Equal(t, tt.expected.Keyword, result.Keyword)
			assert.Equal(t, tt.expected.WithContent, result.WithContent)
			assert.Equal(t, tt.expected.OrderMethod, result.OrderMethod)
			assert.Equal(t, tt.expected.Limit, result.Limit)
			assert.Equal(t, tt.expected.Offset, result.Offset)
		})
	}
}

func TestBookmarksSearchOptions_ToDBGetBookmarksOptions_OrderMethodMapping(t *testing.T) {
	// Test all order method mappings
	orderMappings := []struct {
		domainOrder   BookmarkSearchOrderMethod
		databaseOrder DBOrderMethod
	}{
		{DefaultSearchOrder, DefaultOrder},
		{ByLastAddedSearchOrder, ByLastAdded},
		{ByLastModifiedSearchOrder, ByLastModified},
	}

	for _, mapping := range orderMappings {
		t.Run(
			fmt.Sprintf("order mapping %d to %d", mapping.domainOrder, mapping.databaseOrder),
			func(t *testing.T) {
				options := BookmarksSearchOptions{
					OrderMethod: mapping.domainOrder,
				}

				result := options.ToDBGetBookmarksOptions()
				assert.Equal(t, mapping.databaseOrder, result.OrderMethod)
			},
		)
	}
}
