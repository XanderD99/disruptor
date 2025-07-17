package util

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}

	if len(slice) < size {
		return [][]T{slice}
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := min(i+size, len(slice))
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
