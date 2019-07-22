package grader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	errBadLogin  = errors.New("Bad login")
	errBadPass   = errors.New("Bad password")
	errDBiternal = errors.New("Internal DB error")
)

func (uh *UserHandler) receiveTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only Post Method"))
		return
	}
	log.Println("Income message")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	task := &Task{}
	if err := json.Unmarshal(body, task); err != nil {
		http.Error(w, "Error in unmarshalling JSON", http.StatusInternalServerError)
	}
	log.Println("Task Unmarshalled: ", task)
	go func() {
		runTask(uh.DBCon, task)
	}()
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (uh *UserHandler) getRefreshToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]
	password := vars["password"]

	user, err := getUser(uh.DBCon, email, password)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	accesToken := uh.tm.refreshToken(user)
	if err := user.UpdateToken(uh.DBCon); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errDBiternal.Error()))
	}
	// tokens := map[string]map[string]string{"access_token": accessToken}
	RespondWithJSON(w, http.StatusOK, accesToken)
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.MarshalIndent(payload, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
