package model

// SliceDifference returns the elements that are in haystack but not in needle.
// It's a generic function that works with any comparable type.
func SliceDifference[T comparable](haystack, needle []T) []T {
	// Create a map of needle elements for quick lookup
	needleMap := make(map[T]bool)
	for _, item := range needle {
		needleMap[item] = true
	}

	// Find elements in haystack that are not in needle
	var difference []T
	for _, item := range haystack {
		if !needleMap[item] {
			difference = append(difference, item)
		}
	}

	return difference
}
