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
	params := IdParams{
		timestamp: uint64(time.Now().Unix()),
		counter:   1,
		serverId:  5,
		domain:    domain,
	}
	return generateIdForParams(params)
}

func generateIdForParams(params IdParams) int64 {
	id := params.timestamp
	id = id<<14 | uint64(params.counter)
	id = id<<6 | uint64(params.serverId)
	id = id<<8 | uint64(params.domain)
	return int64(id)
}

func parseIdToParams(id int64) IdParams {
	result := IdParams{}

	result.domain = uint8(id & int64(math.Pow(2, 8)-1))
	id = id >> 8

	result.serverId = uint8(id & int64(math.Pow(2, 6)-1))
	id = id >> 6

	result.counter = uint16(id & int64(math.Pow(2, 14)-1))
	id = id >> 14

	result.timestamp = uint64(id & int64(math.Pow(2, 35)-1))

	return result
}

func extractPart(id int64, bits int) (extracted int64, remainingId int64) {
	mask := int64(math.Pow(2, float64(bits))) - 1
	extracted = id & mask
	remainingId = id >> bits
}

type IdParams struct {
	timestamp uint64
	counter   uint16
	serverId  uint8
	domain    uint8
}
