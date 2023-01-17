package main

import (
	"math"
	"testing"
)

func TestGenerateId(t *testing.T) {
	params := idParams{
		timestamp: uint64(1673989769),
		counter:   1,
		serverId:  5,
		domain:    9,
	}

	var id int64 = 0

	t.Run("Generate", func(t *testing.T) {
		id = generateIdForParams(params)
		var expected int64 = 449358206980867337
		if id != expected {
			t.Errorf("generateIdForParams returned %v, expected %v", id, expected)
		}
	})

	t.Run("Parse", func(t *testing.T) {
		parsedParams := parseIdToParams(id)
		if parsedParams != params {
			t.Errorf("parseIdToParams returned %v, expected %v", parsedParams, params)
		}
	})

	t.Run("Generate timestamp overflow", func(t *testing.T) {
		modifiedParams := params
		modifiedParams.timestamp = 34359738371
		generated := generateIdForParams(modifiedParams)
		var expected int64 = 805324041
		if generated != expected {
			t.Errorf("generateIdForParams returned %v, expected %v", id, expected)
		}

		parsedParams := parseIdToParams(generated)
		modifiedParams.timestamp = modifiedParams.timestamp - uint64(math.Pow(2, 35))
		if parsedParams != modifiedParams {
			t.Errorf("parseIdToParams returned %v, expected %v", parsedParams, modifiedParams)
		}
	})
}
