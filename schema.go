package notorm

import (
	"reflect"
)

type fieldKey struct {
	t reflect.Type
	p uintptr
}

type Schema[T any] struct {
	C     *T
	cache map[fieldKey]string
}

func SchemaOf[T any]() Schema[T] {
	var v T
	s := Schema[T]{C: &v}
	s.initCache()
	return s
}

func SchemaFor[T any](value *T) Schema[T] {
	s := Schema[T]{C: value}
	s.initCache()
	return s
}

func (s *Schema[T]) initCache() {
	s.cache = make(map[fieldKey]string)

	type scanEntry struct {
		v      reflect.Value
		prefix string
	}

	fieldPointers := []scanEntry{{v: reflect.ValueOf(s.C)}}
	for len(fieldPointers) > 0 {
		next := fieldPointers[len(fieldPointers)-1]
		fieldPointers = fieldPointers[:len(fieldPointers)-1]

		strct := next.v.Elem()
		for i := 0; i < strct.NumField(); i++ {
			fv := strct.Field(i)
			f := strct.Type().Field(i)
			if !fv.CanAddr() {
				continue
			}
			name := f.Name
			if alias, ok := f.Tag.Lookup("column"); ok {
				name = alias
			}
			s.cache[fieldKey{
				t: f.Type,
				p: fv.Addr().Pointer(),
			}] = next.prefix + name

			prefix := ""
			if !f.Anonymous {
				if pref, ok := f.Tag.Lookup("prefix"); ok {
					prefix = pref
				} else {
					prefix = name + "."
				}
			}
			if fv.Kind() == reflect.Pointer && !fv.IsNil() && fv.Elem().Kind() == reflect.Struct {
				fieldPointers = append(fieldPointers, scanEntry{v: fv, prefix: next.prefix + prefix})
			} else if fv.Kind() == reflect.Struct {
				fieldPointers = append(fieldPointers, scanEntry{v: fv.Addr(), prefix: next.prefix + prefix})
			}
		}
	}
}

func (s *Schema[T]) NameOf(column any) string {
	if column == nil {
		panic("argument must reference a valid column of schema")
	}
	if reflect.TypeOf(column).Kind() != reflect.Pointer {
		panic("argument must be a pointer to a column of schema")
	}
	if s.cache == nil {
		s.initCache()
	}
	r := reflect.ValueOf(column)
	key := fieldKey{
		t: r.Elem().Type(),
		p: r.Pointer(),
	}
	return s.cache[key]
}

func (s *Schema[T]) Names() []string {
	if s.cache == nil {
		s.initCache()
	}
	names := make([]string, 0, len(s.cache))
	for _, v := range s.cache {
		names = append(names, v)
	}
	return names
}
