package calculator

import (
	"distributed-calculator/internal/models"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type ASTNode struct {
	NodeType    string
	Value       string
	Left        *ASTNode
	Right       *ASTNode
	TaskID      string
	Dependencies []string
}

func ParseExpression(expressionID, expression string, operationTimes map[models.Operation]int64) ([]*models.Task, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	
	ast, err := buildAST(expression)
	if err != nil {
		return nil, err
	}
	
	tasks := []*models.Task{}
	_, err = createTasksFromAST(ast, &tasks, expressionID, operationTimes)
	if err != nil {
		return nil, err
	}
	
	return tasks, nil
}

func buildAST(expression string) (*ASTNode, error) {
	// Если выражение пустое, возвращаем ошибку
	if len(expression) == 0 {
		return nil, fmt.Errorf("пустое выражение")
	}
	
	bracketCount := 0
	for i := len(expression) - 1; i >= 0; i-- {
		char := expression[i]
		if char == ')' {
			bracketCount++
		} else if char == '(' {
			bracketCount--
		}
		
		if bracketCount == 0 && (char == '+' || char == '-') {
			return &ASTNode{
				NodeType: "operation",
				Value:    string(char),
				Left:     mustBuildAST(expression[:i]),
				Right:    mustBuildAST(expression[i+1:]),
			}, nil
		}
	}
	
	bracketCount = 0
	for i := len(expression) - 1; i >= 0; i-- {
		char := expression[i]
		if char == ')' {
			bracketCount++
		} else if char == '(' {
			bracketCount--
		}
		
		if bracketCount == 0 && (char == '*' || char == '/') {
			return &ASTNode{
				NodeType: "operation",
				Value:    string(char),
				Left:     mustBuildAST(expression[:i]),
				Right:    mustBuildAST(expression[i+1:]),
			}, nil
		}
	}
	
	if expression[0] == '(' && expression[len(expression)-1] == ')' {
		bracketCount = 0
		for i := 0; i < len(expression)-1; i++ {
			if expression[i] == '(' {
				bracketCount++
			} else if expression[i] == ')' {
				bracketCount--
			}
			
			if bracketCount == 0 {
				break
			}
		}
		
		if bracketCount != 0 {
			return buildAST(expression[1 : len(expression)-1])
		}
	}
	
	_, err := strconv.ParseFloat(expression, 64)
	if err == nil {
		return &ASTNode{
			NodeType: "number",
			Value:    expression,
		}, nil
	}
	
	return nil, fmt.Errorf("недопустимое выражение: %s", expression)
}

func mustBuildAST(expression string) *ASTNode {
	node, err := buildAST(expression)
	if err != nil {
		panic(err)
	}
	return node
}

func createTasksFromAST(node *ASTNode, tasks *[]*models.Task, expressionID string, operationTimes map[models.Operation]int64) (string, error) {
	if node.NodeType == "number" {
		return node.Value, nil
	}
	
	if node.NodeType == "operation" {
		leftArg, err := createTasksFromAST(node.Left, tasks, expressionID, operationTimes)
		if err != nil {
			return "", err
		}
		
		rightArg, err := createTasksFromAST(node.Right, tasks, expressionID, operationTimes)
		if err != nil {
			return "", err
		}
		
		taskID := uuid.New().String()
		operation := models.Operation(node.Value)
		
		dependencies := []string{}
		if node.Left.TaskID != "" {
			dependencies = append(dependencies, node.Left.TaskID)
		}
		if node.Right.TaskID != "" {
			dependencies = append(dependencies, node.Right.TaskID)
		}
		
		task := &models.Task{
			ID:            taskID,
			ExpressionID:  expressionID,
			Arg1:          leftArg,
			Arg2:          rightArg,
			Operation:     operation,
			OperationTime: operationTimes[operation],
			Completed:     false,
			Dependencies:  dependencies,
		}
		
		*tasks = append(*tasks, task)
		
		node.TaskID = taskID
		
		return taskID, nil
	}
	
	return "", fmt.Errorf("неизвестный тип узла: %s", node.NodeType)
}