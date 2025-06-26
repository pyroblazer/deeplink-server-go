package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis not connected: %v", err)
	}
}

func createLink(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	url := r.URL.Query().Get("url")
	if code == "" || url == "" {
		http.Error(w, "code and url required", http.StatusBadRequest)
		return
	}
	err := rdb.Set(ctx, "dl:"+code, url, 0).Err()
	if err != nil {
		http.Error(w, "could not save", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("saved"))
}

func resolveLink(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	url, err := rdb.Get(ctx, "dl:"+code).Result()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func main() {
	initRedis()

	r := chi.NewRouter()
	r.Get("/create", createLink)
	r.Get("/{code}", resolveLink)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Listening on :" + port)
	http.ListenAndServe(":"+port, r)
}
