package util

import (
	"regexp"
)

func removeRepByLoop(slc []string) []string {
	// mark data
	result := []string{}
	for i := range slc {
		// find \s
		reg := regexp.MustCompile(`[^s]+`)
		r := reg.FindAllString(slc[i], -1)
		if len(r) != 0 {
			flag := true
			for j := range result {
				if slc[i] == result[j] {
					// mark duplicate data and mark false
					flag = false
					break
				}
			}
			// false not to data
			if flag {
				result = append(result, slc[i])
			}
		}
	}
	return result
}

func removeRepByMap(slc []string) []string {
	result := []string{}
	// mark data
	tempMap := map[string]byte{}
	for _, e := range slc {
		// find \s
		reg := regexp.MustCompile(`[^s]+`)
		r := reg.FindAllString(e, -1)
		if len(r) != 0 {
			l := len(tempMap)
			tempMap[e] = 0
			// len metamorphic is not duplicate
			if len(tempMap) != l {
				result = append(result, e)
			}
		}
	}
	return result
}

// RemoveDuplicatesAndEmpty is delete duplicate data from slice.
func RemoveDuplicatesAndEmpty(slc []string) []string {
	// < 1024 use slice
	if len(slc) < 1024 {
		return removeRepByLoop(slc)
	}
	// < 1024 use map
	return removeRepByMap(slc)
}
