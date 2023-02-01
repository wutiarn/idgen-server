package idgen

import (
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

var config = Config{
	InstanceId:           3,
	TimestampBits:        35,
	DomainBits:           8,
	CounterBits:          14,
	InstanceIdBits:       6,
	EpochStartSecond:     1672531200,
	ReservedSecondsCount: 60,
	StartupSecondOffset:  0,
}

var confWrapper, _ = newConfigWrapper(config)
var logger, _ = zap.NewDevelopment()

func TestGenerateId(t *testing.T) {
	params := IdParams{
		Timestamp:  time.Unix(1673989769, 0),
		Counter:    1,
		InstanceId: 5,
		Domain:     9,
	}

	var id uint64 = 0

	t.Run("Generate", func(t *testing.T) {
		id = generateIdForParams(params, confWrapper)
		var expected uint64 = 391531634640137
		if id != expected {
			t.Errorf("generateIdForParams returned %v, expected %v", id, expected)
		}
	})

	t.Run("Parse", func(t *testing.T) {
		parsedParams := parseIdToParams(id, confWrapper)
		if parsedParams != params {
			t.Errorf("parseIdToParams returned %v, expected %v", parsedParams, params)
		}
	})

	t.Run("Generate Timestamp overflow", func(t *testing.T) {
		modifiedParams := params
		overflowValue := uint64(10)
		timestamp := confWrapper.maxTimestamp + 1 + config.EpochStartSecond + overflowValue
		modifiedParams.Timestamp = time.Unix(int64(timestamp), 0)
		generated := generateIdForParams(modifiedParams, confWrapper)
		var expected uint64 = 2684372233
		if generated != expected {
			t.Errorf("generateIdForParams returned %v, expected %v", generated, expected)
		}

		parsedParams := parseIdToParams(generated, confWrapper)
		modifiedParams.Timestamp = time.Unix(int64(overflowValue+config.EpochStartSecond), 0)
		if parsedParams != modifiedParams {
			t.Errorf("parseIdToParams returned %v, expected %v", parsedParams, modifiedParams)
		}
	})
}

func TestTimestampLifetime(t *testing.T) {
	start := config.EpochStartSecond
	end := config.EpochStartSecond + confWrapper.maxTimestamp
	duration := end - start
	years := duration / 60 / 60 / 24 / 365
	yearsThreshold := uint64(1000)
	if years < yearsThreshold {
		t.Errorf("token lifespan is %v years, which is less than %v years threshold", years, yearsThreshold)
	}
}

func TestNewIdGenerator(t *testing.T) {
	domainId := uint64(8)
	generator, err := NewIdGenerator(config, logger)
	if err != nil {
		t.Errorf("Failed to initiate IdGenerator: %e", err)
		return
	}
	id := generator.GenerateSingleId(domainId)

	params := parseIdToParams(id, confWrapper)
	if params.Domain != domainId {
		t.Errorf("Generated id Domain is %v, expected %v", params.Domain, domainId)
	}
	if params.InstanceId != config.InstanceId {
		t.Errorf("Generated id ServerId is %v, expected %v", params.InstanceId, config.InstanceId)
	}
	if !params.Timestamp.Before(time.Now()) {
		t.Errorf("Generated id Timestamp %v is in future", params.Timestamp)
	}

	generator.Shutdown()
}

func TestIncrementCounter(t *testing.T) {
	startTime := time.Now()
	worker := domainWorker{
		logger:           logger,
		configWrapper:    confWrapper,
		ch:               make(chan idGenerationRequest),
		domain:           3,
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
		worker.counter = confWrapper.maxCounterValue
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
		now := uint64(time.Now().Unix())
		worker.currentTimestamp = time.Unix(int64(now-(config.ReservedSecondsCount*3)), 0)
		startTimestamp := worker.currentTimestamp
		worker.incrementCounter()

		timeDelta := worker.currentTimestamp.Unix() - startTimestamp.Unix()
		if timeDelta < int64(float64(config.ReservedSecondsCount)*1.9) {
			t.Errorf("Incorrect time delta with start Timestamp: %v", timeDelta)
		}

		timeDelta = int64(now - uint64(worker.currentTimestamp.Unix()))
		if timeDelta > int64(config.ReservedSecondsCount) || timeDelta < int64(float64(config.ReservedSecondsCount)*0.5) {
			t.Errorf("Incorrect time delta with now: %v", timeDelta)
		}

		if worker.counter != 0 {
			t.Errorf("Incorrect Counter value: %v", worker.counter)
		}
	})
}

func TestStartupSecondOffset(t *testing.T) {
	var cfg = config
	cfg.StartupSecondOffset = -1000
	cfg.ReservedSecondsCount = 2000
	generator, err := NewIdGenerator(cfg, logger)
	if err != nil {
		t.Errorf("Failed to initializa IdGenerator: %e", err)
	}
	worker := generator.domainWorkers[0]
	worker.incrementCounter()
	timestamp := worker.currentTimestamp
	delta := time.Now().Sub(timestamp)
	if delta < time.Second*999 || delta > time.Second*2000 {
		t.Errorf("Unexpected timestamp delta: %v", delta.Seconds())
	}
}
