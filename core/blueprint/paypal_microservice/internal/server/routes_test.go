package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHelloWorldHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock server (without database dependency for this simple test)
	s := &Server{}

	router := gin.New()
	router.GET("/", s.HelloWorldHandler)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	expectedBody := `{"message":"PayPal Payment Service","version":"1.0.0"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}
