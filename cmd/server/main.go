package main

import (
	"fmt"
	"net/http"
	"os"

	muxHandler "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/handlers"
)

const serverSocket = ":8080"

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/update/{metricType}/{metricID}/{metricValue}",
		handlers.UpdateMeticHandler).Methods("POST")

	r.HandleFunc("/update/", handlers.UpdateMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	r.HandleFunc("/value/{metricType}/{metricID}",
		handlers.GetMetricHandler).Methods("GET")

	r.HandleFunc("/value/", handlers.GetMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	fmt.Printf("Starting server at %s\n", serverSocket)

	loggedRouter := muxHandler.LoggingHandler(os.Stdout, r)
	if err := http.ListenAndServe(serverSocket, loggedRouter); err != nil {
		fmt.Println(err.Error())
	}
}
