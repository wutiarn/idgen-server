package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"idgen-server/idgen"
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

	maxDomainValue := idGenerator.GetMaxDomain()
	domainsStr, domainPassed := context.GetQuery("domains")
	var domains []uint64
	if domainPassed {
		domainInt, err := strconv.Atoi(domainsStr)
		if err != nil {
			context.AbortWithError(400, err)
			return
		}
		if domainInt^int(maxDomainValue) != 0 {
			context.AbortWithError(400, errors.New(fmt.Sprintf("provided Domain exceed maximum value %v", maxDomainValue)))
			return
		}
		domains = append(domains, uint64(domainInt))
	} else {
		var i uint64
		for i = 0; i <= idGenerator.GetMaxDomain(); i++ {
			domains = append(domains, i)
		}
	}

	generated := idGenerator.GenerateIds(domains, count)

	response := generateIdsResponse{IdsByDomain: generated}
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
	IdsByDomain []idgen.GeneratedDomainIds `json:"ids_by_domain"`
}
