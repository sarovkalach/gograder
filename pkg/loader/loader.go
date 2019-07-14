package loader

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

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

type FileLoader struct {
	s3Client *minio.Client
	// Router     *mux.Router
	DBCon   *sql.DB
	amqpCon *amqp.Connection
	queue   amqp.Queue
}

type Task struct {
	ID       int
	Graded   bool
	Course   string
	Task     string
	User     string
	Filename string
}

func showDBTasks(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		log.Println("SELECT Error:", err)
	}
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Graded, &task.Course, &task.Task, &task.User, &task.Filename)
		if err != nil {
			log.Println("SELECT Error Read:", err)
		}
		fmt.Println(task)
	}
	rows.Close()
}

func NewFileLoader() *FileLoader {
	f := &FileLoader{}
	f.initS3()
	f.initDBCon()
	f.initAMQPCon()
	return f
}

func (f *FileLoader) initS3() error {
	endpoint := "127.0.0.1:9000"
	accessKeyID := "9013HBZHIRHONH8ZIWL6"
	secretAccessKey := "gKIVgZaWAiuXbugPv9+dT4MKWsKlqxCyXFI+9ys+"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return errS3connetction
	}
	f.s3Client = minioClient
	return nil
}

func (f *FileLoader) initDBCon() error {
	dsn := "kalach:1234@/users?charset=utf8"
	db, err := sql.Open("mysql", dsn)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		fmt.Println(err)
		return errDBconnection
	}
	f.DBCon = db
	return nil
}

func (f *FileLoader) initAMQPCon() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return errAMQPconnetction
	}
	f.amqpCon = conn

	ch, err := conn.Channel()
	if err != nil {
		return errOpenChannel
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"grader", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return errAMQPDeclare
	}
	f.queue = q
	return nil
	// Don't forget
	// defer conn.Close()
}

func (f *FileLoader) saveUserTask(meta map[string]string) error {
	filename := meta["filename"]
	task := &Task{Course: "Golang", Task: "Grader", User: "kalach", Graded: false, Filename: filename}
	addAmqpTask(f.amqpCon, f.queue, task)
	addDBTask(f.DBCon, task)
	uploadS3(f.s3Client, task)
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

	n, err := s3Client.FPutObject("grader", filename, filePath, minio.PutObjectOptions{ContentType: "text/txt"})
	if err != nil {
		log.Fatalln(err, filePath, filename)
	}
	log.Printf("Successfully uploaded  S3 %s of size %d\n", filename, n)
}

func addDBTask(db *sql.DB, task *Task) {
	result, err := db.Exec(
		"INSERT INTO tasks (`course`, `task`, `user`, `filename`) VALUES (?, ?, ?, ?)",
		task.Course,
		task.Task,
		task.User,
		task.Filename,
	)
	if err != nil {
		log.Println("INSERT Error:", err)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error reading Last ID", err)
	}
	log.Printf("Successfully saved task to DB %s. ID = %d\n", task.Filename, lastID)
}

func addAmqpTask(amqpCon *amqp.Connection, queue amqp.Queue, task *Task) error {
	ch, err := amqpCon.Channel()
	if err != nil {
		return errOpenChannel
	}
	defer ch.Close()

	err = ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			UserId:      task.User,
			Type:        task.Course,
			Body:        []byte(task.Task),
		})

	if err != nil {
		return errAMQPMessagePublish
	}
	log.Printf("Successfully pushed %s to AMQP. Queue: %s\n", task.Filename, queue.Name)
	return nil
}
