package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ducmeit1/gitlab-runner-clean/cmd"
	"github.com/ducmeit1/gitlab-runner-clean/logger"
)

func init() {
	logger.InitLogger(logger.InfoLevel)
}

func main() {
	ctx := context.Background()

	client, err := cmd.NewGitLabClient(ctx)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		client.DeleteAllOfflineRunners()
	}()

	total, err := client.GetAllOfflineRunners()
	if err != nil {
		panic(err)
	}

	logger.Infof("deleting total %v offline runners", total)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		<-c

		_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		// A interrupt signal has sent to us, let's shutdown server with gracefully
		logger.Infof("Stopping job...")
	}()

	wg.Done()
}
