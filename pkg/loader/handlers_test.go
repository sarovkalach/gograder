package loader

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sarovkalach/gograder/pkg/grader"
)

func TestReceiverResult(t *testing.T) {
	res := grader.Result{Solved: true, Msg: "Task Solved"}
	result, _ := json.Marshal(res)
	req, _ := http.NewRequest("POST", "http://127.0.0.1:8080/results/1", bytes.NewBuffer(result))

	s := NewServer()
	//Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
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

func TestUploadHandler(t *testing.T) {
	//without cookie
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/upload", nil)
	s := NewServer()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.upload)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	// with cookie
	cookie := newCookie(1)
	req, _ = http.NewRequest("GET", "http://127.0.0.1:8080/upload", nil)
	rr = httptest.NewRecorder()
	req.AddCookie(&cookie)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// post Huge file
	buff := make([]byte, 1<<24)
	req, _ = http.NewRequest("POST", "http://127.0.0.1:8080/results/1", bytes.NewBuffer(buff))
	req.AddCookie(&cookie)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// post file not from form
	buff = make([]byte, 1<<12)
	req, _ = http.NewRequest("POST", "http://127.0.0.1:8080/results/1", bytes.NewBuffer(buff))
	req.AddCookie(&cookie)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
