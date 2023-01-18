package main

import "go.uber.org/zap"

var logger, _ = zap.NewDevelopment()

func main() {
	//goland:noinspection GoUnhandledErrorResult
	defer logger.Sync()
	var config = GetConfig()
	logger.Info("Server started",
		zap.Uint8("nodeId", config.NodeId),
	)
	println("Hello world", config.NodeId)
}
