package algorithm

import "cmp"

func JaccardForSorted[T cmp.Ordered](collection1, collection2 []T) float64 {
	if len(collection1) == 0 || len(collection2) == 0 {
		return 0.0
	}
	s1 := 0
	s2 := 0
	len1 := len(collection1)
	len2 := len(collection2)
	intersection := 0
	for s1 < len1 && s2 < len2 {
		if collection1[s1] == collection2[s2] {
			intersection += 1
			s1 += 1
			s2 += 1
		} else if collection1[s1] < collection2[s2] {
			s1 += 1
		} else {
			s2 += 1
		}
	}

	return float64(intersection) / (float64(len1) + float64(len2) - float64(intersection))
}
