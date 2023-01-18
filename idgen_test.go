package main

import (
	"math"
	"sync"
	"testing"
	"time"
)

func TestGenerateId(t *testing.T) {
	params := IdParams{
		Timestamp: time.Unix(1673989769, 0),
		Counter:   1,
		ServerId:  5,
		Domain:    9,
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

	t.Run("Generate Timestamp overflow", func(t *testing.T) {
		modifiedParams := params
		overflowValue := int64(10)
		timestamp := int64(math.Pow(2, 35)) + epochStart + overflowValue
		modifiedParams.Timestamp = time.Unix(int64(timestamp), 0)
		generated := generateIdForParams(modifiedParams)
		var expected int64 = 2684372233
		if generated != expected {
			t.Errorf("generateIdForParams returned %v, expected %v", generated, expected)
		}

		parsedParams := parseIdToParams(generated)
		modifiedParams.Timestamp = time.Unix(int64(overflowValue+epochStart), 0)
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
	serverId := uint8(3)
	domainId := uint8(8)
	generator := NewIdGenerator(serverId)
	id := generator.GenerateSingleId(domainId)

	params := parseIdToParams(id)
	if params.Domain != domainId {
		t.Errorf("Generated id Domain is %v, expected %v", params.Domain, domainId)
	}
	if params.ServerId != serverId {
		t.Errorf("Generated id ServerId is %v, expected %v", params.ServerId, serverId)
	}
	if !params.Timestamp.Before(time.Now()) {
		t.Errorf("Generated id Timestamp %v is in future", params.Timestamp)
	}

	generator.Shutdown()
}

func TestIncrementCounter(t *testing.T) {
	startTime := time.Now()
	worker := domainWorker{
		ch:               make(chan idGenerationRequest),
		domain:           uint8(3),
		serverId:         5,
		currentTimestamp: startTime,
		counter:          0,
		wg:               &sync.WaitGroup{},
	}

	t.Run("First Counter increment", func(t *testing.T) {
		worker.incrementCounter()

		timeDelta := worker.currentTimestamp.Sub(startTime).Seconds()
		if timeDelta > 1 {
			t.Errorf("Incorrect Timestamp increment: %v", timeDelta)
		}

		if worker.counter != 1 {
			t.Errorf("Incorrect Counter increment: %v", worker.counter)
		}
	})

	t.Run("Sleep until next second", func(t *testing.T) {
		worker.currentTimestamp = time.Now()
		startTimestamp := worker.currentTimestamp
		worker.counter = maxCounterValue
		worker.incrementCounter()

		timeDelta := worker.currentTimestamp.Unix() - startTimestamp.Unix()
		if timeDelta < 1 {
			t.Errorf("Incorrect Timestamp value: %v", timeDelta)
		}

		if worker.counter != 0 {
			t.Errorf("Incorrect Counter value: %v", worker.counter)
		}
	})

	t.Run("Increment to max timedelta", func(t *testing.T) {
		now := time.Now().Unix()
		worker.currentTimestamp = time.Unix(now-(reservedSeconds*3), 0)
		startTimestamp := worker.currentTimestamp
		worker.incrementCounter()

		timeDelta := worker.currentTimestamp.Unix() - startTimestamp.Unix()
		if timeDelta < reservedSeconds*1.9 {
			t.Errorf("Incorrect time delta with start Timestamp: %v", timeDelta)
		}

		timeDelta = now - worker.currentTimestamp.Unix()
		if timeDelta > reservedSeconds || timeDelta < reservedSeconds*0.5 {
			t.Errorf("Incorrect time delta with now: %v", timeDelta)
		}

		if worker.counter != 0 {
			t.Errorf("Incorrect Counter value: %v", worker.counter)
		}
	})
}
