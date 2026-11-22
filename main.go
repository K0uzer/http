package main

import (
	// "encoding/json"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	// "time"
)

// 2 ________________________________________________________________

// mux := http.NewServeMux()

// mux.HandleFunc("GET /health", healthHandler)
// mux.HandleFunc("GET /time", timeHandler)
// mux.HandleFunc("POST /echo", echoHandler)

// mux.healthHandler(func (w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"status": "ok"}`))
// })
// mux.timeHandler(func (w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	res := map[string]string {
// 		"time": time.Now().UTC().Format(time.RFC3339),
// 	}
// 	json.NewEncoder(w).Encode(res)
// })
// mux.echoHandler(func (w http.ResponseWriter, r *http.Request) {
// 	var n json.RawMessage
// 	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
// 		http.Error(w, "невалидный JSON", http.StatusBadRequest)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(n)
// })

// 3 _________________________________________________________________________

	func RequestID(next http.Handler) http.Handler {
		var counter int64
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := atomic.AddInt64(&counter, 1)
			w.Header().Set("X-Request-ID", strconv.FormatInt(id, 10))
			next.ServeHTTP(w, r)
		})
	}

// 4 __________________________________________________________________________

type Task struct {
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Done   bool   `json:"done"`
}

var (
	tasks = make(map[int]Task)
	nextID = 1
	mu sync.Mutex
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", listTasts)
	mux.HandleFunc("POST /tasks", createTask)
	mux.HandleFunc("GET /tasks/{id}", getTask)
	mux.HandleFunc("PUT /tasks/{id}", changeTask)
	mux.HandleFunc("DELETE /tasks/{id}", deleteTast)

	fmt.Println("Сервер запущен на :8080")
	http.ListenAndServe(":8080", mux)
}

func listTasts(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	res := make([]Task, 0, len(tasks))

	for _, el := range tasks {
		res = append(res, el)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	
	var res Task
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "не валидный JSON", http.StatusBadRequest)
		return
	}
	if res.Title == "" {
		http.Error(w, "не валидный Title", http.StatusBadRequest)
		return
	}
	mu.Lock()
	res.ID = nextID
	nextID += 1
	tasks[res.ID] = res
	mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func getTask(w http.ResponseWriter, r *http.Request) {
	idText := r.PathValue("id")
	id, err := strconv.Atoi(idText)
	if err != nil {
		http.Error(w, "Не валидный id", http.StatusBadRequest)
		return
	}
	mu.Lock()
	task, ok := tasks[id]
	mu.Unlock()

	if !ok {
		http.Error(w, "Не корректный id", http.StatusNotFound)
		return
	} 

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func deleteTast(w http.ResponseWriter, r *http.Request) {
	idText := r.PathValue("id")
	id, err := strconv.Atoi(idText)
	if err != nil {
		http.Error(w, "невалидный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	_, ok := tasks[id]
	if ok {
		delete(tasks, id)
	}
	mu.Unlock()
	
	if !ok {
		http.Error(w, "Не корректный id", http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

func changeTask(w http.ResponseWriter, r *http.Request) {
	idText := r.PathValue("id")
	id, err := strconv.Atoi(idText)
	if err != nil {
		http.Error(w, "невалидный id", http.StatusBadRequest)
		return
	} 
	
	var update Task

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "невалидный JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	task, ok := tasks[id]
	if ok {
		if update.Title != "" {
			task.Title = update.Title
		}
		task.Done = update.Done
		tasks[id] = task
	}
	mu.Unlock()

	if !ok {
		http.Error(w, "Не корректный id", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}