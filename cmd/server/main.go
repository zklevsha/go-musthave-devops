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
	"github.com/zklevsha/go-musthave-devops/internal/db"
	"github.com/zklevsha/go-musthave-devops/internal/dumper"
	"github.com/zklevsha/go-musthave-devops/internal/handlers"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

var wg sync.WaitGroup

func main() {
	log.Println("INFO main starting server")
	log.Printf("DEBUG startup flags: %v", os.Args)
	log.Printf("DEBUG ENVs: %v", os.Environ())

	config := config.GetServerConfig()

	logMsg := fmt.Sprintf("INFO main server config: ServerAddress: %s, UseDB: %t",
		config.ServerAddress, config.UseDB)
	if !config.UseDB {
		logMsg += fmt.Sprintf(", StoreInterval: %s, StoreFile: %s, Restore: %t",
			config.StoreInterval, config.StoreFile, config.Restore)
	}
	log.Println(logMsg)

	ctx, cancel := context.WithCancel(context.Background())
	var s structs.Storage
	if config.UseDB {
		s = &db.DBConnector{DSN: config.DSN, Ctx: ctx}
		err := s.Init()
		if err != nil {
			log.Panicf("failed to init connection to database: %s", err.Error())
		}
		defer s.Close()
	} else {
		s = structs.NewMemoryStorage()
		if config.Restore {
			dumper.RestoreData(config.StoreFile, s)
		}
		// Starting dumper
		wg.Add(1)
		go dumper.Start(ctx, &wg, config.StoreInterval, config.StoreFile, s)
	}

	// Starting web server
	handler := handlers.GetHandler(config, ctx, s)
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
