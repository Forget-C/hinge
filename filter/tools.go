package filter

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

func ResourceSortByField(field string, items []interface{}) []interface{} {
	quicksort(strings.Split(field, "."), items, 0, len(items)-1)
	return items
}

func quicksort(field []string, items []interface{}, left, right int) {
	var i, j int
	if left > right {
		return
	}

	sentryValue := find(field, reflect.ValueOf(items[left]))
	sentryItem := items[left]
	i = left
	j = right

	for i < j {
		for find(field, reflect.ValueOf(items[j])) >= sentryValue && i < j {
			j--
		}
		items[i] = items[j]
		for find(field, reflect.ValueOf(items[i])) <= sentryValue && i < j {
			i++
		}
		items[j] = items[i]
	}
	items[i] = sentryItem

	quicksort(field, items, left, i-1)
	quicksort(field, items, i+1, right)
}

// ResourceFilterByField
// @Description:
// @param fields 字段 Spec.Containers.Name, Spec.Containers.Image
// @param value
// @param fuzzy 是否模糊
// @param items
// @return []interface{}
func ResourceFilterByField(fields []string, value string, fuzzy bool, items []interface{}) []interface{} {
	if len(fields) == 0 || len(value) == 0 {
		return items
	}
	var res []interface{}
	for _, item := range items {
		if ResourceFieldMatched(fields, value, fuzzy, item) {
			res = append(res, item)
		}
	}
	return res
}

func ResourceFieldMatched(fields []string, value string, fuzzy bool, item interface{}) bool {
	rv := reflect.ValueOf(item)
	for _, field := range fields {
		fieldS := strings.Split(field, ".")
		v := find(fieldS, rv)
		if fuzzy {
			if strings.Contains(v, value) {
				return true
			}
		} else {
			if v == value {
				return true
			}
		}
	}
	return false
}

// find
// @Description:
// @param field Spec, Containers, Name
// @param t
// @return string
func find(field []string, t reflect.Value) string {
	if len(field) == 0 {
		return ""
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		structField := t.FieldByName(field[0])
		if !structField.IsValid() {
			return ""
		}
		if len(field) == 1 {
			switch structField.Kind() {
			case reflect.Int, reflect.Int32, reflect.Uint32, reflect.Int64:
				return cast.ToString(structField.Int())
			case reflect.String:
				return structField.String()
			case reflect.Map:
				iter := structField.MapRange()
				v := ""
				for iter.Next() {
					v = v + fmt.Sprintf("%s=%s", iter.Key().String(), iter.Value().String())
				}
				return v
			default:
				return ""
			}
		}
		if len(field) >= 2 {
			return find(field[1:], structField)
		}
	case reflect.Slice:
		for i := 0; i < t.Len(); i++ {
			item := t.Index(i)
			if item.IsValid() {
				v := find(field, item)
				if v != "" {
					return v
				}
			}
		}
	}
	return ""
}

var ContainerStructureError = errors.New("container structure must be an array of ptr")

type sortItems []interface{}

func (s sortItems) Less(i, j int) bool {
	x := reflect.ValueOf(s[i])
	y := reflect.ValueOf(s[j])
	if x.Kind() == reflect.Ptr {
		x = x.Elem()
	}
	if y.Kind() == reflect.Ptr {
		y = y.Elem()
	}
	xNameField := x.FieldByName("Name")
	if !xNameField.IsValid() {
		return false
	}
	yNameField := y.FieldByName("Name")
	if !yNameField.IsValid() {
		return false
	}

	return yNameField.String() < xNameField.String()
}
func (s sortItems) Len() int {
	return len(s)
}
func (s sortItems) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
