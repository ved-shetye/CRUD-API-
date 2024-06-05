package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Book struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	PublishedDate string `json:"published_date"`
	ISBN          string `json:"isbn"`
}

var db *sql.DB

func main() {
	var err error
	// Replace the following with your PostgreSQL connection string
	connStr := "user=postgres password=veds2003 dbname=postgres sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	http.HandleFunc("/books", booksHandler)
	http.HandleFunc("/books/", bookHandler)

	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable() {
	tableCreationQuery := `CREATE TABLE IF NOT EXISTS books (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        author TEXT NOT NULL,
        published_date TEXT,
        isbn TEXT UNIQUE NOT NULL
    );`
	_, err := db.Exec(tableCreationQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getBooks(w, r)
	case http.MethodPost:
		createBook(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/books/"):]
	bookID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getBook(w, r, bookID)
	case http.MethodPut:
		updateBook(w, r, bookID)
	case http.MethodDelete:
		deleteBook(w, r, bookID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, author, published_date, isbn FROM books")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	books := []Book{}
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.PublishedDate, &book.ISBN); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}
	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request, id int) {
	var book Book
	err := db.QueryRow("SELECT id, title, author, published_date, isbn FROM books WHERE id = $1", id).Scan(&book.ID, &book.Title, &book.Author, &book.PublishedDate, &book.ISBN)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Book not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(book)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.QueryRow("INSERT INTO books (title, author, published_date, isbn) VALUES ($1, $2, $3, $4) RETURNING id",
		book.Title, book.Author, book.PublishedDate, book.ISBN).Scan(&book.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func updateBook(w http.ResponseWriter, r *http.Request, id int) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE books SET title = $1, author = $2, published_date = $3, isbn = $4 WHERE id = $5",
		book.Title, book.Author, book.PublishedDate, book.ISBN, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	book.ID = id
	json.NewEncoder(w).Encode(book)
}

func deleteBook(w http.ResponseWriter, r *http.Request, id int) {
	_, err := db.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
