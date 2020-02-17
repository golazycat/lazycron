package common

import (
	"fmt"
	"time"
)

func IntSecond(t int) time.Duration {
	return time.Duration(t) * time.Second
}

func GetHost(addr string, port int) string {
	return fmt.Sprintf("%s:%d", addr, port)
}
