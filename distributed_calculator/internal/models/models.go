package models

type ExpressionStatus string

const (
	StatusPending   ExpressionStatus = "PENDING"
	StatusProcessing ExpressionStatus = "PROCESSING"
	StatusCompleted ExpressionStatus = "COMPLETED"
	StatusError     ExpressionStatus = "ERROR"
)

type Operation string

const (
	Addition       Operation = "+"
	Subtraction    Operation = "-"
	Multiplication Operation = "*"
	Division       Operation = "/"
)

type Expression struct {
	ID         string           `json:"id"`
	Expression string           `json:"expression,omitempty"`
	Status     ExpressionStatus `json:"status"`
	Result     *float64         `json:"result,omitempty"`
	Error      string           `json:"error,omitempty"`
}

type Task struct {
	ID            string    `json:"id"`
	ExpressionID  string    `json:"-"`
	Arg1          string    `json:"arg1"`
	Arg2          string    `json:"arg2"`
	Operation     Operation `json:"operation"`
	OperationTime int64     `json:"operation_time"`
	Result        *float64  `json:"result,omitempty"`
	Completed     bool      `json:"-"`
	Dependencies  []string  `json:"-"`
}

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	ID string `json:"id"`
}

type ExpressionListResponse struct {
	Expressions []Expression `json:"expressions"`
}

type ExpressionResponse struct {
	Expression Expression `json:"expression"`
}

type TaskResponse struct {
	Task *Task `json:"task,omitempty"`
}

type TaskResultRequest struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}