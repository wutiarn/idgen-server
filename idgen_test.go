package main

import (
	"testing"
)

func TestGenerateId(t *testing.T) {
	params := IdParams{
		timestamp: uint64(1673989769),
		counter:   1,
		serverId:  5,
		domain:    9,
	}

	id := generateIdForParams(params)
	var expected int64 = 449358206980867337
	if id != expected {
		t.Errorf("generateIdForParams returned %v, expected %v", id, expected)
	}

	parsedParams := parseIdToParams(id)
	if parsedParams != params {
		t.Errorf("parseIdToParams returned %v, expected %v", parsedParams, params)
	}
}
