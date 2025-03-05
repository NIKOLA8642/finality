package orchestrator

import (
	"distributed-calculator/internal/models"
	"fmt"
	"sync"
)

type Repository interface {
	SaveExpression(expression *models.Expression) error
	UpdateExpression(expression *models.Expression) error
	GetExpressionByID(id string) (*models.Expression, error)
	GetAllExpressions() ([]*models.Expression, error)
	SaveTask(task *models.Task) error
	UpdateTask(task *models.Task) error
	GetTaskByID(id string) (*models.Task, error)
	GetReadyTasks() ([]*models.Task, error)
}

type InMemoryRepository struct {
	expressions     map[string]*models.Expression
	tasks           map[string]*models.Task
	tasksByExprID   map[string][]*models.Task
	expressionMutex sync.RWMutex
	taskMutex       sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		expressions:   make(map[string]*models.Expression),
		tasks:         make(map[string]*models.Task),
		tasksByExprID: make(map[string][]*models.Task),
	}
}

func (r *InMemoryRepository) SaveExpression(expression *models.Expression) error {
	r.expressionMutex.Lock()
	defer r.expressionMutex.Unlock()
	
	r.expressions[expression.ID] = expression
	return nil
}

func (r *InMemoryRepository) UpdateExpression(expression *models.Expression) error {
	r.expressionMutex.Lock()
	defer r.expressionMutex.Unlock()
	
	if _, exists := r.expressions[expression.ID]; !exists {
		return fmt.Errorf("expression with ID %s not found", expression.ID)
	}
	
	r.expressions[expression.ID] = expression
	return nil
}

func (r *InMemoryRepository) GetExpressionByID(id string) (*models.Expression, error) {
	r.expressionMutex.RLock()
	defer r.expressionMutex.RUnlock()
	
	expression, exists := r.expressions[id]
	if !exists {
		return nil, fmt.Errorf("expression with ID %s not found", id)
	}
	
	return expression, nil
}

func (r *InMemoryRepository) GetAllExpressions() ([]*models.Expression, error) {
	r.expressionMutex.RLock()
	defer r.expressionMutex.RUnlock()
	
	expressions := make([]*models.Expression, 0, len(r.expressions))
	for _, expression := range r.expressions {
		expressions = append(expressions, expression)
	}
	
	return expressions, nil
}

func (r *InMemoryRepository) SaveTask(task *models.Task) error {
	r.taskMutex.Lock()
	defer r.taskMutex.Unlock()
	
	r.tasks[task.ID] = task
	
	r.tasksByExprID[task.ExpressionID] = append(r.tasksByExprID[task.ExpressionID], task)
	
	return nil
}

func (r *InMemoryRepository) UpdateTask(task *models.Task) error {
	r.taskMutex.Lock()
	defer r.taskMutex.Unlock()
	
	if _, exists := r.tasks[task.ID]; !exists {
		return fmt.Errorf("task with ID %s not found", task.ID)
	}
	
	r.tasks[task.ID] = task
	
	r.checkExpressionCompletion(task.ExpressionID)
	
	return nil
}

func (r *InMemoryRepository) GetTaskByID(id string) (*models.Task, error) {
	r.taskMutex.RLock()
	defer r.taskMutex.RUnlock()
	
	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", id)
	}
	
	return task, nil
}

func (r *InMemoryRepository) GetReadyTasks() ([]*models.Task, error) {
	r.taskMutex.RLock()
	defer r.taskMutex.RUnlock()
	
	readyTasks := []*models.Task{}
	
	for _, task := range r.tasks {
		if task.Completed {
			continue
		}

		allDependenciesCompleted := true
		for _, depID := range task.Dependencies {
			depTask, exists := r.tasks[depID]
			if !exists || !depTask.Completed {
				allDependenciesCompleted = false
				break
			}
		}
		
		if allDependenciesCompleted {
			taskCopy := *task
			
			// Если аргумент - это ID задачи, заменяем его на результат
			if taskArg1, exists := r.tasks[task.Arg1]; exists && taskArg1.Completed && taskArg1.Result != nil {
				taskCopy.Arg1 = fmt.Sprintf("%f", *taskArg1.Result)
			}
			
			if taskArg2, exists := r.tasks[task.Arg2]; exists && taskArg2.Completed && taskArg2.Result != nil {
				taskCopy.Arg2 = fmt.Sprintf("%f", *taskArg2.Result)
			}
			
			readyTasks = append(readyTasks, &taskCopy)
		}
	}
	
	return readyTasks, nil
}


func (r *InMemoryRepository) checkExpressionCompletion(expressionID string) {
	tasks, exists := r.tasksByExprID[expressionID]
	if !exists {
		return
	}
	
	var lastTask *models.Task
	for _, task := range tasks {
		isReferenced := false
		for _, otherTask := range tasks {
			for _, depID := range otherTask.Dependencies {
				if depID == task.ID {
					isReferenced = true
					break
				}
			}
			if isReferenced {
				break
			}
		}
		
		if !isReferenced {
			lastTask = task
			break
		}
	}
	
	if lastTask != nil && lastTask.Completed && lastTask.Result != nil {
		expression, exists := r.expressions[expressionID]
		if exists {
			expression.Status = models.StatusCompleted
			expression.Result = lastTask.Result
		}
	}
}