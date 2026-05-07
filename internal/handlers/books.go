package handlers

import (
	"database/sql"
	"encoding/json"
	"library/internal/models"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BooksHandler struct {
	Db *sql.DB
}

func (B *BooksHandler) ListBooks(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	var lim,off int
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	author := r.URL.Query().Get("author")
	status := r.URL.Query().Get("status")

	if limit == "" {
		lim = 10
	}else{
		lim1, err := strconv.Atoi(limit)
		if err != nil {
			lim = 10
		} else{
			lim = lim1
		}
	}
	if (offset) == ""{
		off = 0
	} else{
		off1, err := strconv.Atoi(offset)
		if err != nil {
			off = 0
		} else{
			off = off1
		}
	}
	query := `Select ID, Title, Author, ISBN, "Year", "Status" FROM Books WHERE 1=1`

	var args []interface{}

	if author != ""{
		query += " AND Author = ?"
    	args = append(args, author)
	}
	if status != "" {
		query += " AND Status = ?"
		args = append(args, status)
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, lim, off)
	
	rows, err := B.Db.Query(query,args...)

	if err != nil{
		slog.Error("Ошибка в выводе книг", "err", err)
		http.Error(w, "Ошибка в выводе книг", http.StatusInternalServerError)
		return
	}
	defer rows.Close()



	var books []models.Book
	for rows.Next() {
		var book  models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Isbn, &book.Year, &book.Status)
		if err != nil{
			slog.Error("Ошибка в выводе книг", "err", err)
			continue
		}
		books = append(books, book)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (B *BooksHandler) InsertBook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var newBook models.Book
	err := json.NewDecoder(r.Body).Decode(&newBook)
	if err != nil{
		slog.Error("Ошибка при чтении JSON", "err", err,)
		http.Error(w, "Ошибка в формате JSON", http.StatusBadRequest)
		return 
	}

	BookId, err := uuid.NewRandom()
	if err != nil {
		slog.Error("Ошибка при генерации UUID","err", err,)	
		http.Error(w, "Ошибка при добавлении в базу", http.StatusInternalServerError)
		return 
	}

	if !CheckISBN(newBook.Isbn){
		slog.Error("Ошибка в номере ISBN")
		http.Error(w, "Ошибка при добавлении в базу - неверный ISBN", http.StatusInternalServerError)
		return 
	}

	queryItem := `INSERT INTO Books(ID, Title, Author, ISBN, "Year", "Status") VALUES(?,?,?,?,?,'Available')`
	_, err = B.Db.Exec(queryItem, BookId.String(), newBook.Title, newBook.Author, newBook.Isbn, newBook.Year)

	if err != nil{
		slog.Error("Ошибка при вставке в БД","err", err, "book_title", newBook.Title)
		http.Error(w, "Ошибка при добавлении в базу", http.StatusInternalServerError)
		return 
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Книга добавлена"))
}

func (B *BooksHandler) InfoAboutBook(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	var book models.Book
	id := chi.URLParam(r, "id")
	query := `SELECT ID, Title, Author, ISBN, "Year", "Status" FROM Books WHERE ID=?`
	line := B.Db.QueryRow(query, id)
	err := line.Scan(&book.ID, &book.Title ,&book.Author, &book.Isbn, &book.Year, &book.Status)
	if err == sql.ErrNoRows{
		slog.Error("Нет информации о книге","err", err)
		http.Error(w, "Книга не найдена", http.StatusNotFound)
		return
	} else if err != nil {
		slog.Error("Ошибка при выводе информации о книге","err", err)
		http.Error(w, "Ошибка сервера", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (B *BooksHandler) ChangeBook(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	var ChangeBook models.Book
	err := json.NewDecoder(r.Body).Decode(&ChangeBook)
	if err != nil {
		slog.Error("Ошибка","err", err)
		http.Error(w, "Ошибка в формате json", http.StatusBadRequest)
		return
	}
	if !CheckISBN(ChangeBook.Isbn){
		slog.Error("Ошибка в номере ISBN")
		http.Error(w, "Ошибка при добавлении в базу", http.StatusInternalServerError)
		return 
	}
	id := chi.URLParam(r, "id")
	request := `UPDATE Books SET Title=?, Author=?, ISBN=?, "Year"=?, "Status"=? WHERE ID=?`
	res, err := B.Db.Exec(request, ChangeBook.Title,ChangeBook.Author,ChangeBook.Isbn,ChangeBook.Year,ChangeBook.Status,id)
	count, _ := res.RowsAffected() 
	if count == 0 {
		slog.Error("Книга не найдена")
		http.Error(w, "Книга не найдена", http.StatusNotFound)
		return
	}
	if err != nil{
		slog.Error("Ошибка: %v","err", err)
		http.Error(w, "Ошибка обновления данных о книге", http.StatusInternalServerError)
		return
	}


	
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Данные о книге обновлены"))
}

func (B *BooksHandler) DeleteBook(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()

	request := `DELETE FROM Books WHERE ID=?`
	id := chi.URLParam(r, "id")
	res ,err := B.Db.Exec(request, id)
	count, _ := res.RowsAffected()
	if err != nil {
		slog.Error("Ошибка при удалении книги","err", err)
		http.Error(w, "Ошибка при удалении книги", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		slog.Error("Книга не найдена")
		http.Error(w, "Книга не найдена", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Данные о книге стёрты"))
}

func CheckISBN(ISBN string) bool {
	clean := strings.ReplaceAll(ISBN, "-", "")
	clean = strings.ReplaceAll(clean, " ", "")
	
	match, _ := regexp.MatchString(`^\d{13}$`, clean)
	if !match {
		return false
	}

	var sum int
	for i, char := range clean{
		digit, _ := strconv.Atoi(string(char))
		if i == 12 {
			break
		}
		if i%2 == 0 {
			sum += digit *1
		} else{
			sum += digit *3
		}
	}
	checkDigit := (10 - (sum % 10)) % 10
	lastDigit, _ := strconv.Atoi(string(clean[12]))
	return checkDigit == lastDigit
}