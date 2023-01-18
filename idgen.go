package main

import (
	"log"
	"math"
	"time"
)

const epochStart = int64(1672531200) // 2023-01-01 00:00:00
const timestampBits = 35
const counterBits = 14
const serverIdBits = 6
const domainBits = 8

type IdGenerator struct {
	domainWorkers []domainWorker
}

type domainWorker struct {
	ch       chan IdGenerationRequest
	domain   int8
	serverId int8
}

type IdGenerationRequest struct {
	count    int
	resultCh chan int64
}

func NewIdGenerator(serverId int8) *IdGenerator {
	generator := IdGenerator{}
	domainCount := int(math.Pow(2, domainBits))
	for i := 0; i < domainCount; i++ {
		println("Starting counter goroutine for domain", i)
		goroutine := domainWorker{
			ch:       make(chan IdGenerationRequest),
			domain:   int8(i),
			serverId: serverId,
		}
		go goroutine.start()
		generator.domainWorkers = append(generator.domainWorkers, goroutine)
	}
	return &generator
}

func (g *IdGenerator) GenerateSingleId(domain int8) int64 {
	return <-g.GenerateId(domain, 1)
}

func (g *IdGenerator) GenerateId(domain int8, count int) chan int64 {
	result := make(chan int64)
	request := IdGenerationRequest{
		count:    count,
		resultCh: result,
	}
	worker := g.domainWorkers[domain]
	worker.ch <- request
	return result
}

func (this *domainWorker) start() {
	timestamp := time.Now()
	for request := range this.ch {
		for i := 0; i < request.count; i++ {
			params := idParams{
				timestamp: timestamp,
				counter:   0,
				serverId:  this.serverId,
				domain:    this.domain,
			}
			id := generateIdForParams(params)
			request.resultCh <- id
		}
		close(request.resultCh)
		log.Printf("Generated %v ids in domain %v", request.count, this.domain)
	}
	log.Printf("Domain %v worker closed", this.domain)
}

func (g *IdGenerator) shutdown() {
	for _, worker := range g.domainWorkers {
		close(worker.ch)
	}
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
		counter:   int16(counter),
		serverId:  int8(serverId),
		domain:    int8(domain),
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

type idParams struct {
	timestamp time.Time
	counter   int16
	serverId  int8
	domain    int8
}
