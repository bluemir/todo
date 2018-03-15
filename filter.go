package main

import "strings"

type filter map[string]string

func (f filter) isOk(labels map[string]string) bool {
	for key, value := range f {
		val, ok := labels[key]
		if !ok {
			return false
		}

		if val != value {
			return false
		}
	}
	return true
}
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

type filters []filter

func (fs filters) isOk(labels map[string]string) bool {
	if len(fs) == 0 {
		return true
	}
	for _, f := range fs {
		if f.isOk(labels) {
			return true
		}
	}
	return false
}
