package orchestrator

import (
	"distributed-calculator/internal/calculator"
	"distributed-calculator/internal/models"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

type Service struct {
	repo           Repository
	operationTimes map[models.Operation]int64
}

func NewService(repo Repository, operationTimes map[models.Operation]int64) *Service {
	return &Service{
		repo:           repo,
		operationTimes: operationTimes,
	}
}

func (s *Service) ProcessExpression(expr string) (*models.Expression, error) {
	expressionID := uuid.New().String()

	expression := &models.Expression{
		ID:         expressionID,
		Expression: expr,
		Status:     models.StatusProcessing,
	}

	if err := s.repo.SaveExpression(expression); err != nil {
		return nil, fmt.Errorf("failed to save expression: %w", err)
	}

	tasks, err := calculator.ParseExpression(expressionID, expr, s.operationTimes)
	if err != nil {
		expression.Status = models.StatusError
		expression.Error = err.Error()
		_ = s.repo.UpdateExpression(expression)
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	if len(tasks) == 0 {
		value, err := strconv.ParseFloat(expr, 64)
		if err != nil {
			expression.Status = models.StatusError
			expression.Error = "Invalid expression"
			_ = s.repo.UpdateExpression(expression)
			return nil, fmt.Errorf("invalid expression: %s", expr)
		}

		expression.Status = models.StatusCompleted
		expression.Result = &value
		_ = s.repo.UpdateExpression(expression)
		return expression, nil
	}

	for _, task := range tasks {
		if err := s.repo.SaveTask(task); err != nil {
			expression.Status = models.StatusError
			expression.Error = err.Error()
			_ = s.repo.UpdateExpression(expression)
			return nil, fmt.Errorf("failed to save task: %w", err)
		}
	}

	return expression, nil
}

func (s *Service) GetExpressionByID(id string) (*models.Expression, error) {
	return s.repo.GetExpressionByID(id)
}

func (s *Service) GetAllExpressions() ([]*models.Expression, error) {
	return s.repo.GetAllExpressions()
}

func (s *Service) GetTaskForProcessing() (*models.Task, error) {
	readyTasks, err := s.repo.GetReadyTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to get ready tasks: %w", err)
	}

	if len(readyTasks) == 0 {
		return nil, nil
	}

	return readyTasks[0], nil
}

func (s *Service) ProcessTaskResult(taskID string, result float64) error {
	task, err := s.repo.GetTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Completed = true
	task.Result = &result

	if err := s.repo.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}
