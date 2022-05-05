package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/zklevsha/go-musthave-devops/internal/config"
	"github.com/zklevsha/go-musthave-devops/internal/dumper"
	"github.com/zklevsha/go-musthave-devops/internal/handlers"
)

var wg sync.WaitGroup

func main() {
	log.Println("INFO main starting server")
	config := config.GetServerConfig()
	log.Printf("INFO main server config: ServerAddress: %s, StoreInterval: %s, StoreFile: %s, Restore: %t",
		config.ServerAddress, config.StoreInterval, config.StoreFile, config.Restore)
	ctx, cancel := context.WithCancel(context.Background())

	if config.Restore {
		dumper.RestoreData(config.StoreFile)
	}

	// Starting dumper
	wg.Add(1)
	go dumper.Start(ctx, &wg, config.StoreInterval, config.StoreFile)

	// Starting web server
	handler := handlers.GetHandler()
	fmt.Printf("INFO main starting web server at %s\n", config.ServerAddress)

	srv := &http.Server{
		Addr:    config.ServerAddress,
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
