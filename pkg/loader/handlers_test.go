package loader

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sarovkalach/gograder/pkg/grader"
)

type badRes struct {
	t  int
	c  struct{}
	f  float64
	ch chan int
}

func TestReceiverResult(t *testing.T) {
	res := grader.Result{Solved: true, Msg: "Task Solved"}
	result, _ := json.Marshal(res)
	req, _ := http.NewRequest("POST", "http://127.0.0.1:8080/results/1", bytes.NewBuffer(result))

	s := NewServer()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.receiverResult)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	req, _ = http.NewRequest("POST", "http://127.0.0.1:8080/results/1", bytes.NewBuffer([]byte("Response:test")))
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(s.receiverResult)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}
