package idgen

import (
	"go.uber.org/zap"
	"math"
	"sync"
	"time"
)

type IdGenerator struct {
	logger        *zap.Logger
	configWrapper configWrapper
	domainWorkers []domainWorker
	wg            *sync.WaitGroup
}

func NewIdGenerator(config Config, logger *zap.Logger) (*IdGenerator, error) {
	configWrapper, err := newConfigWrapper(config)
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	domainCount := configWrapper.maxDomain + 1
	startTimestamp := time.Now().Add(time.Second * time.Duration(config.StartupSecondOffset))
	generator := IdGenerator{
		logger:        logger,
		configWrapper: configWrapper,
		domainWorkers: make([]domainWorker, 0, domainCount),
		wg:            &wg,
	}
	var i uint64
	for i = 0; i < domainCount; i++ {
		goroutine := domainWorker{
			logger:           logger,
			configWrapper:    configWrapper,
			ch:               make(chan idGenerationRequest, 100),
			domain:           i,
			currentTimestamp: startTimestamp,
			counter:          0,
			wg:               &wg,
		}
		go goroutine.start()
		generator.domainWorkers = append(generator.domainWorkers, goroutine)
	}
	logger.Info("Id generator initialized",
		zap.Uint64("domainCount", domainCount))
	return &generator, nil
}

func (g *IdGenerator) GenerateIdsForDomain(domain uint64, count int) chan uint64 {
	result := make(chan uint64, count)
	request := idGenerationRequest{
		count:    count,
		resultCh: result,
	}
	worker := g.domainWorkers[domain]
	worker.ch <- request
	return result
}

func (g *IdGenerator) GenerateIds(domains []uint64, count int) []GeneratedDomainIds {
	channels := make([]chan uint64, 0, len(domains))
	for _, domain := range domains {
		ch := g.GenerateIdsForDomain(domain, count)
		channels = append(channels, ch)
	}
	result := make([]GeneratedDomainIds, 0, len(domains))
	for i, channel := range channels {
		domain := domains[i]
		domainResult := make([]uint64, 0, count)
		for id := range channel {
			domainResult = append(domainResult, id)
		}
		result = append(result, GeneratedDomainIds{
			Domain: domain,
			Ids:    domainResult,
		})
	}
	return result
}

func (g *IdGenerator) GenerateSingleId(domain uint64) uint64 {
	return <-g.GenerateIdsForDomain(domain, 1)
}

func (g *IdGenerator) Shutdown() {
	for _, worker := range g.domainWorkers {
		close(worker.ch)
	}
	g.wg.Wait()
	g.logger.Info("All domains worker finished")
}

func (g *IdGenerator) GetMaxDomain() uint64 {
	return g.configWrapper.maxDomain
}

func (g *IdGenerator) ParseIdToParams(id uint64) IdParams {
	return parseIdToParams(id, g.configWrapper)
}

// ---------- Domain worker goroutines implementation ----------

type idGenerationRequest struct {
	count    int
	resultCh chan uint64
}

type domainWorker struct {
	logger           *zap.Logger
	configWrapper    configWrapper
	ch               chan idGenerationRequest
	domain           uint64
	currentTimestamp time.Time
	counter          uint64
	wg               *sync.WaitGroup
}

func (w *domainWorker) start() {
	w.wg.Add(1)
	for request := range w.ch {
		w.updateTimestamp()
		for i := 0; i < request.count; i++ {
			w.incrementCounter()
			params := IdParams{
				Timestamp:  w.currentTimestamp,
				Counter:    w.counter,
				InstanceId: w.configWrapper.config.InstanceId,
				Domain:     w.domain,
			}
			id := generateIdForParams(params, w.configWrapper)
			request.resultCh <- id
		}
		close(request.resultCh)
	}
	w.wg.Done()
}

func (w *domainWorker) updateTimestamp() {
	timeDelta := time.Now().Unix() - w.currentTimestamp.Unix()
	reservedSecondsCount := int64(w.configWrapper.config.ReservedSecondsCount)
	if timeDelta > reservedSecondsCount {
		w.currentTimestamp = time.Unix(time.Now().Unix()-reservedSecondsCount, 0)
		w.counter = 0
		w.logger.Debug("Using earliest reserve second",
			zap.Time("timestamp", w.currentTimestamp),
			zap.Uint64("domain", w.domain))
		return
	}

	if w.counter < w.configWrapper.maxCounterValue {
		return
	}

	if timeDelta > 0 {
		w.currentTimestamp = w.currentTimestamp.Add(time.Second)
		w.counter = 0

		w.logger.Debug("Using next reserve second",
			zap.Int64("left", timeDelta-1),
			zap.Time("timestamp", w.currentTimestamp),
			zap.Uint64("domain", w.domain))
		return
	}

	waitDuration := time.Until(w.currentTimestamp.Add(time.Second))
	w.logger.Info("Sleeping until next second",
		zap.Duration("duration", waitDuration),
		zap.Uint64("Domain", w.domain))
	time.Sleep(waitDuration)
	w.currentTimestamp = w.currentTimestamp.Add(time.Second)
	w.counter = 0
	w.logger.Debug("Using realtime second",
		zap.Time("timestamp", w.currentTimestamp),
		zap.Uint64("domain", w.domain))
}

func (w *domainWorker) incrementCounter() {
	if w.counter >= w.configWrapper.maxCounterValue {
		w.updateTimestamp()
		return
	}

	w.counter++
}

// ---------- ID generation internals ----------

type IdParams struct {
	Timestamp  time.Time `json:"timestamp"`
	Counter    uint64    `json:"counter"`
	InstanceId uint64    `json:"InstanceId"`
	Domain     uint64    `json:"domain"`
}

func generateIdForParams(params IdParams, configWrapper configWrapper) uint64 {
	var id uint64 = 0
	timestamp := uint64(params.Timestamp.Unix() - int64(configWrapper.config.EpochStartSecond))
	id = encodePart(id, timestamp, configWrapper.config.TimestampBits)
	id = encodePart(id, params.Counter, configWrapper.config.CounterBits)
	id = encodePart(id, params.InstanceId, configWrapper.config.InstanceIdBits)
	id = encodePart(id, params.Domain, configWrapper.config.DomainBits)
	return id
}

func parseIdToParams(id uint64, configWrapper configWrapper) IdParams {
	domain, id := extractPart(id, configWrapper.config.DomainBits)
	instanceId, id := extractPart(id, configWrapper.config.InstanceIdBits)
	counter, id := extractPart(id, configWrapper.config.CounterBits)
	timestamp, id := extractPart(id, configWrapper.config.TimestampBits)

	return IdParams{
		Timestamp:  time.Unix(int64(timestamp+configWrapper.config.EpochStartSecond), 0),
		Counter:    counter,
		InstanceId: instanceId,
		Domain:     domain,
	}
}

func encodePart(srcId uint64, value uint64, bits uint8) uint64 {
	mask := uint64(math.Pow(2, float64(bits))) - 1
	value = value & mask
	return srcId<<bits | value
}

func extractPart(id uint64, bits uint8) (extracted uint64, remainingId uint64) {
	mask := uint64(math.Pow(2, float64(bits))) - 1
	extracted = id & mask
	remainingId = id >> bits
	return
}

type GeneratedDomainIds struct {
	Domain uint64   `json:"domain"`
	Ids    []uint64 `json:"ids"`
}
