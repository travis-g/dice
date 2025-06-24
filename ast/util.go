package main

// ptr returns a pointer to the passed value.
func ptr[T any](v T) *T {
	return &v
}
