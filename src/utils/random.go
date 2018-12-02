package util

import (
	"math/rand"
	"time"
)

// GenerateRandomNumber is make random number
func GenerateRandomNumber(start int, end int, count int) []int {
	// check range
	if end < start || (end-start) < count {
		return nil
	}
	// return slice
	nums := make([]int, 0)
	// make random number by time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		// make random number
		num := r.Intn((end - start)) + start
		// check repeat
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}
