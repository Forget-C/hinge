package filter

import (
	"reflect"
	sysSort "sort"

	"github.com/Forget-C/hinge/index"
	"github.com/Forget-C/hinge/store"
)

type Builder interface {
	Page(page uint32) Builder
	Limit(limit uint32) Builder
	Sort(field string) Builder
	Where(index index.IndexName, s string) Builder
	WhereField(fields []string, word string, fuzzy bool) Builder
	Find(v interface{}) (uint32, error)
}

func NewDefaultFilter(store store.Store) Builder {
	return &filterBuilder{
		store:      store,
		conditions: map[index.IndexName][]string{},
	}
}

type filterBuilder struct {
	conditions   map[index.IndexName][]string
	sort         bool
	sortField    string
	store        store.Store
	page         uint32
	limit        uint32
	filterFields []string
	filterWord   string
	filterFuzzy  bool
}

func (f *filterBuilder) WhereField(fields []string, word string, fuzzy bool) Builder {
	f.filterFields = fields
	f.filterWord = word
	f.filterFuzzy = fuzzy
	return f
}

func (f *filterBuilder) stopIndex() int {
	return int(f.page * f.limit)
}

func (f *filterBuilder) startIndex() int {
	return int((f.page - 1) * f.limit)
}

func (f *filterBuilder) Limit(limit uint32) Builder {
	f.limit = limit
	return f
}

func (f *filterBuilder) Page(page uint32) Builder {
	if page == 0 {
		page = 1
	}
	f.page = page
	return f
}

func (f *filterBuilder) filterWithFields(ss string, index index.IndexName, fields []string, word string, fuzzy bool) ([]interface{}, error) {
	if f.store == nil {
		return nil, nil
	}
	var (
		items []interface{}
		err   error
	)
	if ss == "" || index == "" {
		items = f.store.List()
	} else {
		items, err = f.store.IndexFilter(ss, index)
		if err != nil {
			return nil, err
		}
	}
	return ResourceFilterByField(fields, word, fuzzy, items), nil
}

func (f *filterBuilder) list(v interface{}) (uint32, error) {
	items, err := f.filterWithFields("", "", f.filterFields, f.filterWord, f.filterFuzzy)
	if err != nil {
		return 0, err
	}
	total, err := f.decode(items, v, f.sort, f.sortField, f.startIndex(), f.stopIndex())
	return uint32(total), err
}

func (f *filterBuilder) Where(indexName index.IndexName, s string) Builder {
	if indexName == "" || s == "" {
		return f
	}
	f.conditions[indexName] = append(f.conditions[indexName], s)
	return f
}

func (f *filterBuilder) Sort(field string) Builder {
	f.sort = true
	f.sortField = field
	return f
}

func (f *filterBuilder) Find(v interface{}) (uint32, error) {
	if len(f.conditions) == 0 {
		return f.list(v)
	}
	var result []interface{}
	for index, conditions := range f.conditions {
		var lasts []interface{}
		for _, condition := range conditions {
			items, err := f.filterWithFields(condition, index, f.filterFields, f.filterWord, f.filterFuzzy)
			if err != nil {
				return 0, err
			}
			if len(items) == 0 {
				return 0, nil
			}
			if len(lasts) == 0 {
				lasts = items
				continue
			}
			var swap []interface{}
			for _, item := range items {
				for _, last := range lasts {
					if item == last {
						swap = append(swap, item)
					}
				}
			}
			if len(swap) == 0 {
				return 0, nil
			}
			lasts = swap
		}
		if len(result) == 0 {
			result = lasts
			continue
		}
		var swap []interface{}
		for _, last := range lasts {
			for _, r := range result {
				if last == r {
					swap = append(swap, last)
				}
			}
		}
		if len(swap) == 0 {
			return 0, nil
		}
		result = swap
	}
	total, err := f.decode(result, v, f.sort, f.sortField, f.startIndex(), f.stopIndex())
	return uint32(total), err
}

func (f *filterBuilder) decode(items sortItems, v interface{}, sort bool, sortField string, start, stop int) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}
	length := len(items)
	if sort && len(sortField) == 0 {
		sysSort.Sort(items)
	} else if sort {
		ResourceSortByField(sortField, items)
	}
	if stop <= 0 || stop > length {
		stop = length
	}
	if start < 0 || start >= length {
		start = 0
	}
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		return 0, ContainerStructureError
	}
	elem := value.Elem()
	if elem.Kind() != reflect.Slice {
		return 0, ContainerStructureError
	}
	itemsX := items[start:stop]
	for i := 0; i < len(itemsX); i++ {
		elem.Set(reflect.Append(elem, reflect.ValueOf(itemsX[i])))
	}
	return length, nil
}
