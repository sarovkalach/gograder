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
	task := &Task{Course: "Golang", Name: "hw9", Status: 0, Filename: "hw9.go", UserID: 1}
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
	task := &Task{Course: "Golang", Name: "hw9", Status: 0, Filename: "hw9.go", UserID: 1}
	addDBTask(f.DBCon, task)
}

// func TestDBInsertTask(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
// 	}
// 	defer db.Close()

// 	task := &Task{Course: "Golang", Name: "hw9", Status: 0, Filename: "hw9.go", UserID: 1}
// 	mock.ExpectExec("INSERT INTO tasks (`status`, `course`, `name`, `filename`, `user_id`) VALUES (?, ?, ?, ?,?)").WithArgs(
// 		task.Status,
// 		task.Course,
// 		task.Name,
// 		task.Filename,
// 		task.UserID,
// 	)
// 	addDBTask(db, task)
// 	// mock.ExpectBegin()
// 	// mock.ExpectExec("UPDATE products").WillReturnResult(sqlmock.NewResult(1, 1))
// 	// mock.ExpectExec("INSERT INTO product_viewers").WithArgs(2, 3).WillReturnResult(sqlmock.NewResult(1, 1))
// 	// mock.ExpectCommit()

// }

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
		Encrypt(user.Password),
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
		err = rows.Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken)
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
		err = rows.Scan(&task.ID, &task.Status, &task.Course, &task.Name, &task.Filename, &task.S3BucketName, &task.UserID)
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
	task := &Task{Course: "Golang", Name: "hw9", Status: 0, Filename: "hw9.go", UserID: 1}
	addAmqpTask(f.amqpCon, f.queue, task)
}
