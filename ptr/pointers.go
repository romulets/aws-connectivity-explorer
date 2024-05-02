package ptr

//func Ref[T any](v T) *T {
//	return &v
//}

func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}

	return *p
}

func IsEmpty[T comparable](p *T) bool {
	var zero T
	return p == nil || *p == zero
}
