package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"deeplink-server/internal"

	"github.com/go-chi/chi/v5"
)

func setup() {
	internal.InitRedis()
}

func TestCreateHandler(t *testing.T) {
	setup()

	req := httptest.NewRequest("GET", "/create?code=test123&url=https://example.com", nil)
	w := httptest.NewRecorder()

	CreateHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestRedirectHandler(t *testing.T) {
	setup()

	// manually insert data
	internal.Rdb.Set(internal.Ctx, "dl:test123", "https://example.com", 0)

	// simulate request to /test123
	req := httptest.NewRequest("GET", "/test123", nil)
	w := httptest.NewRecorder()

	// simulate chi URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", "test123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	RedirectHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("Expected 302, got %d", resp.StatusCode)
	}
}
