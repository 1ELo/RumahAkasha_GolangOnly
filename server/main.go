package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Event struct {
	Title       string  `json:"title"`
	Date        string  `json:"date"`
	Time        string  `json:"time"`
	Image       string  `json:"image"`
	Description string  `json:"description"`
	Fee         float64 `json:"fee"`
	Status      string  `json:"status"`
}

type MenuItemCoffee struct {
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Kategori    string  `json:"kategori"`
}

type MenuItemNoncoffee struct {
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Kategori    string  `json:"kategori"`
}

type MenuItemSignature struct {
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Kategori    string  `json:"kategori"`
}

type MenuItemFood struct {
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Kategori    string  `json:"kategori"`
}

type Barista struct {
	ID          int    `json:"id_barista"`
	Name        string `json:"nama_barista"`
	Photo       string `json:"foto_barista"`
	Description string `json:"deskripsi"`
	Year        string `json:"tahun_kerja"`
	JobDesk     string `json:"job_desk"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "user2:123456@tcp(192.168.244.133:3306)/rumahakasha")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Function Owner & Staff
	http.HandleFunc("/api/create-event", createEvent)
	http.HandleFunc("/api/edit-event", editEvent)
	http.HandleFunc("/api/delete-event", delEvent)

	// Owner Menu
	http.HandleFunc("/api/formCoffee", createMenuCoffee)
	http.HandleFunc("/api/formNonCoffee", createMenuNoncoffee)
	http.HandleFunc("/api/formSignature", createMenuSignature)
	http.HandleFunc("/api/formFood", createMenuFood)
	http.HandleFunc("/api/editMenu", editMenu)
	http.HandleFunc("/api/delete-menu", delMenu)

	// Owner Barista
	http.HandleFunc("/api/create-barista", createBarista)
	http.HandleFunc("/api/edit-barista", editBarista)
	http.HandleFunc("/api/delete-barista", delBarista)

	// Tes DB
	http.HandleFunc("/test-db", testDBConnection)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func testDBConnection(w http.ResponseWriter, r *http.Request) {
	err := db.Ping()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to database: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Database connection successful!")
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed !", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var event Event
	event.Title = r.FormValue("title")
	event.Date = r.FormValue("date")
	event.Time = r.FormValue("time")
	event.Description = r.FormValue("description")
	event.Fee, _ = strconv.ParseFloat(r.FormValue("fee"), 64)
	event.Status = r.FormValue("status")

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	event.Image = filePath

	_, err = db.Exec("INSERT INTO events (title, date, time, image, description, fee, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		event.Title, event.Date, event.Time, event.Image, event.Description, event.Fee, event.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event created successfully"})
}

func editEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed !", http.StatusMethodNotAllowed)
		return
	}

	var event Event
	event.Title = r.FormValue("title")
	event.Date = r.FormValue("date")
	event.Time = r.FormValue("time")
	event.Description = r.FormValue("description")
	event.Fee, _ = strconv.ParseFloat(r.FormValue("fee"), 64)
	event.Status = r.FormValue("status")

	// Handle image update
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		event.Image = filePath
	}

	// Update event in the database
	_, err = db.Exec("UPDATE events SET title = ?, date = ?, time = ?, image = ?, description = ?, fee = ?, status = ? WHERE id = ?",
		event.Title, event.Date, event.Time, event.Image, event.Description, event.Fee, event.Status, r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event updated successfully"})
}

func delEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed !", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM events WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event deleted successfully"})
}

// COFFEE -- OWNER
func createMenuCoffee(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed !", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var menuCoffee MenuItemCoffee
	menuCoffee.Name = r.FormValue("name")
	menuCoffee.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
	menuCoffee.Description = r.FormValue("description")
	menuCoffee.Kategori = r.FormValue("kategori")

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	menuCoffee.Image = filePath

	_, err = db.Exec("INSERT INTO menus (name, image, price, description, kategori) VALUES (?, ?, ?, ?, ?)",
		menuCoffee.Name, menuCoffee.Image, menuCoffee.Price, menuCoffee.Description, "coffee")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Menu Coffee created successfully"})
}

func editMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed !", http.StatusMethodNotAllowed)
		return
	}

	var menuCoffee MenuItemCoffee
	menuCoffee.Name = r.FormValue("name")
	menuCoffee.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
	menuCoffee.Description = r.FormValue("description")
	menuCoffee.Kategori = r.FormValue("kategori")

	// Handle image update
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		menuCoffee.Image = filePath
	}

	// Update event in the database
	_, err = db.Exec("UPDATE menus SET name = ?, image = ?, price = ?, description = ?, kategori = ? WHERE id_menu = ?",
		menuCoffee.Name, menuCoffee.Image, menuCoffee.Price, menuCoffee.Description, menuCoffee.Kategori, r.FormValue("id_menu"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Menu Coffee updated successfully"})
}

func delMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed !", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Menu Coffee ID is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM menus WHERE id_menu = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Menu Coffee deleted successfully"})
}

// Non Coffee
func createMenuNoncoffee(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed !", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var menuNoncoffee MenuItemNoncoffee
	menuNoncoffee.Name = r.FormValue("name")
	menuNoncoffee.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
	menuNoncoffee.Description = r.FormValue("description")
	menuNoncoffee.Kategori = r.FormValue("Kategori")

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	menuNoncoffee.Image = filePath

	_, err = db.Exec("INSERT INTO menus (name, image, price, description, kategori) VALUES (?, ?, ?, ?, ?)",
		menuNoncoffee.Name, menuNoncoffee.Image, menuNoncoffee.Price, menuNoncoffee.Description, "noncoffee")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Menu Non-Coffee created successfully"})
}

// Signature
func createMenuSignature(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var menuSignature MenuItemSignature
	menuSignature.Name = r.FormValue("name")
	menuSignature.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
	menuSignature.Description = r.FormValue("description")
	menuSignature.Kategori = r.FormValue("kategori")

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	menuSignature.Image = filePath

	_, err = db.Exec("INSERT INTO menus (name, image, price, description, kategori) VALUES (?, ?, ?, ?, ?)",
		menuSignature.Name, menuSignature.Image, menuSignature.Price, menuSignature.Description, "signature")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Menu Signature created successfully"})
}

// FOOD
func createMenuFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed !", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var menuFood MenuItemFood
	menuFood.Name = r.FormValue("name")
	menuFood.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
	menuFood.Description = r.FormValue("description")
	menuFood.Kategori = r.FormValue("kategori")

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	menuFood.Image = filePath

	_, err = db.Exec("INSERT INTO menus (name, image, price, description, kategori) VALUES (?, ?, ?, ?, ?)",
		menuFood.Name, menuFood.Image, menuFood.Price, menuFood.Description, "food")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Menu Coffee created successfully"})
}

// Barista API

func createBarista(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var barista Barista
	barista.Name = r.FormValue("nama_barista")
	barista.Description = r.FormValue("deskripsi")
	barista.Year = r.FormValue("tahun_kerja")
	barista.JobDesk = r.FormValue("job_desk")

	file, handler, err := r.FormFile("foto_barista")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	barista.Photo = filePath

	_, err = db.Exec("INSERT INTO baristas (nama_barista, foto_barista, deskripsi, tahun_kerja, job_desk) VALUES (?, ?, ?, ?, ?)",
		barista.Name, barista.Photo, barista.Description, barista.Year, barista.JobDesk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Barista created successfully"})
}

func editBarista(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var barista Barista
	barista.ID, _ = strconv.Atoi(r.FormValue("id_barista"))
	barista.Name = r.FormValue("nama_barista")
	barista.Description = r.FormValue("deskripsi")
	barista.Year = r.FormValue("tahun_kerja")
	barista.JobDesk = r.FormValue("job_desk")

	file, handler, err := r.FormFile("foto_barista")
	if err == nil {
		defer file.Close()
		filePath := fmt.Sprintf("/mnt/laravel_barista_images/%s", handler.Filename)
		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		barista.Photo = filePath
	}

	if barista.Photo != "" {
		_, err = db.Exec("UPDATE baristas SET nama_barista = ?, foto_barista = ?, deskripsi = ?, tahun_kerja = ?, job_desk = ? WHERE id_barista = ?",
			barista.Name, barista.Photo, barista.Description, barista.Year, barista.JobDesk, barista.ID)
	} else {
		_, err = db.Exec("UPDATE baristas SET nama_barista = ?, deskripsi = ?, tahun_kerja = ?, job_desk = ? WHERE id_barista = ?",
			barista.Name, barista.Description, barista.Year, barista.JobDesk, barista.ID)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Barista updated successfully"})
}

func delBarista(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id_barista")
	if id == "" {
		http.Error(w, "Barista ID is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM baristas WHERE id_barista = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Barista deleted successfully"})
}
