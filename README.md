# Library API System

Современная система учета книг для библиотеки, разработанная на Go. Поддерживает управление книгами, регистрацию читателей и контроль выдачи книг.

## Стек технологий
- **Язык:** Go 1.21+
- **БД:** SQLite 
- **Аутентификация:** JWT (RS256)
- **Логирование:** slog (JSON format)
- **Безопасность:** bcrypt для хеширования паролей

## Как запустить

1. **Клонируйте репозиторий:**
   ```bash
   git clone <https://github.com/spothyjohn-tech/Library/tree/main>
   cd library-api
   ```

2. **Настройте переменные окружения:**
   Создайте файл `.env` на основе `.env.example`.

3. **Запустите сервер:**
   ```
    go run cmd/main.go
   ```
   *При первом запуске база данных создастся автоматически, и будет добавлен системный администратор:*
   - **Email:** `admin@admin.com`
   - **Password:** `admin123`

### Список эндпоинтов и права доступа:

- [Публично] POST /login — Вход и получение токена
- [Публично] GET /books — Просмотр всех книг
- [Публично] GET /books/{id} — получить информацию о конкретной книге.

- [Только АДМИН] POST /books — Добавить книгу
- [Только АДМИН] PUT /books/{id} — Изменить книгу
- [Только АДМИН] DELETE /books/{id} — Удалить книгу
- [Только АДМИН] POST /users — Регистрация читателя
- [Только АДМИН] POST /issues — Выдача книги (+14 дней)
- [Только АДМИН] POST /returns — Возврат книги

- [ЮЗЕР / АДМИН] GET /users/{id}/books — Посмотреть свои книги

## Примеры запросов (curl)

### 1. Авторизация (Логин)
```bash
curl -X POST http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"email": "admin@admin.com", "password": "admin123"}'
```

### 2. Добавление книги (Только для Админа)
```bash
curl -X POST http://localhost:8080/books \
     -H "Authorization: Bearer <ВАШ_ТОКЕН>" \
     -H "Content-Type: application/json" \
     -d '{"title": "The Go Programming Language", "author": "Alan Donovan", "isbn": "9785389016828", "year": 2015}'
```

### 3. Получение списка книг (Публично)
```bash
curl -X GET "http://localhost:8080/books?author=Alan&status=Available"
```

### 4. Регистрация нового читателя (Только для Админа)
```bash
curl -X POST http://localhost:8080/users \
     -H "Authorization: Bearer <ВАШ_ТОКЕН>" \
     -d '{"name": "Ivan Ivanov", "email": "ivan@example.com", "password": "securepassword"}'
```

### 5. Выдача книги
```bash
curl -X POST http://localhost:8080/issues \
     -H "Authorization: Bearer <ВАШ_ТОКЕН>" \
     -d '{"bookid": "<UUID_КНИГИ>", "userid": "<UUID_ПОЛЬЗОВАТЕЛЯ>"}'
```

## Ролевая модель
- **User:** Просмотр списка книг, просмотр своих текущих выдач.
- **Admin:** Полный доступ (CRUD книг, регистрация пользователей, оформление выдачи/возврата).

