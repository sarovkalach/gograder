package grader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/minio/minio-go/v6"
)

var (
	callBackURL = "http://127.0.0.1:8080/results"
)

var (
	errS3connetction = errors.New("Can not connect to S3")
)

var (
	amqpDSN      = "amqp://guest:guest@localhost:5672/"
	s3URL        = "127.0.0.1:9000"
	defaultQueue = "grader"
)

type Grader struct {
	s3Client *minio.Client
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
	g := &Grader{}
	err := g.initS3()
	if err != nil {
		panic(err)
	}
	return &Grader{}
}

func (g *Grader) Run() {
	log.Println("Grader Start")
	http.HandleFunc("/", g.ReceiveTask)
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

func runTask(t *Task) {
	cmd := exec.Command("sleep", "10")
	if err := cmd.Run(); err != nil {
		res := &Result{Solved: false, Msg: err.Error()}
		sendResult(res)
		fmt.Println("SEND Error TO URL:>", callBackURL)
	}
	sendResult(&Result{Solved: true, Msg: "Task solved", ID: t.ID})
}

func sendResult(res *Result) {
	result, err := json.Marshal(*res)
	if err != nil {
		log.Println("error in Marshall")
	}
	req, err := http.NewRequest("POST", callBackURL, bytes.NewBuffer(result))
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

	fmt.Println("response Status from Server:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body from Server:", string(body))
}
