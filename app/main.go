package main

import (
	"git2.g4lab.com/devops/helper-tools/kubecheck/app/command"
	"go.uber.org/zap"
	"log"
	"os"
)

// application version
var VERSION = "dev"

func main() {
	zapLog, err := zap.NewProduction()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	sugar := zapLog.Sugar()

	app := command.NewKubeCheckCommand(VERSION, sugar)

	if err := app.Run(os.Args); err != nil {
		sugar.Fatal(err)
	}
}
