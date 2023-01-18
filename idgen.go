package main

import (
	"log"
	"math"
	"sync"
	"time"
)

const epochStart = int64(1672531200) // 2023-01-01 00:00:00
const timestampBits = 35
const counterBits = 14
const serverIdBits = 6
const domainBits = 8
const reservedSeconds = 10

var maxCounterValue = uint16(math.Pow(2, float64(counterBits)) - 1)

type IdGenerator struct {
	domainWorkers []domainWorker
	wg            *sync.WaitGroup
}

func NewIdGenerator(serverId uint8) *IdGenerator {
	wg := sync.WaitGroup{}
	generator := IdGenerator{
		wg: &wg,
	}
	domainCount := int(math.Pow(2, domainBits))
	for i := 0; i < domainCount; i++ {
		println("Starting counter goroutine for domain", i)
		goroutine := domainWorker{
			ch:               make(chan idGenerationRequest),
			domain:           uint8(i),
			serverId:         serverId,
			currentTimestamp: time.Now(),
			counter:          0,
			wg:               &wg,
		}
		go goroutine.start()
		generator.domainWorkers = append(generator.domainWorkers, goroutine)
	}
	return &generator
}

func (g *IdGenerator) GenerateId(domain uint8, count int) chan int64 {
	result := make(chan int64)
	request := idGenerationRequest{
		count:    count,
		resultCh: result,
	}
	worker := g.domainWorkers[domain]
	worker.ch <- request
	return result
}

func (g *IdGenerator) GenerateSingleId(domain uint8) int64 {
	return <-g.GenerateId(domain, 1)
}

func (g *IdGenerator) Shutdown() {
	for _, worker := range g.domainWorkers {
		close(worker.ch)
	}
	g.wg.Wait()
	log.Printf("All domains worker finished")
}

// ---------- Domain worker goroutines implementation ----------

type idGenerationRequest struct {
	count    int
	resultCh chan int64
}

type domainWorker struct {
	ch               chan idGenerationRequest
	domain           uint8
	serverId         uint8
	currentTimestamp time.Time
	counter          uint16
	wg               *sync.WaitGroup
}

func (w *domainWorker) start() {
	timestamp := time.Now()
	w.wg.Add(1)
	for request := range w.ch {
		for i := 0; i < request.count; i++ {
			w.incrementCounter()
			params := idParams{
				timestamp: timestamp,
				counter:   w.counter,
				serverId:  w.serverId,
				domain:    w.domain,
			}
			id := generateIdForParams(params)
			request.resultCh <- id
		}
		close(request.resultCh)
		log.Printf("Generated %v ids in domain %v", request.count, w.domain)
	}
	log.Printf("Domain %v worker finished", w.domain)
	w.wg.Done()
}

func (w *domainWorker) incrementCounter() {
	timeDelta := time.Now().Unix() - w.currentTimestamp.Unix()
	if timeDelta > reservedSeconds {
		w.currentTimestamp = time.Unix(time.Now().Unix()-reservedSeconds, 0)
		w.counter = 0
		return
	}

	if w.counter < maxCounterValue {
		w.counter++
		return
	}

	if timeDelta > 0 {
		w.currentTimestamp = w.currentTimestamp.Add(time.Second)
		w.counter = 0
		return
	}

	waitDuration := time.Until(w.currentTimestamp.Add(time.Second))
	log.Printf("Sleeping until next second (%v)", waitDuration)
	time.Sleep(waitDuration)
	w.currentTimestamp = w.currentTimestamp.Add(time.Second)
	w.counter = 0
}

// ---------- ID generation internals ----------

type idParams struct {
	timestamp time.Time
	counter   uint16
	serverId  uint8
	domain    uint8
}

func generateIdForParams(params idParams) int64 {
	var id int64 = 0
	timestamp := params.timestamp.Unix() - epochStart
	id = encodePart(id, timestamp, timestampBits)
	id = encodePart(id, int64(params.counter), counterBits)
	id = encodePart(id, int64(params.serverId), serverIdBits)
	id = encodePart(id, int64(params.domain), domainBits)
	return id
}

func parseIdToParams(id int64) idParams {
	domain, id := extractPart(id, domainBits)
	serverId, id := extractPart(id, serverIdBits)
	counter, id := extractPart(id, counterBits)
	timestamp, id := extractPart(id, timestampBits)

	return idParams{
		timestamp: time.Unix(int64(timestamp+epochStart), 0),
		counter:   uint16(counter),
		serverId:  uint8(serverId),
		domain:    uint8(domain),
	}
}

func encodePart(srcId int64, value int64, bits int) int64 {
	mask := int64(math.Pow(2, float64(bits))) - 1
	value = value & mask
	return srcId<<bits | value
}

func extractPart(id int64, bits int) (extracted int64, remainingId int64) {
	mask := int64(math.Pow(2, float64(bits))) - 1
	extracted = id & mask
	remainingId = id >> bits
	return
}
