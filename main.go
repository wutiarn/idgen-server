package main

import "go.uber.org/zap"

var logger, _ = zap.NewDevelopment()
var idGenerator = NewIdGenerator(8)

func main() {
	//goland:noinspection GoUnhandledErrorResult
	defer logger.Sync()
	var config = GetConfig()
	logger.Info("Server started",
		zap.Uint8("nodeId", config.NodeId),
	)
	runServer()
}
