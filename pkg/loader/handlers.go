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
	"strconv"
	"text/template"

	"github.com/gorilla/sessions"
)

var (
	errBadLogopass         = errors.New("Bad credentials")
	errCannotCreateSession = errors.New("Cannot create session")
	errCannotParseForm     = errors.New("Cannot parse form")
	errCannotCreateUser    = errors.New("Cannot create user")
	errDBiternal           = errors.New("Internal DB error")
	errNoToken             = errors.New("No token")
	errBadToken            = errors.New("Bad Token")
)

var (
	maxFileSize = 5 * 1024 * 1025
)

func logError(args ...interface{}) {
	log.Println(args...)
}

// return error if were errors while save session
func saveSessionErr(session *sessions.Session, r *http.Request, w http.ResponseWriter) error {
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func (s *Server) stat(w http.ResponseWriter, r *http.Request) {
	session, err := s.ss.Get(r, "_cookie")
	fmt.Println("SESSION VALUES:", session.Values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !session.IsNew {
		t := template.Must(template.ParseFiles("web/templates/stat.html"))
		tasks := GetTask(s.Uploader.DBCon, session.Values["user"].(SessionUser).ID)
		t.Execute(w, tasks)
	} else {
		http.Redirect(w, r, "/", 302)
	}

}

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) {
	session, err := s.ss.Get(r, "_cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := UserByEmail(s.Uploader.DBCon, r.PostFormValue("email"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("User not found"))
		return
	}

	if user.Password != Encrypt(r.PostFormValue("password")) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Password"))
		return
	}

	session.Values["user"] = SessionUser{ID: user.ID, Authenticated: true}

	if err := saveSessionErr(session, r, w); err != nil {
		return
	}
	http.Redirect(w, r, "/stat", http.StatusFound)
	return

}

//main page
func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("web/templates/index.html"))
	t.Execute(w, nil)
}

// registration form
func (s *Server) signup(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("web/templates/signup.html"))
	t.Execute(w, nil)

}

func (s *Server) signupAccount(w http.ResponseWriter, r *http.Request) {
	session, err := s.ss.Get(r, "_cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !session.IsNew {
		http.Redirect(w, r, "/stat", http.StatusFound)
		return
	}

	r.ParseMultipartForm(int64(maxFileSize))
	meta := map[string]string{
		"email":    r.PostFormValue("email"),
		"password": Encrypt(r.PostFormValue("password")),
	}
	fmt.Println("signupAccount META:", meta)
	id, err := CreateUser(s.Uploader.DBCon, meta)
	if err == errUserIsExists {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	session.Values["user"] = SessionUser{ID: id, Authenticated: true}
	if err := saveSessionErr(session, r, w); err != nil {
		return
	}
	http.Redirect(w, r, "/stat", http.StatusFound)

}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	session, err := s.ss.Get(r, "_cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !session.IsNew {
		http.Redirect(w, r, "/stat", http.StatusFound)
		return
	}

	t := template.Must(template.ParseFiles("web/templates/login.html"))
	t.Execute(w, nil)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	session, err := s.ss.Get(r, "_cookie")
	if err != nil {
		log.Println("logout:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := session.Values["user"].(SessionUser)
	user.Authenticated = false
	session.Values["user"] = user
	session.Options.MaxAge = -1
	if err := saveSessionErr(session, r, w); err != nil {
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t := template.Must(template.ParseFiles("web/templates/upload.html"))
		t.Execute(w, nil)
		return
	}

	session, err := s.ss.Get(r, "_cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session.IsNew {
		http.Redirect(w, r, "/", http.StatusFound)
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
	fi, err := f.Stat()
	if fi.Size() > int64(maxFileSize) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Huge file"))
		return
	}

	meta := map[string]string{
		"filename": handler.Filename,
		"user_id":  strconv.Itoa(session.Values["user"].(SessionUser).ID),
		"course":   "golang", //	hardcoded
		"name":     "hw9",    //	hardcoded
		"bucket":   defaultBucketName,
	}

	if err := s.Uploader.saveUserTask(meta); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Internal server Error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File Uploaded"))
}

func (s *Server) receiverResult(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errNoToken.Error()))
		return
	}

	//hardcoded "1"
	res, err := s.token.Check(token)
	if !res || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errBadToken.Error()))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	result := Result{}
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "Error in unmarshalling JSON", http.StatusInternalServerError)
		return
	}

	log.Println("RESULT:", result)
	if err := saveResulstDB(result, s.Uploader.DBCon); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Internal server Error while saving results"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
