package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zklevsha/go-musthave-devops/internal/handlers"
)

const serverSocket = ":8080"

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/update/{metricType}/{metricName}/{metricValue}",
		handlers.UpdateMeticHandler).Methods("POST")

	r.HandleFunc("/update/", handlers.UpdateMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	r.HandleFunc("/value/{metricType}/{metricName}",
		handlers.GetMetricHandler).Methods("GET")

	r.HandleFunc("/value/", handlers.GetMetricJSONHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	fmt.Printf("Starting server at %s\n", serverSocket)
	if err := http.ListenAndServe(serverSocket, r); err != nil {
		fmt.Println(err.Error())
	}
}
