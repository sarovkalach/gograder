package loader

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	// _ "github.com/go-sql-driver/mysql"
	// "github.com/gorilla/mux"
	// "github.com/gorilla/mux"
	"github.com/minio/minio-go/v6"
	"github.com/streadway/amqp"
)

var (
	fullPath              = "/home/kalach/gograder/minio-data/grader/"
	errDBconnection       = errors.New("Can not connect to DataBase")
	errAMQPconnetction    = errors.New("Can not connect to AMQP")
	errS3connetction      = errors.New("Can not connect to S3")
	errOpenChannel        = errors.New("Failed to open a channel")
	errAMQPDeclare        = errors.New("Failed to declare a queue")
	errAMQPMessagePublish = errors.New("Failed to publish a message")
)

var (
	// mysqlDSN = "kalach:1234@/grader?charset=utf8"
	mysqlDSN = "kalach:1234@tcp(mysql:3306)/grader"
	// amqpDSN = "amqp://guest:guest@localhost:5672/"
	amqpDSN = "amqp://guest:guest@rabbitmq:5672/"
	// s3URL = "127.0.0.1:9000"
	s3URL             = "minio:9000"
	defaultQueue      = "grader"
	defaultBucketName = "grader"
)

type FileLoader struct {
	s3Client *minio.Client
	// Router     *mux.Router
	DBCon   *sql.DB
	amqpCon *amqp.Connection
	queue   amqp.Queue
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

func showDBTasks(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		log.Println("SELECT Error:", err)
	}
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Status, &task.Course, &task.Name, &task.Filename, &task.S3BucketName, &task.UserID)
		if err != nil {
			log.Println("SELECT Error Read:", err)
		}
		fmt.Println(task)
	}
	rows.Close()
}

func NewFileLoader() (*FileLoader, error) {
	f := &FileLoader{}
	if err := f.initDBCon(); err != nil {
		return nil, err
	}
	if err := f.initS3(); err != nil {
		return nil, err
	}
	if err := f.initAMQPCon(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *FileLoader) initS3() error {
	accessKeyID := "9013HBZHIRHONH8ZIWL6"
	secretAccessKey := "gKIVgZaWAiuXbugPv9+dT4MKWsKlqxCyXFI+9ys+"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(s3URL, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return errS3connetction
	}
	f.s3Client = minioClient
	return nil
}

func (f *FileLoader) initDBCon() error {
	// dsn := "kalach:1234@/grader?charset=utf8"
	db, err := sql.Open("mysql", mysqlDSN)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		fmt.Println(err)
		return errDBconnection
	}
	f.DBCon = db
	return nil
}

func (f *FileLoader) initAMQPCon() error {
	conn, err := amqp.Dial(amqpDSN)
	if err != nil {
		// return err
		return errAMQPconnetction
	}
	f.amqpCon = conn

	ch, err := conn.Channel()
	if err != nil {
		return errOpenChannel
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		defaultQueue, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return errAMQPDeclare
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	f.queue = q
	return nil
	// Don't forget
	// defer conn.Close()
}

func (f *FileLoader) saveUserTask(meta map[string]string) error {
	id, _ := strconv.Atoi(meta["user_id"])
	task := &Task{
		Course:       meta["course"],
		Name:         meta["name"],
		Filename:     meta["filename"],
		S3BucketName: meta["bucket"],
		UserID:       id, //stub must be access token with email or id
	}

	uploadS3(f.s3Client, task)
	addDBTask(f.DBCon, task)
	addAmqpTask(f.amqpCon, f.queue, task)
	return nil
}

func uploadS3(s3Client *minio.Client, task *Task) {
	filename := task.Filename
	filePath := "/tmp/" + filename
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	n, err := s3Client.FPutObject(task.S3BucketName, filename, filePath, minio.PutObjectOptions{ContentType: "text/txt"})
	if err != nil {
		log.Fatalln(err, filePath, filename)
	}
	log.Printf("Successfully uploaded  S3 %s of size %d\n", filename, n)
}

func addDBTask(db *sql.DB, task *Task) {
	result, err := db.Exec(
		"INSERT INTO tasks (`status`, `course`, `name`, `filename`, s3bucketname,`user_id`) VALUES (?, ?, ?, ?,?,?)",
		task.Status,
		task.Course,
		task.Name,
		task.Filename,
		task.S3BucketName,
		task.UserID,
	)
	if err != nil {
		log.Println("INSERT Error:", err)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error reading Last ID", err)
	}
	task.ID = int(lastID)
	log.Printf("Successfully saved task to DB %s. ID = %d\n", task.Filename, lastID)
}

func addAmqpTask(amqpCon *amqp.Connection, queue amqp.Queue, task *Task) {
	ch, err := amqpCon.Channel()
	if err != nil {
		log.Println(errOpenChannel)
	}
	defer ch.Close()

	jsonTask, _ := json.Marshal(task)
	err = ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			// UserId:      task.User,
			// Type:        task.Course,
			Body: jsonTask,
		})

	if err != nil {
		log.Println(errAMQPMessagePublish)
	}
	log.Printf("Successfully pushed %s to AMQP. Queue: %s\n", task.Filename, queue.Name)
}
