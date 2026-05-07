package handlers

import (
	"database/sql"
	"encoding/json"
	"library/internal/models"
	"net/http"
	"time"
	"errors"
	"log/slog"
	"golang.org/x/crypto/bcrypt"
	"library/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct{
	Db *sql.DB
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func (U *UserHandler) InsertUser(w http.ResponseWriter, r * http.Request){
	defer r.Body.Close()

	var newUser models.User

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		slog.Error("Ошибка в добавлении пользователя", "err", err)
		http.Error(w, "Ошибка в формате json", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost) 
	if err != nil {
		slog.Error("Ошибка при хешировании пароля", "err", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	UserId, err := uuid.NewRandom()
	if err != nil {
		slog.Error("Ошибка в создании UUID", "err", err)
		http.Error(w, "Ошибка при добавлении в базу", http.StatusInternalServerError)
		return
	}

	queryItem := `INSERT INTO Users(ID, "Name", Email, Password, Role, RegistrationDate) VALUES(?,?,?,?,?,?)`

	_, err = U.Db.Exec(queryItem, UserId, newUser.Name_user, newUser.Email, hashedPassword, "user",time.Now())
	if err != nil{
		slog.Error("Ошибка при добавлении в базу", "err", err)
		http.Error(w, "Ошибка при добавлении в базу", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(" Пользователь зарегистрирован"))
}

func (U *UserHandler) GetUserBooks(w http.ResponseWriter, r * http.Request){
	defer r.Body.Close()

	UserId := chi.URLParam(r, "id")

	query := `SELECT DISTINCT Books.ID, Books.Title, Books.Author, Books.ISBN, Books."Year", Books."Status" FROM Issue, Books, Users WHERE Issue.ReaderID = ? AND Books.ID = Issue.BookID AND Issue.ReturnDate IS NULL`
	rows, err := U.Db.Query(query, UserId)
	if err != nil{
		slog.Error("Ошибка выдачи книг пользователя", "err", err)
		http.Error(w, "Ошибка при добавлении в базу", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var UserBooks []models.Book

	for rows.Next() {
		var Book models.Book
		err = rows.Scan(&Book.ID, &Book.Title, &Book.Author, &Book.Isbn, &Book.Year, &Book.Status)
		if err != nil{
			slog.Error("Ошибка", "err", err)
			continue
		}
		UserBooks = append(UserBooks, Book)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserBooks)
}

func (U *UserHandler) IssueBook(w http.ResponseWriter, r *http.Request) error{
	defer r.Body.Close()
	var IssueBook models.Issue

	tx, err := U.Db.Begin()
	if err != nil{
		slog.Error("Ошибка транзакции", "err", err)
		http.Error(w, "Ошибка транзакции", http.StatusInternalServerError)
		return err
	}
	defer tx.Rollback()

	err = json.NewDecoder(r.Body).Decode(&IssueBook)
	if err != nil{
		slog.Error("Ошибка", "err", err)
		http.Error(w, "Ошибка в формате json", http.StatusBadRequest)
		return err
	}

	query := `SELECT "Status" FROM Books WHERE ID=? AND "Status"='Available'`
	line := tx.QueryRow(query, IssueBook.ID_book).Scan(&IssueBook.ID_book)
	if errors.Is(line, sql.ErrNoRows){
		http.Error(w, "Данной книги нет или она уже выдана", http.StatusBadGateway)
		return nil
	}

 	IssueUUID, err := uuid.NewRandom()

	if err != nil{
		slog.Error("Ошибка создания uuid","err", err)
		http.Error(w, "Ошибка при добавлении в базу", http.StatusInternalServerError)
		return err
	}

	query = `INSERT INTO Issue(ID, BookID, ReaderID,IssueDate, DueDate) VALUES(?,?,?,?,?)`
	NowTime := time.Now()
	_, err = tx.Exec(query, IssueUUID, IssueBook.ID_book, IssueBook.ID_user, time.Now(), NowTime.AddDate(0,0,14))
	if err != nil{
		slog.Error("Ошибка выдачи книг пользователю", "err", err)
		http.Error(w, "Ошибка при добавлении в базу", http.StatusInternalServerError)
		return err
	}

	query = `UPDATE Books SET "Status"='Issued' WHERE ID=?`

	_, err = tx.Exec(query, IssueBook.ID_book)
	if err != nil{
		slog.Error("Не удалось обновить данные о книге", "err", err)
		http.Error(w, "Ошибка при обновлении данных", http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Пользователь взял книгу"))
	return tx.Commit()
}

func (U *UserHandler) ReturnBook(w http.ResponseWriter, r *http.Request) error{
	defer r.Body.Close()
	var IssueBook models.Issue

	tx, err := U.Db.Begin()
	if err != nil{
		slog.Error("Ошибка транзакции", "err", err)
		http.Error(w, "Ошибка транзакции", http.StatusInternalServerError)
		return err
	}
	defer tx.Rollback()

	err = json.NewDecoder(r.Body).Decode(&IssueBook)
	if err != nil{
		slog.Error("Ошибка","err", err)
		http.Error(w, "Ошибка в формате json", http.StatusBadRequest)
		return err
	}
	if IssueBook.ID == uuid.Nil{
		slog.Error("Не указан ID выдачи")
		http.Error(w, "Ошибка в формате json", http.StatusBadRequest)
		return nil
	}

	query := `SELECT "Status" FROM Books WHERE ID=? AND "Status"='Issued'`
	line := tx.QueryRow(query, IssueBook.ID_book).Scan(&IssueBook.ID_book)
	
	if errors.Is(line, sql.ErrNoRows){
		slog.Error("Ошибка", "err", err)
		http.Error(w, "Данной книги нет или она уже возвращена", http.StatusBadGateway)
		return err
	}

	query = `SELECT BookID, ReaderID FROM Issue WHERE BookID=? AND ReaderID=? AND ReturnDate is NULL`
	line = tx.QueryRow(query, IssueBook.ID_book, IssueBook.ID_user).Scan(&IssueBook.ID_book)
	if errors.Is(line, sql.ErrNoRows){
		slog.Error("Эта книга не была выдана этому читателю", "err", err)
		http.Error(w, "Данный читатель не брал эту книгу", http.StatusBadGateway)
		return err
	}

	query = `UPDATE Books SET "Status"='Available' WHERE ID=?`

	_, err = tx.Exec(query, IssueBook.ID_book)
	if err != nil{
		slog.Error("Не удалось обновить данные о книге", "err", err)
		http.Error(w, "Ошибка при обновлении данных", http.StatusInternalServerError)
		return err
	}

	query = `UPDATE Issue SET ReturnDate=? WHERE ID=?`

	_, err = tx.Exec(query,time.Now(),IssueBook.ID)
	if err != nil{
		slog.Error("Не удалось обновить данные о книге", "err", err)
		http.Error(w, "Ошибка при обновлении данных", http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Пользователь отдал книгу"))
	return tx.Commit()
}

func (U *UserHandler) Login(w http.ResponseWriter, r *http.Request){
	 defer r.Body.Close()
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		slog.Error("Ошибка", "err", err)
		http.Error(w, "Ошибка в формате json", http.StatusBadRequest)
		return
	}
	var user models.User
	
	query := `SELECT ID, Password, Role FROM Users WHERE Email=?`
	res := U.Db.QueryRow(query, req.Email).Scan(&user.ID, &user.Password, &user.Role)
	if errors.Is(res, sql.ErrNoRows){
		slog.Error("Данного пользователя нет")
		http.Error(w, "Данного пользователя нет", http.StatusBadGateway)
		return 
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
    if err != nil { 
		http.Error(w, "Неверный пароль", http.StatusUnauthorized); 
		return 
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		slog.Error("Генерация токена не удалась", "err", err)
		http.Error(w, "Генерация токена не удалась", http.StatusBadGateway)
		return 
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}