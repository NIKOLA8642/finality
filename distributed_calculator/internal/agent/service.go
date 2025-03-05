package agent

import (
	"bytes"
	"distributed-calculator/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Service struct {
	orchestratorURL string
	client          *http.Client
}

func NewService(orchestratorURL string) *Service {
	return &Service{
		orchestratorURL: orchestratorURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Service) GetTask() (*models.Task, error) {
	// Отправляем GET-запрос к оркестратору
	resp, err := s.client.Get(fmt.Sprintf("%s/internal/task", s.orchestratorURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response models.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Task, nil
}

func (s *Service) ProcessTask(task *models.Task) error {
	// Преобразуем аргументы в числа
	arg1, err := strconv.ParseFloat(task.Arg1, 64)
	if err != nil {
		return fmt.Errorf("invalid arg1: %w", err)
	}

	arg2, err := strconv.ParseFloat(task.Arg2, 64)
	if err != nil {
		return fmt.Errorf("invalid arg2: %w", err)
	}

	var result float64
	switch task.Operation {
	case models.Addition:
		result = arg1 + arg2
	case models.Subtraction:
		result = arg1 - arg2
	case models.Multiplication:
		result = arg1 * arg2
	case models.Division:
		if arg2 == 0 {
			return fmt.Errorf("division by zero")
		}
		result = arg1 / arg2
	default:
		return fmt.Errorf("unknown operation: %s", task.Operation)
	}

	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	return s.SendTaskResult(task.ID, result)
}

func (s *Service) SendTaskResult(taskID string, result float64) error {
	request := models.TaskResultRequest{
		ID:     taskID,
		Result: result,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Post(
		fmt.Sprintf("%s/internal/task", s.orchestratorURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to send task result: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}