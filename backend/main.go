package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	router "github.com/MicrosoftStudentChapter/Link-Generator/pkg/router"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func main() {
	var opts *redis.Options
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		fmt.Println("Connecting Redis via REDIS_URL")
		var err error
		opts, err = redis.ParseURL(redisURL)
		if err != nil {
			panic(err)
		}
	} else {
		redisAddr := os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			redisAddr = ":6379"
		}
		fmt.Println("Connecting Redis to: ", redisAddr)
		opts = &redis.Options{
			Addr:     redisAddr,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		}
	}
	conn := redis.NewClient(opts)
	router.Mem = conn
	res, err := conn.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Redis [PING]: ", res)

	r := mux.NewRouter()

	r.HandleFunc("/links/all", router.GetAllLinks).Methods(http.MethodOptions, http.MethodGet)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service is Alive"))
	}).Methods(http.MethodOptions, http.MethodGet)
	r.HandleFunc("/add-link", router.AddLink).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/{link}", router.HandleRouting).Methods(http.MethodOptions, http.MethodGet)

	r.Use(LoggingMiddleware)
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(HandlePreflight)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	fmt.Println("Server started at port", port)

	http.ListenAndServe(":"+port, r)
}

// Middlewares

func HandlePreflight(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Logger:  ", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
