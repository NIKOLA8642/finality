package orchestrator

import (
	"distributed-calculator/internal/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

func (h *Handlers) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	// Декодируем запрос
	var request models.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}
	
	if request.Expression == "" {
		http.Error(w, "Expression is required", http.StatusUnprocessableEntity)
		return
	}
	
	expression, err := h.service.ProcessExpression(request.Expression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.CalculateResponse{
		ID: expression.ID,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	expressions, err := h.service.GetAllExpressions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	response := models.ExpressionListResponse{
		Expressions: make([]models.Expression, 0, len(expressions)),
	}
	
	for _, expr := range expressions {
		response.Expressions = append(response.Expressions, *expr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	expression, err := h.service.GetExpressionByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	response := models.ExpressionResponse{
		Expression: *expression,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем задачу для обработки
	task, err := h.service.GetTaskForProcessing()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if task == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response := models.TaskResponse{
			Task: nil,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}
	
	response := models.TaskResponse{
		Task: task,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) ProcessTaskResultHandler(w http.ResponseWriter, r *http.Request) {
	var request models.TaskResultRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}
	
	if request.ID == "" {
		http.Error(w, "Task ID is required", http.StatusUnprocessableEntity)
		return
	}
	
	if err := h.service.ProcessTaskResult(request.ID, request.Result); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}