# API Организационной структуры

REST API для управления организационной структурой компании: подразделения (с древовидной иерархией) и сотрудники.

**Основные возможности:**
- Создание, просмотр, обновление и удаление подразделений
- Поддержка дерева подразделений (parent_id)
- Создание сотрудников в подразделениях
- Получение подразделения с вложенными подразделениями (до глубины 5) и списком сотрудников
- Перемещение подразделения с проверкой на циклы
- Удаление подразделения в двух режимах: cascade (удалить всё поддерево) и reassign (перенести сотрудников в другой отдел)

**Технологии:**
- Go 1.23+
- PostgreSQL
- GORM + Goose (миграции)
- Docker + docker-compose
- Чистый net/http (без сторонних роутеров)

Архитектура: слои models → repository → service → transport/http

## Быстрый запуск в Docker 

1. Склонируйте репозиторий:

```bash
git clone https://github.com/iiincognito/org-structure.git
cd org-structure
cp .env.example .env
docker compose up --build
```
- поднимется контейнер PostgreSQL
- миграции (goose up)
- соберётся и запустится Go-приложение
- API будет доступно по адресу: http://localhost:8080

## Ручки для тестирования 
- POST /departments,Создать подразделение,"{""name"": ""Головной офис""}"
- POST /departments/{id}/employees,Создать сотрудника в подразделении,"{""full_name"": ""Иванов И.И."", ""position"": ""CEO""}"
- GET /departments/{id}?depth=2&include_employees=true,Получить подразделение + дерево + сотрудники,
- PATCH /departments/{id},Обновить название или родителя,"{""name"": ""Новое название""}"
- DELETE /departments/{id}?mode=cascade,Удалить подразделение каскадно,
- DELETE /departments/{id}?mode=reassign&reassign_to_department_id=5,Перенести сотрудников и удалить,