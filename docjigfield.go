package mdd

type StructField[T any] struct {
	fieldName string
	key       string
	origKey   string
	required  bool
	convert   func(value string, t *T) (any, error)
}

func (s *StructField[T]) Alias(alias ...string) *StructField[T] {
	return s
}

func (s *StructField[T]) Required() *StructField[T] {
	s.required = true
	return s
}

func (s *StructField[T]) As(convert func(value string, t *T) (any, error)) *StructField[T] {
	s.convert = convert
	return s
}
