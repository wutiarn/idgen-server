package main

import (
	"math"
	"time"
)

const epochStart = uint64(1672531200) // 2023-01-01 00:00:00
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
	var id int64 = 0
	id = encodePart(id, params.timestamp-epochStart, timestampBits)
	id = encodePart(id, params.counter, counterBits)
	id = encodePart(id, params.serverId, serverIdBits)
	id = encodePart(id, params.domain, domainBits)
	return id
}

func parseIdToParams(id int64) idParams {
	result := idParams{}

	result.domain, id = extractPart(id, domainBits)
	result.serverId, id = extractPart(id, serverIdBits)
	result.counter, id = extractPart(id, counterBits)
	result.timestamp, id = extractPart(id, timestampBits)

	result.timestamp = result.timestamp + epochStart

	return result
}

func encodePart(srcId int64, value uint64, bits int) int64 {
	mask := uint64(math.Pow(2, float64(bits))) - 1
	value = value & mask
	return srcId<<bits | int64(value)
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
