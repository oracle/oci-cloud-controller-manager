package types

import (
	"strconv"
	"strings"
)

// GetOcpu returns the ocpu for the provided shape
func GetOcpu(shape string) int {
	parts := strings.Split(shape, ".")
	ocpu, _ := strconv.Atoi(parts[len(parts)-1])
	return ocpu
}
