package main

import (
	"math"
	"time"
)

const timestampBits = 35
const counterBits = 14
const serverIdBits = 6
const domainBits = 8

func GenerateId(domain uint8) int64 {
	params := idParams{
		timestamp: uint64(time.Now().Unix()),
		counter:   1,
		serverId:  5,
		domain:    uint64(domain),
	}
	return generateIdForParams(params)
}

func generateIdForParams(params idParams) int64 {
	id := params.timestamp
	id = id<<14 | uint64(params.counter)
	id = id<<6 | uint64(params.serverId)
	id = id<<8 | uint64(params.domain)
	return int64(id)
}

func parseIdToParams(id int64) idParams {
	result := idParams{}

	result.domain, id = extractPart(id, 8)
	result.serverId, id = extractPart(id, 6)
	result.counter, id = extractPart(id, 14)
	result.timestamp, id = extractPart(id, 35)

	return result
}

func extractPart(id int64, bits int) (extracted uint64, remainingId int64) {
	mask := int64(math.Pow(2, float64(bits))) - 1
	extracted = uint64(id & mask)
	remainingId = id >> bits
	return
}

type idParams struct {
	timestamp uint64
	counter   uint64
	serverId  uint64
	domain    uint64
}
