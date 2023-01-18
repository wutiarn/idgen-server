package main

import "go.uber.org/zap"

var logger, _ = zap.NewDevelopment()
var config = GetConfig()
var idGenerator = NewIdGenerator(config.NodeId)

func main() {
	//goland:noinspection GoUnhandledErrorResult
	defer logger.Sync()
	logger.Info("Server started",
		zap.Uint8("nodeId", config.NodeId),
	)
	runServer()
}
