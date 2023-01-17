package main

import (
	"math"
	"time"
)

const epochStart = int64(1672531200) // 2023-01-01 00:00:00
const timestampBits = 35
const counterBits = 14
const serverIdBits = 6
const domainBits = 8

func GenerateId(domain int8) int64 {
	params := idParams{
		timestamp: time.Now(),
		counter:   1,
		serverId:  5,
		domain:    domain,
	}
	return generateIdForParams(params)
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
