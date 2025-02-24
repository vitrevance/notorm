package notorm

import (
	"fmt"
	"reflect"
)

type Scanner interface {
	Scan(...any) error
}

type ColumnAwareScanner interface {
	Scanner
	Columns() ([]string, error)
}

func dereferenceStruct(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}
	return t
}

// Matcher must return a positive integer - a match priority, or 0 if no match
func FindFieldIndexByMatcher(matcher func(reflect.StructField) int, structType reflect.Type) (result []int) {
	if dereferenceStruct(structType).Kind() != reflect.Struct {
		return nil
	}
	found := false
	type iterData struct {
		t     reflect.Type
		index []int
	}
	current := []iterData{}
	next := []iterData{{
		t:     dereferenceStruct(structType),
		index: nil,
	}}
	var nextCount map[reflect.Type]int
	visited := map[reflect.Type]bool{}

	bestMatchPriority := 0
	bestMatchCount := 0
	var bestMatchT reflect.Type

	for len(next) > 0 {
		current, next = next, current[:0]
		count := nextCount
		nextCount = nil

		for _, scan := range current {
			t := scan.t
			if visited[t] {
				continue
			}
			visited[t] = true
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if !f.IsExported() {
					continue
				}
				priority := matcher(f)
				if priority > 0 && priority >= bestMatchPriority {
					if priority == bestMatchPriority {
						bestMatchCount++
					} else {
						bestMatchPriority = priority
						bestMatchCount = 1
						result = append(scan.index, i)
					}
					bestMatchT = t
					found = true
				}

				var ntyp reflect.Type = dereferenceStruct(f.Type)

				if found || ntyp == nil || ntyp.Kind() != reflect.Struct {
					continue
				}

				if nextCount[ntyp] > 0 {
					nextCount[ntyp] = 2
					continue
				}
				if nextCount == nil {
					nextCount = map[reflect.Type]int{}
				}
				nextCount[ntyp] = 1
				if count[t] > 1 {
					nextCount[ntyp] = 2
				}
				var index []int
				index = append(index, scan.index...)
				index = append(index, i)
				next = append(next, iterData{ntyp, index})
			}
		}
		if found {
			if bestMatchCount > 1 || count[bestMatchT] > 1 {
				return nil
			}
			break
		}
	}
	return
}

func findFieldIndexForColumn(col string, structType reflect.Type) (result []int) {
	matcher := func(f reflect.StructField) int {
		if tag, ok := f.Tag.Lookup("column"); ok {
			if tag == col {
				return 2
			}
		}
		if col == f.Name {
			return 1
		}
		return 0
	}
	return FindFieldIndexByMatcher(matcher, structType)
}

func ScanPrepared(target any, cols []string, s Scanner) error {
	if reflect.ValueOf(target).IsNil() || reflect.TypeOf(target).Kind() != reflect.Pointer {
		panic("invalid pointer top target")
	}
	pointers := make([]any, len(cols))

	for i, col := range cols {
		f := findFieldIndexForColumn(col, reflect.TypeOf(target))
		if f == nil {
			return fmt.Errorf("missing field for column %s", cols[i])
		}
		val, err := reflect.ValueOf(target).Elem().FieldByIndexErr(f)
		if err != nil {
			return fmt.Errorf("cannot scan through nil pointer %s: %w", cols[i], err)
		}
		pointers[i] = val.Addr().Interface()
	}
	return s.Scan(pointers...)
}

func Scan(target any, s ColumnAwareScanner) error {
	cols, err := s.Columns()
	if err != nil {
		return err
	}
	return ScanPrepared(target, cols, s)
}
