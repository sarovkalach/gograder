package loader

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var (
	errUserIsExists = errors.New("Such user is exist")
)

type Server struct {
	Router   *mux.Router
	Uploader *FileLoader
	// SessionStore *sessions.CookieStore
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

func NewServer() *Server {
	uploader, err := NewFileLoader()
	if err != nil {
		panic(err)
	}
	s := &Server{
		Router:   mux.NewRouter(),
		Uploader: uploader,
		// SessionStore: sessions.NewCookieStore([]byte(os.Getenv("TEST"))),
	}
	s.initRoutes()
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
	row := DBCon.QueryRow("SELECT * FROM users where email = ?", email)
	user := &User{}
	// err := row.Scan(&user.Id, &user.Email, &user.Password, &user.FirstName, &user.LastName)
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		//check User exists
		fmt.Println("ERROR UserByEmail:", err)
		return nil, err
	}
	return user, nil
}

func CreateUser(DBCon *sql.DB, meta map[string]string) (int, error) {
	_, err := UserByEmail(DBCon, meta["email"])
	if err == nil {
		return 0, errUserIsExists
	}

	hashedPasswd := fmt.Sprintf("%x", sha256.Sum256([]byte(meta["password"])))
	result, err := DBCon.Exec(
		"INSERT INTO users (`email`, `password`) VALUES (?, ?)",
		meta["email"],
		hashedPasswd,
	)

	lastID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error reading Last ID", err)
	}
	log.Printf("Successfully saved user:  %s. Last Insert ID = %d\n", meta["email"], lastID)
	return int(lastID), nil
}

func (s *Server) Encrypt(passwd string) string {
	hashedPasswd := fmt.Sprintf("%x", sha256.Sum256([]byte(passwd)))
	return hashedPasswd
}

func GetUserTasks(DBCon *sql.DB, id string) []Task {
	userID, _ := strconv.Atoi(id)
	rows, err := DBCon.Query("SELECT * FROM tasks WHERE user_id = ?", userID)
	if err != nil {
		log.Println("SELECT Error:", err)
	}
	tasks := make([]Task, 0, 16)
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Status, &task.Course, &task.Name, &task.Filename, &task.UserID)
		if err != nil {
			log.Println("SELECT Error Read:", err)
		}
		// fmt.Println(task)
		tasks = append(tasks, *task)
	}
	rows.Close()
	return tasks
}
