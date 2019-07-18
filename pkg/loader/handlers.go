package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/sarovkalach/gograder/pkg/grader"
)

var (
	errUserNotFound        = errors.New("User not found")
	errCannotCreateSession = errors.New("Cannot create session")
	errCannotParseForm     = errors.New("Cannot parse form")
	errCannotCreateUser    = errors.New("Cannot create user")
	errDBiternal           = errors.New("Internal DB error")
)

var (
	maxFileSize = 5 * 1024 * 1025
)

func logError(args ...interface{}) {
	log.Println(args...)
}

func (s *Server) stat(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("_cookie")
	loggedIn := (err != http.ErrNoCookie)
	if loggedIn {
		t := template.Must(template.ParseFiles("../../web/templates/stat.html"))
		tasks := GetTask(s.Uploader.DBCon, cookie.Value)
		t.Execute(w, tasks)
	} else {
		http.Redirect(w, r, "/", 302)
	}

}

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) {
	user, err := UserByEmail(s.Uploader.DBCon, r.PostFormValue("email"))
	if err != nil {
		logError(errUserNotFound)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}
	if user.Password == s.Encrypt(r.PostFormValue("password")) {
		cookie := newCookie(user.ID)
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/stat", 302)
	} else {
		http.Redirect(w, r, "/login", 302)
	}
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("../../web/templates/index.html"))
	t.Execute(w, nil)
}

func (s *Server) signup(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("../../web/templates/signup.html"))
	t.Execute(w, nil)

}

func (s *Server) signupAccount(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("SingupAccount", r.FormValue("email"))
	r.ParseMultipartForm(int64(maxFileSize))
	meta := map[string]string{
		"email":    r.PostFormValue("email"),
		"password": r.PostFormValue("password"),
	}
	fmt.Println("signupAccount META:", meta)
	id, err := CreateUser(s.Uploader.DBCon, meta)
	if err == errUserIsExists {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	cookie := newCookie(id)
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/stat", 302)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("../../web/templates/login.html"))
	t.Execute(w, nil)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("_cookie")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	session.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, session)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("_cookie")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if r.Method == "GET" {
		t := template.Must(template.ParseFiles("../../web/templates/upload.html"))
		t.Execute(w, nil)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad file consistent"))
		return
	}
	defer r.Body.Close()

	if len(body) > maxFileSize {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Huge file"))
		return
	}

	err = r.ParseMultipartForm(int64(maxFileSize))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Content-Type isn't multipart/form-data"))
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println("Error Retrieving the File:", err)
		return
	}
	defer file.Close()

	f, err := os.OpenFile("/tmp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	io.Copy(f, file)

	meta := map[string]string{"filename": handler.Filename}
	if err := s.Uploader.saveUserTask(meta); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Internal server Error"))
	}
	t := template.Must(template.ParseFiles("../../web/templates/success.html"))
	t.Execute(w, nil)
}

func (s *Server) receiverResult(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Server Receinve task ID:", mux.Vars(r)["id"])
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	result := grader.Result{}
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "Error in unmarshalling JSON", http.StatusInternalServerError)
	}
	log.Println("RESULT:", result)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
