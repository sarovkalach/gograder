package loader

import (
	"log"
	"os"
	"testing"
)

func TestUpload(t *testing.T) {
	f := NewFileLoader()
	filename := "/home/kalach/log.txt"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	// buckets, err := f.s3Client.ListBuckets()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// for _, bucket := range buckets {
	// 	fmt.Println(bucket)
	// }
	f.uploadS3(file)
}

func TestDBcon(t *testing.T) {
	NewFileLoader()

}

func TestDBSelect(t *testing.T) {
	f := NewFileLoader()
	showDBTasks(f.DBCon)
}

func TestDBInsert(t *testing.T) {
	f := NewFileLoader()
	task := &Task{Course: "Golang", Task: "Grader", User: "kalach", Graded: false, Filename: "test.txt"}
	addDBTask(f.DBCon, task)
}

func TestAmqpSend(t *testing.T) {
	f := NewFileLoader()
	task := &Task{Course: "Golang", Task: "Grader", User: "kalach", Graded: false, Filename: "test.txt"}
	addAmqpTask(f.amqpCon, task)
}
