package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"idgen-server/idgen"
	"math/rand"
	"strconv"
	"strings"
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
		domainStrs := strings.Split(domainsStr, ",")
		for _, domainStr := range domainStrs {
			domainInt, err := strconv.ParseUint(domainStr, 10, 64)
			if err != nil {
				context.AbortWithError(400, err)
				return
			}
			if domainInt > maxDomainValue {
				context.AbortWithError(400, errors.New(fmt.Sprintf("provided domain %v exceed maximum value %v", domainInt, maxDomainValue)))
				return
			}
			domains = append(domains, domainInt)
		}
	} else {
		domain := uint64(rand.Uint32()) & maxDomainValue
		domains = append(domains, domain)
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
