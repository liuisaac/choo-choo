package parser

import (
	"errors"
	"fmt"
	"strings"
)

type Query struct {
	Op    string
	Key   string
	Value string
}

// ParseQuery parses a string query into a Query struct.
// It supports the following operations:
// - SET key value
// - GET key
// - DELETE key
// - INFO
// Modelled after SQL-like syntax, but simplified for key-value operations
func ParseQuery(q string) (Query, error) {
	parts := strings.Fields(q)
	if len(parts) == 0 {
		return Query{}, errors.New("empty query")
	}

	switch strings.ToUpper(parts[0]) {
		case "SET":
			if len(parts) != 3 {
				return Query{}, errors.New("SET requires key and value")
			}
			return Query{Op: "set", Key: parts[1], Value: parts[2]}, nil
		case "GET", "DELETE":
			if len(parts) != 2 {
				return Query{}, fmt.Errorf("%s requires a key", parts[0])
			}
			return Query{Op: strings.ToLower(parts[0]), Key: parts[1]}, nil
		case "INFO":
			return Query{Op: "info"}, nil
		default:
			return Query{}, fmt.Errorf("unknown op: %s", parts[0])
	}
}
