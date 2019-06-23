package structs

type Set struct {
	items map[interface{}]struct{}
}

func Initialise() *Set {
	return &Set{make(map[interface{}]struct{}, 0)}
}
func (s *Set) Add(item interface{}) {
	s.items[item] = struct{}{}
}

func (s *Set) Delete(item interface{}) {
	delete(s.items, item)
}

func (s *Set) Exists(item interface{}) bool {
	_, ok := s.items[item]
	return ok
}

func (s *Set) Items() []interface{} {
	keys := make([]interface{}, 0)
	for k := range s.items {
		keys = append(keys, k)
	}
	return keys
}
