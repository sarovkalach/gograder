package loader

import (
	"crypto/sha256"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	errUserIsExists = errors.New("Such user is exist")
)

type Server struct {
	Router   *mux.Router
	Uploader *FileLoader
	ss       *sessions.CookieStore
}

func NewServer() *Server {
	uploader, err := NewFileLoader()
	if err != nil {
		panic(err)
	}
	s := &Server{
		Router:   mux.NewRouter(),
		Uploader: uploader,
		// it must no be saved in code or smth else. Only by os.Getenv. Hardcoded
		ss: sessions.NewCookieStore([]byte("PRI")),
	}
	s.initRoutes()
	gob.Register(SessionUser{})
	return s
}

func (s *Server) Run() {
	log.Println("Loader Start")
	log.Fatal(http.ListenAndServe(":8080", s.Router))
}

func (s *Server) initRoutes() {
	s.Router.HandleFunc("/", s.index)
	s.Router.HandleFunc("/authenticate", s.authenticate).Methods("POST")
	s.Router.HandleFunc("/login", s.login)
	s.Router.HandleFunc("/logout", s.logout)
	s.Router.HandleFunc("/signup", s.signup).Methods("GET")
	s.Router.HandleFunc("/signup_account", s.signupAccount).Methods("POST")
	s.Router.HandleFunc("/upload", s.upload)
	s.Router.HandleFunc("/stat", s.stat)
	s.Router.HandleFunc("/results/{id:[0-9]+}", s.receiverResult).Methods("POST")
}

func UserByEmail(DBCon *sql.DB, email string) (*User, error) {
	row := DBCon.QueryRow("SELECT * FROM users WHERE email = ?", email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		//check User exists
		return nil, err
	}
	return user, nil
}

func CreateUser(DBCon *sql.DB, meta map[string]string) (int, error) {
	_, err := UserByEmail(DBCon, meta["email"])
	if err == nil {
		return 0, errUserIsExists
	}

	result, err := DBCon.Exec(
		"INSERT INTO users (`email`, `password`) VALUES (?, ?)",
		meta["email"],
		meta["password"],
	)

	if err != nil {
		return 0, errDBiternal
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error reading Last ID", err)
	}
	log.Printf("Successfully saved user:  %s. Last Insert ID = %d\n", meta["email"], lastID)
	return int(lastID), nil
}

func GetTask(DBCon *sql.DB, userID int) []Task {
	rows, err := DBCon.Query("SELECT * FROM tasks WHERE user_id = ?", userID)
	if err != nil {
		log.Println("SELECT Error:", err)
	}
	tasks := make([]Task, 0, 16)
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Status, &task.Course, &task.Name, &task.Filename, &task.S3BucketName, &task.UserID)
		if err != nil {
			log.Println("SELECT Error Read:", err)
		}
		// fmt.Println(task)
		tasks = append(tasks, *task)
	}
	rows.Close()
	return tasks
}

func Encrypt(passwd string) string {
	hashedPasswd := fmt.Sprintf("%x", sha256.Sum256([]byte(passwd)))
	return hashedPasswd
}
