package main

import (
	"context"
	"os"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	if len(os.Args) != 2 {
		logger.Error("invalid number of arguments; provider only the path to the config file")
		os.Exit(1)
	}
	c, err := os.ReadFile(os.Args[1])
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	config, err := LoadConfig(c)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("starting beacon",
		zap.String("domain", config.Domain),
		zap.Int("TTL", config.TTL),
	)
	beacon, err := New(config, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	if err = beacon.Start(context.Background()); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
