package util

import "strconv"

func SliceMapIntToString(elems []int) []string {
	strElems := make([]string, len(elems))
	for i := range elems {
		strElems[i] = strconv.Itoa(elems[i])
	}
	return strElems
}
