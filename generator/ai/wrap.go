package ai

func Wrap[T any](v T) *T {
	return &v
}
