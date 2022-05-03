package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/zklevsha/go-musthave-devops/internal/dumper"
	"github.com/zklevsha/go-musthave-devops/internal/handlers"
)

const serverAddressDefault = "localhost:8080"
const storeIntervalDefault = time.Duration(300 * time.Second)
const storeFileDefault = "/tmp/devops-metrics-db.json"
const restoreDefault = true

var wg sync.WaitGroup

type serverConfig struct {
	serverAddress string
	storeInterval time.Duration
	storeFile     string
	restore       bool
}

func getServerConfig() serverConfig {
	c := serverConfig{
		serverAddress: serverAddressDefault,
		storeInterval: storeIntervalDefault,
		storeFile:     storeFileDefault,
		restore:       restoreDefault,
	}

	s := os.Getenv("ADDRESS")
	if s != "" {
		c.serverAddress = s
	}

	s = os.Getenv("STORE_INTERVAL")
	if s != "" {
		storeInterval, err := time.ParseDuration(s)
		if err != nil {
			log.Printf("WARN web failed to parse env var STORE_INTERVAL=%s: %s. Using default (%v)",
				s, err.Error(), c.storeInterval)
		}
		c.storeInterval = storeInterval
	}

	s = os.Getenv("STORE_FILE")
	if s != "" {
		c.storeFile = s
	}

	s = os.Getenv("RESTORE")
	if s != "" {
		restore, err := strconv.ParseBool(s)
		if err != nil {
			log.Printf("WARN web failed to parse env var RESTORE=%s: %s. Using default (%t)",
				s, err.Error(), c.restore)
		} else {
			c.restore = restore
		}
	}

	return c

}

func main() {
	config := getServerConfig()
	ctx, cancel := context.WithCancel(context.Background())

	// Init dumper
	wg.Add(1)
	go dumper.Start(ctx, &wg, config.storeInterval, config.storeFile, config.restore)

	// Starting web server
	handler := handlers.GetHandler()
	fmt.Printf("Starting web server at %s\n", config.serverAddress)

	srv := &http.Server{
		Addr:    config.serverAddress,
		Handler: handler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start web server: %s\n", err)
		}
	}()
	log.Print("Server Started\n")

	// Handling shutdown
	sig := <-done
	log.Printf("INFO main got a signal '%v', start shutting down...\n", sig)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	cancel()
	wg.Wait()
	log.Print("Server Exited Properly")

}
