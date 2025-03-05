package main

import (
	"distributed-calculator/internal/models"
	"distributed-calculator/internal/orchestrator"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func main() {
	operationTimes := make(map[models.Operation]int64)
	
	additionTime, err := strconv.ParseInt(getEnv("TIME_ADDITION_MS", "1000"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid TIME_ADDITION_MS: %v", err)
	}
	operationTimes[models.Addition] = additionTime
	
	subtractionTime, err := strconv.ParseInt(getEnv("TIME_SUBTRACTION_MS", "1000"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid TIME_SUBTRACTION_MS: %v", err)
	}
	operationTimes[models.Subtraction] = subtractionTime
	
	multiplicationTime, err := strconv.ParseInt(getEnv("TIME_MULTIPLICATIONS_MS", "2000"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid TIME_MULTIPLICATIONS_MS: %v", err)
	}
	operationTimes[models.Multiplication] = multiplicationTime
	
	divisionTime, err := strconv.ParseInt(getEnv("TIME_DIVISIONS_MS", "3000"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid TIME_DIVISIONS_MS: %v", err)
	}
	operationTimes[models.Division] = divisionTime
	
	repo := orchestrator.NewInMemoryRepository()
	
	service := orchestrator.NewService(repo, operationTimes)
	
	handlers := orchestrator.NewHandlers(service)
	
	router := mux.NewRouter()
	
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/calculate", handlers.CalculateHandler).Methods("POST")
	apiRouter.HandleFunc("/expressions", handlers.GetExpressionsHandler).Methods("GET")
	apiRouter.HandleFunc("/expressions/{id}", handlers.GetExpressionHandler).Methods("GET")
	
	internalRouter := router.PathPrefix("/internal").Subrouter()
	internalRouter.HandleFunc("/task", handlers.GetTaskHandler).Methods("GET")
	internalRouter.HandleFunc("/task", handlers.ProcessTaskResultHandler).Methods("POST")
	
	port := getEnv("PORT", "8080")
	log.Printf("Orchestrator starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}