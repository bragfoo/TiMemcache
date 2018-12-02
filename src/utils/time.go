package util

import (
	"strconv"
	"time"
)

// TimeUnix is time for unix
func TimeUnix() string {
	naiveTime := time.Now().Unix()
	naiveTimeString := strconv.FormatInt(naiveTime, 10)
	return naiveTimeString
}
