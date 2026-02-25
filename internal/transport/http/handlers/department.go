package handlers

import (
	"encoding/json"
	"errors"
	"github.com/iiincognito/org-structure/internal/service"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handlers struct {
	DeptService service.DepartmentService
	Logger      *slog.Logger
}

type createDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id,omitempty"`
}

type createEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at,omitempty"` // "2025-03-15"
}

type updateDepartmentRequest struct {
	Name     *string `json:"name,omitempty"`
	ParentID *uint   `json:"parent_id,omitempty"`
}

func (h *Handlers) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req createDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	dept, err := h.DeptService.CreateDepartment(r.Context(), req.Name, req.ParentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidName):
			h.respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrDuplicateNameInParent):
			h.respondError(w, http.StatusConflict, err.Error())
		default:
			h.respondError(w, http.StatusInternalServerError, "внутренняя ошибка")
		}
		return
	}

	h.respondJSON(w, http.StatusCreated, dept)
}

func (h *Handlers) GetDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/departments/")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "некорректный id подразделения")
		return
	}

	depthStr := r.URL.Query().Get("depth")
	depth := 1
	if depthStr != "" {
		d, _ := strconv.Atoi(depthStr)
		if d >= 1 && d <= 5 {
			depth = d
		}
	}

	includeEmployees := r.URL.Query().Get("include_employees") != "false"

	dept, err := h.DeptService.GetDepartment(r.Context(), uint(id), depth, includeEmployees)
	if err != nil {
		if errors.Is(err, service.ErrDepartmentNotFound) {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	h.respondJSON(w, http.StatusOK, dept)
}

func (h *Handlers) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из пути
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	idStr := strings.TrimSuffix(path, "/") // на случай лишнего слеша
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		h.respondError(w, http.StatusBadRequest, "некорректный или отсутствующий id подразделения")
		return
	}

	// Читаем тело запроса
	var req updateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	// Если ничего не передали — можно вернуть ошибку или просто 200 без изменений
	if req.Name == nil && req.ParentID == nil {
		h.respondError(w, http.StatusBadRequest, "не переданы поля для обновления (name или parent_id)")
		return
	}

	updatedDept, err := h.DeptService.UpdateDepartment(
		r.Context(),
		uint(id),
		req.Name,
		req.ParentID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDepartmentNotFound):
			h.respondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidName):
			h.respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrSelfAsParent),
			errors.Is(err, service.ErrCycleInHierarchy):
			h.respondError(w, http.StatusConflict, err.Error())
		default:
			h.Logger.Error("ошибка обновления подразделения", "id", id, "err", err)
			h.respondError(w, http.StatusInternalServerError, "внутренняя ошибка")
		}
		return
	}

	h.respondJSON(w, http.StatusOK, updatedDept)
}

func (h *Handlers) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	idStr := strings.TrimSuffix(path, "/")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		h.respondError(w, http.StatusBadRequest, "некорректный или отсутствующий id подразделения")
		return
	}

	// Читаем параметры запроса
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "cascade" // значение по умолчанию, если не указано
	}

	var reassignTo *uint
	if mode == "reassign" {
		reassignStr := r.URL.Query().Get("reassign_to_department_id")
		if reassignStr == "" {
			h.respondError(w, http.StatusBadRequest, "для mode=reassign обязателен параметр reassign_to_department_id")
			return
		}
		rID, err := strconv.ParseUint(reassignStr, 10, 32)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "некорректный reassign_to_department_id")
			return
		}
		reassignTo = new(uint)
		*reassignTo = uint(rID)
	}

	err = h.DeptService.DeleteDepartment(r.Context(), uint(id), mode, reassignTo)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDepartmentNotFound):
			h.respondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidDeleteMode):
			h.respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrInvalidReassignTarget):
			h.respondError(w, http.StatusBadRequest, err.Error())
		default:
			h.Logger.Error("ошибка удаления подразделения", "id", id, "mode", mode, "err", err)
			h.respondError(w, http.StatusInternalServerError, "внутренняя ошибка")
		}
		return
	}

	// 204 No Content — стандартный ответ для успешного DELETE
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) respondJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handlers) respondError(w http.ResponseWriter, code int, message string) {
	h.respondJSON(w, code, map[string]string{"error": message})
}

func (h *Handlers) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	parts := strings.SplitN(path, "/employees", 2)
	if len(parts) < 2 || parts[1] != "" {
		h.respondError(w, http.StatusBadRequest, "ожидается путь /departments/{id}/employees")
		return
	}

	idStr := parts[0]
	deptID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "некорректный id подразделения")
		return
	}

	var req createEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	emp, err := h.DeptService.CreateEmployee(
		r.Context(),
		uint(deptID),
		req.FullName,
		req.Position,
		req.HiredAt,
	)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "отдел с id"):
			h.respondError(w, http.StatusNotFound, err.Error())
		case strings.Contains(err.Error(), "некорректный формат hired_at"):
			h.respondError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondError(w, http.StatusInternalServerError, "внутренняя ошибка")
		}
		return
	}

	h.respondJSON(w, http.StatusCreated, emp)
}

func (h *Handlers) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		h.Logger.Info("HTTP запрос",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("duration", time.Since(start).String()),
		)
	})
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
