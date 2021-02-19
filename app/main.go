package main

import (
	"log"
	"os"

	"github.com/lanDeleih/kubecheck/app/command"
	"go.uber.org/zap"
)

var VERSION = "dev"

func main() {
	logger := newLogger()
	app := command.NewKubeCheckCommand(VERSION, logger)

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}

func newLogger() *zap.SugaredLogger {
	zapLog, err := zap.NewProduction()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	return zapLog.Sugar()
}
