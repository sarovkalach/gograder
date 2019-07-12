package loader

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	// _ "github.com/go-sql-driver/mysql"
	// "github.com/gorilla/mux"
	// "github.com/gorilla/mux"
	"github.com/minio/minio-go/v6"
	"github.com/streadway/amqp"
)

type FileLoader struct {
	s3Client *minio.Client
	// Router     *mux.Router
	DBCon   *sql.DB
	amqpCon *amqp.Connection
}

type Task struct {
	ID       int
	Course   string
	Task     string
	User     string
	Graded   bool
	Filename string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func showDBTasks(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		log.Println("SELECT Error:", err)
	}
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Course, &task.Task, &task.User, &task.Graded)
		if err != nil {
			log.Println("SELECT Error Read:", err)
		}
		fmt.Println(task)
	}
	rows.Close()
}

func addDBTask(db *sql.DB, task *Task) {
	result, err := db.Exec(
		"INSERT INTO tasks (`course`, `task`, `user`) VALUES (?, ?, ?)",
		task.Course,
		task.Task,
		task.User,
	)
	if err != nil {
		log.Println("INSERT Error:", err)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error reading Last ID", err)
	}
	fmt.Println("Last ID", lastID)
}

func addAmqpTask(amqpCon *amqp.Connection, task *Task) {
	ch, err := amqpCon.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// body := "Hello World!"
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			UserId:      task.User,
			Type:        task.Course,
			Body:        []byte(task.Task),
		})
	failOnError(err, "Failed to publish a message")
}

func NewFileLoader() *FileLoader {
	f := &FileLoader{}
	f.initS3()
	f.initDBCon()
	f.initQueueCon()
	return f
}

func (f *FileLoader) initS3() {
	endpoint := "127.0.0.1:9000"
	accessKeyID := "9013HBZHIRHONH8ZIWL6"
	secretAccessKey := "gKIVgZaWAiuXbugPv9+dT4MKWsKlqxCyXFI+9ys+"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln("Failed to connect to S3", err)
	}
	f.s3Client = minioClient
}

func (f *FileLoader) initDBCon() {
	dsn := "root:1234@/tasks?charset=utf8"
	db, err := sql.Open("mysql", dsn)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		log.Fatalln("Failed to connect to MySQL", err)
	}
	f.DBCon = db
}

func (f *FileLoader) initQueueCon() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	f.amqpCon = conn
	// Don't forget
	// defer conn.Close()
}

func (f *FileLoader) saveUserTask() {

}

func (f *FileLoader) uploadS3(file *os.File) {
	filename := "/home/kalach/log.txt"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	n, err := f.s3Client.FPutObject("grader", "log.txt", filename, minio.PutObjectOptions{ContentType: "text/txt"})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Successfully uploaded %s of size %d\n", "log.txt", n)
}

var uploadFormTmpl = []byte(`
<html>
	<body>
	<form action="/upload" method="post" enctype="multipart/form-data">
		Image: <input type="file" name="my_file">
		<input type="submit" value="Upload">
	</form>
	</body>
</html>
`)

type Server struct {
	Router   *mux.Router
	Uploader *FileLoader
}

func NewServer() *Server {
	s := &Server{Router: mux.NewRouter(), Uploader: NewFileLoader()}
	s.initRoutes()
	return s
}

func (a *Server) initRoutes() {
	a.Router.HandleFunc("/", a.uploadFile)
}

func (a *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	w.Write(uploadFormTmpl)
}
