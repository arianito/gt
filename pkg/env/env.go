package env

import (
"os"
"strconv"
"strings"
)

func String(key string, def ...string) string {
	op := os.Getenv(key)
	if op == "" {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
	return op
}
func Array(key string, def ...string) []string {
	op := os.Getenv(key)
	if op == "" {
		return def
	}
	return strings.Split(op, ",")
}


func Int64(key string, def ...int64) int64 {
	op := os.Getenv(key)
	value, err := strconv.ParseInt(op, 10,64)
	if err != nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	return value
}

func Float(key string, def ...float64) float64 {
	op := os.Getenv(key)
	value, err := strconv.ParseFloat(op,64)
	if err != nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	return value
}

