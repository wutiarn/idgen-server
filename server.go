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

	var domain uint8
	domainStr, domainPassed := context.GetQuery("domain")
	if domainPassed {
		domainInt, err := strconv.Atoi(domainStr)
		if err != nil {
			context.AbortWithError(400, err)
			return
		}
		if domainInt&int(maxDomainValue) != 0 {
			context.AbortWithError(400, errors.New(fmt.Sprintf("Provided domain exceed maximum value %v", maxCounterValue)))
			return
		}
		domain = uint8(domainInt)
	} else {
		domain = uint8(rand.Uint32()) & maxDomainValue
	}

	idsCh := idGenerator.GenerateId(domain, count)
	ids := make([]int64, 0, count)
	for id := range idsCh {
		ids = append(ids, id)
	}
	response := generateIdsResponse{Ids: ids}
	context.JSON(200, response)
}

type generateIdsResponse struct {
	Ids []int64 `json:"ids"`
}
