package main

import (
	"distributed-calculator/internal/agent"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	orchestratorURL := getEnv("ORCHESTRATOR_URL", "http://localhost:8080")

	computingPower, err := strconv.Atoi(getEnv("COMPUTING_POWER", "4"))
	if err != nil {
		log.Fatalf("Invalid COMPUTING_POWER: %v", err)
	}
	service := agent.NewService(orchestratorURL)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	
	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go runWorker(service, i, stopChan, &wg)
	}
	
	<-stopChan
	log.Println("Shutting down agent...")
	
	wg.Wait()
	log.Println("Agent stopped")
}

func runWorker(service *agent.Service, id int, stopChan <-chan os.Signal, wg *sync.WaitGroup) {
	defer wg.Done()
	
	log.Printf("Worker %d started", id)
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-stopChan:
			log.Printf("Worker %d stopping", id)
			return
		case <-ticker.C:
			task, err := service.GetTask()
			if err != nil {
				log.Printf("Worker %d failed to get task: %v", id, err)
				time.Sleep(1 * time.Second) 
				continue
			}
			
			if task == nil {
				continue
			}
			
			log.Printf("Worker %d processing task %s: %s %s %s", id, task.ID, task.Arg1, task.Operation, task.Arg2)
			if err := service.ProcessTask(task); err != nil {
				log.Printf("Worker %d failed to process task %s: %v", id, task.ID, err)
				continue
			}
			
			log.Printf("Worker %d completed task %s", id, task.ID)
		}
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}