package calculator

import (
	"distributed-calculator/internal/models"
	"testing"
)

func TestParseExpression(t *testing.T) {
	operationTimes := map[models.Operation]int64{
		models.Addition:       1000,
		models.Subtraction:    1000,
		models.Multiplication: 2000,
		models.Division:       3000,
	}

	tasks, err := ParseExpression("expr1", "2 + 3", operationTimes)
	if err != nil {
		t.Errorf("Failed to parse simple expression: %v", err)
		return
	}
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
		return
	}
	if tasks[0].Arg1 != "2" || tasks[0].Arg2 != "3" || tasks[0].Operation != models.Addition {
		t.Errorf("Incorrect task for simple expression: %+v", tasks[0])
	}

	tasks, err = ParseExpression("expr2", "2 + 3 * 4", operationTimes)
	if err != nil {
		t.Errorf("Failed to parse expression with precedence: %v", err)
		return
	}
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
		return
	}
	multiplicationTask := findTaskByOperation(tasks, models.Multiplication)
	additionTask := findTaskByOperation(tasks, models.Addition)
	if multiplicationTask == nil || additionTask == nil {
		t.Errorf("Missing expected tasks in the result")
		return
	}
	if len(additionTask.Dependencies) != 1 || additionTask.Dependencies[0] != multiplicationTask.ID {
		t.Errorf("Incorrect dependencies in the tasks")
	}

	tasks, err = ParseExpression("expr3", "(2 + 3) * 4", operationTimes)
	if err != nil {
		t.Errorf("Failed to parse expression with brackets: %v", err)
		return
	}
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
		return
	}
	additionTask = findTaskByOperation(tasks, models.Addition)
	multiplicationTask = findTaskByOperation(tasks, models.Multiplication)
	if multiplicationTask == nil || additionTask == nil {
		t.Errorf("Missing expected tasks in the result")
		return
	}
	if len(multiplicationTask.Dependencies) != 1 || multiplicationTask.Dependencies[0] != additionTask.ID {
		t.Errorf("Incorrect dependencies in the tasks")
	}
}

func findTaskByOperation(tasks []*models.Task, operation models.Operation) *models.Task {
	for _, task := range tasks {
		if task.Operation == operation {
			return task
		}
	}
	return nil
}