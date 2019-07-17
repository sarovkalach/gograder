package loader

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetTask(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	result := []string{"id", "status", "course", "name", "filename", "user_id"}
	rows := sqlmock.NewRows(result)
	rows.AddRow(1, 0, "Golang", "99hw/1", "main.go", 1)
	mock.ExpectQuery("SELECT \\* FROM tasks WHERE user_id = \\?").WithArgs(1).WillReturnRows(rows)
	records := GetTask(db, "1")
	fmt.Println(records)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	email := "akalchov@mail.ru"
	password := "1234"
	mock.ExpectExec("INSERT INTO users\\(email, password\\)").
		WithArgs(email, password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	CreateUser(db, map[string]string{"email": email, "password": password})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateUserDB(t *testing.T) {
	s := NewServer()
	email := "aakalchov@mail.ru"
	password := "1234"
	_, err := CreateUser(s.Uploader.DBCon, map[string]string{"email": email, "password": password})
	if err != nil {
		t.Error(err)
	}
}

// func TestUser(t *testing.T) {
// 	s := NewServer()
// 	u, err := UserByEmail(s.Uploader.DBCon, "akalachov@mail.ru")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Println(u)
// }

func TestUserByEmail(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	result := []string{"id", "email", "password"}
	rows := sqlmock.NewRows(result)
	email := "akalachov@mail.ru"
	password := "1234"
	rows.AddRow(1, email, password)

	mock.ExpectQuery("SELECT \\* FROM tasks WHERE user_id = \\?").WithArgs(1).WillReturnRows(rows)

	db.QueryRow("SELECT * FROM users where email = ?", email)
	// _, err := UserByEmail(db, email)
	// if err != nil {
	// 	t.Error(err)
	// }
}
