package http

import (
	"context"
	"github.com/iiincognito/org-structure/internal/service"
	"github.com/iiincognito/org-structure/internal/transport/http/handlers"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func NewServer(
	port string,
	deptService service.DepartmentService,
	logger *slog.Logger,
) *Server {
	mux := http.NewServeMux()

	// Регистрируем обработчики
	h := &handlers.Handlers{
		DeptService: deptService,
		Logger:      logger,
	}

	// Базовые маршруты
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /departments", h.CreateDepartment)
	mux.HandleFunc("GET /departments/", h.GetDepartment)       // /departments/{id}
	mux.HandleFunc("PATCH /departments/", h.UpdateDepartment)  // /departments/{id}
	mux.HandleFunc("DELETE /departments/", h.DeleteDepartment) // /departments/{id}

	// Сотрудники
	mux.HandleFunc("POST /departments/", h.CreateEmployee) // /departments/{id}/employees

	// Можно добавить другие маршруты позже

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h.LoggingMiddleware(mux),
	}

	return &Server{
		httpServer: srv,
		logger:     logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("HTTP сервер запущен", "addr", s.httpServer.Addr)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Ошибка HTTP-сервера", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Получен сигнал завершения, graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Ошибка graceful shutdown", "err", err)
		return err
	}

	s.logger.Info("HTTP сервер остановлен")
	return nil
}
