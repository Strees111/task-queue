package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	workerpool "projectgo/internal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Task struct{
	Id string `json:"id"`
	Payload string `json:"payload"`
	Max_retries int `json:"max_retries"`
	Status string `json:"-"`
}

var wp *workerpool.WorkerPool
var server *http.Server

func ReadEnv() error{
	file, err := os.Open(".env")
	if err != nil {
		return err
	}
	defer file.Close()
	text := bufio.NewScanner(file)
	for text.Scan(){
        str := strings.TrimSpace(text.Text())
        if str == "" || strings.HasPrefix(str, "#") {
            continue // Пропускаем пустые строки и комментарии
        }
        
        res := strings.SplitN(str, "=", 2)
		if len(res) == 2{
            key := strings.TrimSpace(res[0])
            value := strings.TrimSpace(res[1])
            
            // Проверяем, что переменная окружения НЕ установлена
            if os.Getenv(key) == "" {
                os.Setenv(key, value)
            }
		}
	}
	return nil
}

func main(){
	ReadEnv()
	WORKERS, err := strconv.Atoi(os.Getenv("WORKERS"))
	
	if err != nil{
		WORKERS = 4
	}

	QUEUE_SIZE, err := strconv.Atoi(os.Getenv("QUEUE_SIZE"))
	
	if err != nil{
		QUEUE_SIZE = 64
	}
	log.Println("Starting pool with WORKERS:", WORKERS, "QUEUE_SIZE:", QUEUE_SIZE)
	wp = workerpool.NewWorkerPool(WORKERS, QUEUE_SIZE)


	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,    // SIGINT
    	syscall.SIGTERM, // SIGTERM
	)

	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/enqueue", Handler)
	mux.HandleFunc("/healthz", Handler)
	server = &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	
	go func(){
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("Server error: %v", err)
        }
	}()

	log.Println("Server started")

	<-ctx.Done()

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    log.Println("^C pressed, shutting down")

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }

	if err := wp.Stop(); err != nil{
		log.Printf("Worker pool stop error: %v", err)
	}

	log.Println("Server stopped")
}

func Handler(w http.ResponseWriter, r *http.Request){
    switch r.URL.Path {
    case "/healthz":
        if r.Method == http.MethodGet {
            HandlerGet(w, r)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    case "/enqueue":
        if r.Method == http.MethodPost {
            HandlerPost(w, r)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    default:
        http.Error(w, "Not found", http.StatusNotFound)
    }
}

func HandlerGet(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func HandlerPost(w http.ResponseWriter, r *http.Request){
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil{
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.Id == "" || task.Max_retries <= 0 {
        http.Error(w, "Invalid task parameters", http.StatusBadRequest)
        return
    }

	err := wp.Submit(func() {
		task.Status = "queued"
		log.Println("Task id: ", task.Id, " status: ",task.Status)
		processTask(&task)
	})
	if err != nil{
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(task)
}

func processTask(task *Task) {
	var err error
	baseDelay := 100 * time.Millisecond
	maxDelay := 5 * time.Second

	for attempt := 1; attempt <= task.Max_retries; attempt++ {
		err = doSomethingAlmostReliable(task)
		if err == nil {
			break 
		}
		log.Println("Task id: ", task.Id," attempt: ", attempt)
		if attempt == task.Max_retries {
			break
		}

		backoffTime := baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
		if backoffTime > maxDelay {
			backoffTime = maxDelay
		}
		// Full Jitter
		jitter := time.Duration(rand.Int63n(int64(backoffTime)))

		time.Sleep(jitter)
	}
}

func doSomethingAlmostReliable(task *Task) error {
	task.Status = "running"
	log.Println("Task id: ", task.Id, " status: ",task.Status)
	// Симуляция работы 
    workTime := 100 + rand.Intn(400)
    time.Sleep(time.Duration(workTime) * time.Millisecond)

	if rand.Intn(10) < 2 {
		task.Status = "failed"
		log.Println("Task id: ", task.Id, " status: ",task.Status)
		return errors.New("temporary failure")
	}
	task.Status = "done"
	log.Println("Task id: ", task.Id, " status: ",task.Status)
	return nil
}