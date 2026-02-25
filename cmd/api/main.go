package main

import (
	"github.com/iiincognito/org-structure/internal/config"
	postgres2 "github.com/iiincognito/org-structure/internal/repository/postgres"
	"github.com/iiincognito/org-structure/internal/service"
	"github.com/iiincognito/org-structure/internal/transport/http"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log/slog"
	"os"
	"time"
)

func main() {
	// 1. Логирование (slog — стандарт с Go 1.21+)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Запуск приложения...")

	// 2. Загрузка конфигурации (.env)
	_ = godotenv.Load() // если .env есть — загрузится, если нет — ок
	cfg := config.Load()

	if cfg.DB_DSN == "" {
		slog.Error("DB_DSN не задан в .env или в переменных окружения")
		os.Exit(1)
	}

	// 3. Подключение к базе данных (GORM)
	db, err := gorm.Open(postgres.Open(cfg.DB_DSN), &gorm.Config{})
	if err != nil {
		slog.Error("Не удалось подключиться к PostgreSQL", "error", err)
		os.Exit(1)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		slog.Error("Ping к базе данных не удался", "error", err)
		os.Exit(1)
	}

	slog.Info("Успешно подключились к PostgreSQL")

	// 4. Инициализация репозитория
	repo := postgres2.NewPostgresRepository(db)

	// 5. Инициализация сервиса (бизнес-логика)
	deptService := service.NewDepartmentService(repo)

	// 6. Инициализация HTTP-сервера (transport/http)
	httpServer := http.NewServer(
		cfg.Port,
		deptService,
		logger,
	)

	// 7. Запуск сервера (с graceful shutdown внутри)
	slog.Info("Запускаем HTTP-сервер...")
	if err := httpServer.Start(); err != nil {
		slog.Error("HTTP-сервер завершился с ошибкой", "error", err)
		os.Exit(1)
	}
}
