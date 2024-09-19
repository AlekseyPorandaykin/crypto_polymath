package util

func BatchSlice[T interface{}](data []T, count int) [][]T {
	batch := make([][]T, 0, 100)
	tempBatch := make([]T, 0, count)
	for _, item := range data {
		if len(tempBatch) >= count {
			batch = append(batch, tempBatch)
			tempBatch = make([]T, 0, count)
		}
		tempBatch = append(tempBatch, item)
	}
	batch = append(batch, tempBatch)

	return batch
}

func ClearSlice[T interface{}](data []T, fn func(item T) bool) []T {
	result := make([]T, 0, len(data))
	for _, item := range data {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}
func ModifySlice[T interface{}, K interface{}](data []T, fn func(T) K) []K {
	result := make([]K, 0, len(data))
	for _, item := range data {
		result = append(result, fn(item))
	}
	return result
}

func UniqSlice[T comparable](data []T) []T {
	uniqKeys := make(map[T]struct{})
	for _, item := range data {
		if _, has := uniqKeys[item]; has {
			continue
		}
		uniqKeys[item] = struct{}{}
	}
	result := make([]T, 0, len(uniqKeys))
	for key := range uniqKeys {
		result = append(result, key)
	}
	return result
}
