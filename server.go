package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"strconv"
)

func runServer() {
	r := gin.Default()
	r.GET("/generate", handleGenerateRequest)
	r.GET("/parse", parseId)
	err := r.Run()
	if err != nil {
		panic(err)
	}
}

//goland:noinspection GoUnhandledErrorResult
func handleGenerateRequest(context *gin.Context) {
	countStr := context.DefaultQuery("count", "1")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		context.AbortWithError(400, err)
		return
	}

	var domain uint64
	maxDomainValue := idGenerator.GetMaxDomain()
	domainStr, domainPassed := context.GetQuery("domain")
	if domainPassed {
		domainInt, err := strconv.Atoi(domainStr)
		if err != nil {
			context.AbortWithError(400, err)
			return
		}
		if domainInt^int(maxDomainValue) != 0 {
			context.AbortWithError(400, errors.New(fmt.Sprintf("provided Domain exceed maximum value %v", maxDomainValue)))
			return
		}
		domain = uint64(domainInt)
	} else {
		domain = uint64(rand.Uint32()) & maxDomainValue
	}

	idsCh := idGenerator.GenerateId(domain, count)
	ids := make([]uint64, 0, count)
	for id := range idsCh {
		ids = append(ids, id)
	}
	response := generateIdsResponse{Ids: ids}
	context.JSON(200, response)
}

//goland:noinspection GoUnhandledErrorResult
func parseId(context *gin.Context) {
	idStr, idPassed := context.GetQuery("id")
	if !idPassed {
		context.AbortWithError(400, errors.New("required param id was not provided"))
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		context.AbortWithError(400, errors.New("failed to parse provided id to int64"))
		return
	}
	params := idGenerator.ParseIdToParams(id)
	context.JSON(200, params)
}

type generateIdsResponse struct {
	Ids []uint64 `json:"ids"`
}
