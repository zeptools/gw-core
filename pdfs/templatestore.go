package pdfs

type TemplateStore[T any] struct {
	templates map[string]T
}

func NewTemplateStore[T any]() *TemplateStore[T] {
	return &TemplateStore[T]{templates: make(map[string]T)}
}

func (s *TemplateStore[T]) Store(key string, template T) {
	s.templates[key] = template
}

func (s *TemplateStore[T]) Get(key string) (T, bool) {
	t, ok := s.templates[key]
	return t, ok
}

func (s *TemplateStore[T]) Remove(id string) {
	delete(s.templates, id)
}
