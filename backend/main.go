package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	router "github.com/MicrosoftStudentChapter/Link-Generator/pkg/router"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

var allowedOrigins map[string]bool

const (
	rateLimitMax    = 10
	rateLimitWindow = time.Minute
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

	allowedOrigins = map[string]bool{}
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:5173"
	}
	for _, o := range strings.Split(origins, ",") {
		allowedOrigins[strings.TrimSpace(o)] = true
	}

	r := mux.NewRouter()

	r.HandleFunc("/links/all", router.GetAllLinks).Methods(http.MethodOptions, http.MethodGet)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service is Alive"))
	}).Methods(http.MethodOptions, http.MethodGet)
	r.Handle("/add-link", RateLimitMiddleware(http.HandlerFunc(router.AddLink))).Methods(http.MethodOptions, http.MethodPost)
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
		if origin := r.Header.Get("Origin"); allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
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

// RateLimitMiddleware caps requests per client IP using a Redis counter,
// since /add-link is an unauthenticated public write endpoint.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		key := "ratelimit:add-link:" + clientIP(r)
		count, err := router.Mem.Incr(ctx, key).Result()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if count == 1 {
			router.Mem.Expire(ctx, key, rateLimitWindow)
		}
		if count > rateLimitMax {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Too many requests, please try again later"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
