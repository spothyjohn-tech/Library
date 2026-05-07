package models

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID     uuid.UUID `json:"id"`
	Title  string    `json:"title"`
	Author string    `json:"author"`
	Isbn   string    `json:"isbn"`
	Year   int       `json:"year"`
	Status string    `json:"status"`
}

type User struct {
	ID                uuid.UUID `json:"id"`
	Name_user         string    `json:"name"`
	Email             string    `json:"email"`
	Password          string    `json:"password,omitempty"` 
	Role              string    `json:"role"`  
	Registration_date time.Time `json:"registrationdate"`
}

type Issue struct {
	ID          uuid.UUID  `json:"id"`
	ID_book     uuid.UUID  `json:"bookid"`
	ID_user     uuid.UUID  `json:"userid"`
	Issue_date  time.Time  `json:"issue"`
	Due_date    time.Time  `json:"duedate"`
	Return_date *time.Time `json:"returndate"`
}

