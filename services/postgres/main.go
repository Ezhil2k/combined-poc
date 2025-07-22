package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/users/", userHandler)
	log.Println("Server started at :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users, err := GetAllUsers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(users)
	case http.MethodPost:
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		if u.Name == "" || u.Email == "" {
			http.Error(w, "name and email required", http.StatusBadRequest)
			return
		}
		id, err := CreateUser(u.Name, u.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		if u.Name == "" || u.Email == "" {
			http.Error(w, "name and email required", http.StatusBadRequest)
			return
		}
		err := UpdateUser(id, u.Name, u.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		err := DeleteUser(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
