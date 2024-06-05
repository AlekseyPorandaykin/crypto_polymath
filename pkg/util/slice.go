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
