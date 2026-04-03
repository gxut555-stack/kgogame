package common

import "sort"

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	var sl = make([]K, 0, len(m))
	for k := range m {
		sl = append(sl, k)
	}
	return sl
}

func GetMapKeysWithSort[K Number, V any](m map[K]V, asc bool) []K {
	keys := GetMapKeys(m)
	sort.Slice(keys, func(i, j int) bool {
		if asc {
			return keys[i] < keys[j]
		}
		return keys[i] > keys[j]
	})
	return keys
}
