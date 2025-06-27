package handlers

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"deeplink-server/internal"

	"github.com/go-chi/chi/v5"
	"github.com/skip2/go-qrcode"
)

func StartServer() {
	r := chi.NewRouter()
	r.Get("/create", CreateHandler)
	r.Get("/qr/{code}", QRHandler)
	r.Get("/stats/{code}", StatsHandler)
	r.Get("/{code}", RedirectHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server running on :" + port)
	http.ListenAndServe(":"+port, r)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	url := r.URL.Query().Get("url")
	expire := r.URL.Query().Get("expire")
	if code == "" || url == "" {
		http.Error(w, "code and url required", http.StatusBadRequest)
		return
	}

	key := "dl:" + code
	var err error
	if expire != "" {
		dur, err := time.ParseDuration(expire)
		if err != nil {
			http.Error(w, "invalid expiration", http.StatusBadRequest)
			return
		}
		err = internal.Rdb.Set(internal.Ctx, key, url, dur).Err()
	} else {
		err = internal.Rdb.Set(internal.Ctx, key, url, 0).Err()
	}
	if err != nil {
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("saved"))
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	key := "dl:" + code
	url, err := internal.Rdb.Get(internal.Ctx, key).Result()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ip := getIP(r)
	internal.Rdb.Incr(internal.Ctx, key+":count")
	internal.Rdb.RPush(internal.Ctx, key+":visits", fmt.Sprintf("%s|%s", ip, time.Now().Format(time.RFC3339)))

	http.Redirect(w, r, url, http.StatusFound)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	key := "dl:" + code

	count, _ := internal.Rdb.Get(internal.Ctx, key+":count").Result()
	logs, _ := internal.Rdb.LRange(internal.Ctx, key+":visits", 0, -1).Result()

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Visits: " + count + "\n"))
	for _, l := range logs {
		w.Write([]byte(l + "\n"))
	}
}

func QRHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	key := "dl:" + code
	url, err := internal.Rdb.Get(internal.Ctx, key).Result()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "QR generation failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}
