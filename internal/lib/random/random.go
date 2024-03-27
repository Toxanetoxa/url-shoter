package random

import (
	"golang.org/x/exp/rand"
	"strings"
	"time"
)

func NewRandomString(length int64) string {
	rand.Seed(uint64(time.Now().UnixNano()))
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

	var b strings.Builder

	for i := 0; length > int64(i); i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}

	return b.String()
}
