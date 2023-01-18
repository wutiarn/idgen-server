package main

import (
	"math"
	"testing"
	"time"
)

func TestGenerateId(t *testing.T) {
	params := idParams{
		timestamp: time.Unix(1673989769, 0),
		counter:   1,
		serverId:  5,
		domain:    9,
	}

	var id int64 = 0

	t.Run("Generate", func(t *testing.T) {
		id = generateIdForParams(params)
		var expected int64 = 391531634640137
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
		overflowValue := int64(10)
		timestamp := int64(math.Pow(2, 35)) + epochStart + overflowValue
		modifiedParams.timestamp = time.Unix(int64(timestamp), 0)
		generated := generateIdForParams(modifiedParams)
		var expected int64 = 2684372233
		if generated != expected {
			t.Errorf("generateIdForParams returned %v, expected %v", generated, expected)
		}

		parsedParams := parseIdToParams(generated)
		modifiedParams.timestamp = time.Unix(int64(overflowValue+epochStart), 0)
		if parsedParams != modifiedParams {
			t.Errorf("parseIdToParams returned %v, expected %v", parsedParams, modifiedParams)
		}
	})
}

func TestTimestampLifetime(t *testing.T) {
	start := epochStart
	end := epochStart + int64(math.Pow(2, timestampBits)-1)
	duration := end - start
	years := duration / 60 / 60 / 24 / 365
	yearsThreshold := int64(1000)
	if years < yearsThreshold {
		t.Errorf("token lifespan is %v years, which is less than %v years threshold", years, yearsThreshold)
	}
}

func TestNewIdGenerator(t *testing.T) {
	serverId := int8(3)
	domainId := int8(8)
	generator := NewIdGenerator(serverId)
	id := generator.GenerateSingleId(domainId)

	params := parseIdToParams(id)
	if params.domain != domainId {
		t.Errorf("Generated id domain is %v, expected %v", params.domain, domainId)
	}
	if params.serverId != serverId {
		t.Errorf("Generated id serverId is %v, expected %v", params.serverId, serverId)
	}
	if !params.timestamp.Before(time.Now()) {
		t.Errorf("Generated id timestamp %v is in future", params.timestamp)
	}

	generator.shutdown()
}
