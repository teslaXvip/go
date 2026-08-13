package algorithm

import "cmp"

func BinarySearch[T cmp.Ordered](arr []T, target T) int {
	begin := 0
	end := len(arr) - 1
	for begin <= end {
		middle := (begin + end) / 2
		if arr[middle] == target {
			return middle
		}

		if arr[middle] < target {
			begin = middle + 1
		} else {
			end = middle - 1
		}
	}
	return -1
}
