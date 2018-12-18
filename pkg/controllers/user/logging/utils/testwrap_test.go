package utils

import (
	"fmt"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	fmt.Println(RandStringBytes(10))
}
