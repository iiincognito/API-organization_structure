package main

import (
	"fmt"
	"github.com/iiincognito/org-structure/internal/config"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func main() {
	// Загружаем .env (если есть)
	_ = godotenv.Load()

	cfg := config.Load()

	// Подключаемся к базе
	dsn := cfg.DB_DSN
	if dsn == "" {
		log.Fatal("DB_DSN не задан в .env или переменных окружения")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе: %v", err)
	}

	// Простая проверка подключения
	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Ping к базе не прошёл: %v", err)
	}

	log.Println("Успешно подключились к PostgreSQL")

	// Минимальный HTTP-сервер
	mux := http.NewServeMux()

	// Простой health-check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Пока заглушка для будущего эндпоинта
	mux.HandleFunc("GET /api/v1/departments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"message": "Скоро здесь будет список подразделений"}`)
	})

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запускается на :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Сервер упал: %v", err)
	}
}
