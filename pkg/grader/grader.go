package grader

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v6"
	"github.com/sarovkalach/gograder/pkg/jwt"
)

var (
	callBackURL = "http://127.0.0.1:8080/results"
	mysqlDSN    = "kalach:1234@/grader?charset=utf8"
)

var (
	errS3connetction = errors.New("Can not connect to S3")
	errDBconnection  = errors.New("Can not connect to DataBase")
)

var (
	amqpDSN      = "amqp://guest:guest@localhost:5672/"
	s3URL        = "127.0.0.1:9000"
	defaultQueue = "grader"
)

var (
	emailReg    = regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}`)
	passwordReg = regexp.MustCompile(`[/\w|\W+/g]{8,}`)
)

type Grader struct {
	s3Client *minio.Client
	Router   *mux.Router
}

type Task struct {
	ID           int    `json:"id"`
	Status       int    `json:"status"`
	Course       string `json:"course"`
	Name         string `json:"name"`
	Filename     string `json:"filename"`
	S3BucketName string `json:"bucket"`
	UserID       int    `json:"user_id"`
}

type Result struct {
	ID     int    `json:"id"`
	Solved bool   `json:"solved"`
	Msg    string `json:"msg"`
}

func NewGrader() *Grader {
	g := &Grader{Router: mux.NewRouter()}
	if err := g.initRoutes(); err != nil {
		panic(err)
	}
	if err := g.initS3(); err != nil {
		panic(err)
	}
	return g
}

func (g *Grader) initRoutes() error {
	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return err
	}
	uh := &UserHandler{
		DBCon: db,
		token: jwt.NewJwtToken(jwt.SignKey),
	}

	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		fmt.Println(err)
		return errDBconnection
	}
	// refreshTokenStr := fmt.Sprintf("/get_token/{email:%s}&{password:%s}", emailReg, passwordReg)
	http.HandleFunc("/", uh.ReceiveTask)
	// g.Router.HandleFunc(refreshTokenStr, uh.GetRefreshToken).Methods("GET")
	return nil
}

func (g *Grader) Run() {
	log.Println("Grader Start")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func (g *Grader) initS3() error {
	accessKeyID := "9013HBZHIRHONH8ZIWL6"
	secretAccessKey := "gKIVgZaWAiuXbugPv9+dT4MKWsKlqxCyXFI+9ys+"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(s3URL, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return errS3connetction
	}
	g.s3Client = minioClient
	return nil
}

// func runTask(DBCon *sql.DB, tm *JwtToken, t *Task) {
func runTask(DBCon *sql.DB, t *Task, token string) {
	cmd := exec.Command("docker", "run", t.Name)
	if err := cmd.Run(); err != nil {
		res := &Result{Solved: false, Msg: err.Error()}
		sendResult(res, token)
		fmt.Println("SEND Error TO URL:>", callBackURL)
		return
	}
	sendResult(&Result{Solved: true, Msg: "Task solved", ID: t.ID}, token)
}

func sendResult(res *Result, token string) {
	result, err := json.Marshal(*res)
	if err != nil {
		log.Println("error in Marshall")
	}
	req, err := http.NewRequest("POST", callBackURL, bytes.NewBuffer(result))
	req.Header.Set("Authorization", token)
	if err != nil {
		log.Println("NewRequest ERR:", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error in send REsult")
	}
	defer resp.Body.Close()

	fmt.Println("response Status from Loader:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body from Loader:", string(body))
}
