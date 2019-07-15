package loader

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"
	// "github.com/gorilla/sessions"
)

var (
	errUserNotFound        = errors.New("User not found")
	errCannotCreateSession = errors.New("Cannot create session")
	errCannotParseForm     = errors.New("Cannot parse form")
	errCannotCreateUser    = errors.New("Cannot create user")
	errDBiternal           = errors.New("Internal DB error")
)

func logError(args ...interface{}) {
	log.Println(args...)
}

func newCookie(id int) http.Cookie {
	expiration := time.Now().Add(3 * time.Hour)
	cookie := http.Cookie{
		Name:     "_cookie",
		Value:    strconv.Itoa(id), // hardcoderd
		HttpOnly: true,
		Expires:  expiration,
	}
	return cookie
}

func (s *Server) stat(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("_cookie")
	loggedIn := (err != http.ErrNoCookie)
	if loggedIn {
		t := template.Must(template.ParseFiles("../../web/templates/stat.html"))
		tasks := s.GetUserTasks(cookie.Value)
		t.Execute(w, tasks)
	} else {
		http.Redirect(w, r, "/", 302)
	}

}

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) {
	user, err := s.UserByEmail(r.PostFormValue("email"))
	if err != nil {
		logError(errUserNotFound)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}
	if user.Password == s.Encrypt(r.PostFormValue("password")) {
		// if err != nil {
		// 	logError(errCannotCreateSession)
		// }
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
	r.ParseMultipartForm(5 * 1024 * 1025)
	meta := map[string]string{
		"email":    r.PostFormValue("email"),
		"password": r.PostFormValue("password"),
	}
	fmt.Println("signupAccount META:", meta)
	id, err := s.CreateUser(meta)
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

	err = r.ParseMultipartForm(5 * 1024 * 1025)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Huge file"))
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
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
