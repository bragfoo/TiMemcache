package util

import (
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
)

// GetUUID is name uuid func
func GetUUID() string {
	tempID := fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))
	uuID := strings.Replace(tempID, "-", "", -1)
	return uuID
}
