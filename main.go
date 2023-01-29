package main

import (
	"go.uber.org/zap"
	"idgen-server/idgen"
	"os"
)

var logger, _ = zap.NewDevelopment()
var config = GetConfig()
var idGenerator *idgen.IdGenerator

func main() {
	//goland:noinspection GoUnhandledErrorResult
	defer logger.Sync()
	var err error
	idGenerator, err = idgen.NewIdGenerator(config.IdGen, logger)
	if err != nil {
		logger.Error("Failed to initialize IdGenerator",
			zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Server started",
		zap.Uint64("nodeId", config.IdGen.InstanceId),
	)
	runServer()
}
