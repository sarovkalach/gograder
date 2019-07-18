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

func TestUserByEmail(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	result := []string{"id", "email", "password"}
	rows := sqlmock.NewRows(result)
	email := "akalachov@mail.ru"
	badEmail := "bvqe8rbvwer9@mail.ru"
	password := "1234"
	id := 1
	rows.AddRow(id, email, password)

	mock.ExpectQuery("SELECT \\* FROM users WHERE email = \\?").WithArgs(email).WillReturnRows(rows)

	UserByEmail(db, email)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	// Non existing user
	rows = sqlmock.NewRows(result)
	mock.ExpectQuery("SELECT \\* FROM users WHERE email = \\?").WithArgs(badEmail).WillReturnRows(rows)
	UserByEmail(db, badEmail)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

func TestCreateUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	email := "bakalchov@mail.ru"
	password := "1234"
	result := []string{"id", "email", "password"}
	rows := sqlmock.NewRows(result)
	mock.ExpectQuery("SELECT \\* FROM users WHERE email = \\?").WithArgs(email).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO users (`email`, `password`) VALUES (?, ?)").
		WithArgs(email, password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	CreateUser(db, map[string]string{"email": email, "password": password})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
