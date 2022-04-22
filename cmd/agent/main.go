package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zklevsha/go-musthave-devops/agent/poller"
	"github.com/zklevsha/go-musthave-devops/agent/reporter"
)

const pollInterval = time.Duration(2 * time.Second)
const reportInterval = time.Duration(10 * time.Second)
const serverSocket = "127.0.0.1:8080"

var wg sync.WaitGroup

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	wg.Add(2)
	go poller.Poll(ctx, &wg, pollInterval)
	time.Sleep(time.Second)
	go reporter.Report(ctx, &wg, serverSocket, reportInterval)
	sig := <-c
	log.Printf("INFO main got a signal '%v', start shutting down...\n", sig)
	cancel()
	wg.Wait()
	log.Printf("INFO main Shutdown complete")
}
