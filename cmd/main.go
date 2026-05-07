package main

import (
	"context"
	"library/internal/database"
	"library/internal/handlers"
	"library/internal/midlware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

type MyClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	db, err := database.InitDB(os.Getenv("DB_Directory"))
	if err != nil {
		slog.Error("Произошла ошибка при запуске базы данных: ", "err" ,err)
		os.Exit(1)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is work"))
	})

	h := &handlers.BooksHandler{Db: db}
	hu := &handlers.UserHandler{Db: db}
	r.Post("/login", hu.Login)
	r.Get("/books", h.ListBooks)
	r.Get("/books/{id}", h.InfoAboutBook)
	
	r.Group(func(r chi.Router){
		r.Use(midlware.AuthMiddleware("admin"))
		r.Post("/books", h.InsertBook)
		r.Put("/books/{id}", h.ChangeBook)
		r.Delete("/books/{id}", h.DeleteBook)
		r.Post("/users", hu.InsertUser)
		r.Post("/issues", makeHandler(hu.IssueBook))
		r.Post("/returns", makeHandler(hu.ReturnBook))
	})

	r.Group(func(r chi.Router){
		r.Use(midlware.AuthMiddleware("user"))
		r.Get("/users/{id}/books", hu.GetUserBooks)
	})
	port := os.Getenv("Server_Port")

	server := &http.Server{
		Addr: port,
		Handler: r,
	}
	go func(){
		slog.Info("Сервер запущен", "url", "http://localhost"+port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed{
			slog.Error("Ошибка в запуске сервера:", "err", err)
			os.Exit(1)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Завершение работы")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil{
		slog.Error("Ошибка при остановке сервера:", "err", err)
		os.Exit(1)
	}
	slog.Info("Сервер успешно остановлен")

}

func makeHandler(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		if err := fn(w,r); err != nil{
			//http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
