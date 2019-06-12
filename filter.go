package main

import (
	"strings"
)

type filter map[string]string
type filters []filter

func parseFilters(limits []string) (filters, error) {
	// TODO error handler
	filters := []filter{}

	// AND in same string, OR in other string
	for _, str := range limits {
		f := filter{}
		arr := strings.Split(str, ",")
		for _, s := range arr {
			a := strings.SplitN(s, "=", 2)
			f[a[0]] = a[1]
		}
		filters = append(filters, f)
	}
	return filters, nil
}

func (fs filters) isOk(item map[string]string) bool {
	if len(fs) == 0 {
		return true
	}
	for _, f := range fs {
		if f.isOk(item) {
			return true
		}
	}
	return false
}
func (f filter) isOk(item map[string]string) bool {
	for key, value := range f {
		val, ok := item[key]
		if !ok {
			return false
		}

		if val != value {
			return false
		}
	}
	return true
}
func (fs filters) filter(items map[string]Item) []Item {
	res := []Item{}
	for name, item := range items {
		item["name"] = name
		if !fs.isOk(item) {
			continue
		}
		res = append(res, item)
	}

	return res
}

type Filter struct {
	Key   string
	Value string
	Op    string
}

func (f *Filter) isOk(item Item) bool {
	switch f.Op {
	case "%":
		strings.Contains(item[f.Key], f.Value)
	case "?":
		_, ok := item[f.Key]
		return ok
	case "!=":
		return item[f.Key] != f.Value
	default:
		return item[f.Key] == f.Value
	}
	return false
}

type Filters struct {
	OR []struct {
		AND []Filter
	}
}

func (fs *Filters) filter(items map[string]Item) map[string]Item {
	result := map[string]Item{}
	for name, item := range items {
		item["name"] = name
	NextFilter:
		for _, f := range fs.OR {
			for _, ff := range f.AND {
				if !ff.isOk(item) {
					continue NextFilter
				}
			}
			result[name] = item
			break
		}
	}
	return result
}
