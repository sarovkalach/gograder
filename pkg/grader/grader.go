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
	callBackURL = "http://127.0.0.1:8080/results/1"
)

var (
	errS3connetction = errors.New("Can not connect to S3")
)

var (
	mysqlDSN     = "kalach:1234@/grader?charset=utf8"
	amqpDSN      = "amqp://guest:guest@localhost:5672/"
	s3URL        = "127.0.0.1:9000"
	defaultQueue = "grader"
)

type Grader struct {
	s3Client *minio.Client
}

type Task struct {
	Name    string `json:"name"`
	User    string `json:"user,omitempty"`
	Timeout int    `json:"timeout"`
}

type Result struct {
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
	cmd := exec.Command("touch", "/home/kalach/Grader_Command.txt")
	if err := cmd.Run(); err != nil {
		res := &Result{Solved: false, Msg: err.Error()}
		sendResult(res)
		fmt.Println("SEND Error TO URL:>", callBackURL)
	}
	sendResult(&Result{Solved: false, Msg: "Task solved"})
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

// func (q *Queuer) SendTask() {

// 	fmt.Println("URL:>", graderURL)
// 	var jsonStr = []byte(`{"name":"Test message", "timeout":3}`)
// 	req, err := http.NewRequest("POST", graderURL, bytes.NewBuffer(jsonStr))
// 	if err != nil {
// 		log.Println("NewRequest ERR:", err)
// 	}
// 	req.Header.Set("X-Custom-Header", "myvalue")
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{Timeout: 5 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	fmt.Println("response Status:", resp.Status)
// 	fmt.Println("response Headers:", resp.Header)
// 	body, _ := ioutil.ReadAll(resp.Body)
// 	fmt.Println("response Body:", string(body))
// }
