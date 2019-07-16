package loader

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestUpload(t *testing.T) {
	f, err := NewFileLoader()
	if err != nil {
		t.Error(err)
	}
	filename := "/home/kalach/log.txt"
	file, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	file.Close()
	task := &Task{Course: "Golang", Task: "Grader", User: "kalach", Graded: false, Filename: "test.txt"}
	uploadS3(f.s3Client, task)
}

func TestDBcon(t *testing.T) {
	NewFileLoader()
}

func TestDBInsertTask(t *testing.T) {
	f, err := NewFileLoader()
	if err != nil {
		t.Error(err)
	}
	task := &Task{Course: "Golang", Task: "Grader", User: "kalach", Graded: false, Filename: "test.txt"}
	addDBTask(f.DBCon, task)
}

func TestDBInsertUser(t *testing.T) {
	f, err := NewFileLoader()
	if err != nil {
		t.Error(err)
	}
	// user := &User{Email: "akalachov@mail.ru", Password: "fcwecvervev", LastName: "Alex", FirstName: "Kalachov"}
	user := &User{Email: "akalachov@mail.ru", Password: "1234"}
	result, err := f.DBCon.Exec(
		// "INSERT INTO users (`email`, `password`, `last_name`, `first_name`) VALUES (?, ?, ?, ?)",
		"INSERT INTO users (`email`, `password`) VALUES (?, ?)",
		user.Email,
		user.Password,
		// user.LastName,
		// user.FirstName,
	)
	if err != nil {
		t.Error(err)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("LAST ID = ", lastID)
}

func TestDBSelectUsers(t *testing.T) {
	f, err := NewFileLoader()
	if err != nil {
		t.Error(err)
	}

	rows, err := f.DBCon.Query("SELECT * FROM users")
	if err != nil {
		t.Error(err)
	}
	for rows.Next() {
		user := &User{}
		err = rows.Scan(&user.ID, &user.Email, &user.Password)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(user)
	}
	rows.Close()

}

func TestDBSelectTasks(t *testing.T) {
	f, err := NewFileLoader()
	if err != nil {
		t.Error(err)
	}

	rows, err := f.DBCon.Query("SELECT * FROM tasks")
	if err != nil {
		t.Error(err)
	}
	tasks := make([]Task, 0, 16)
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Graded, &task.Course, &task.Task, &task.User, &task.Filename)
		if err != nil {
			log.Println("SELECT Error Read:", err)
		}
		fmt.Println(task)
		tasks = append(tasks, *task)
	}
	rows.Close()

}

func TestAmqpSend(t *testing.T) {
	f, err := NewFileLoader()
	if err != nil {
		t.Error(err)
	}
	task := &Task{Course: "Golang", Task: "Grader", User: "kalach", Graded: false, Filename: "test.txt"}
	addAmqpTask(f.amqpCon, f.queue, task)
}
