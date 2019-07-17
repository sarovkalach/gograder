package loader

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetUserTasks(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	result := []string{"id", "status", "course", "name", "filename", "user_id"}
	rows := sqlmock.NewRows(result)
	rows.AddRow(1, 0, "Golang", "99hw/1", "main.go", 1)
	mock.ExpectQuery("SELECT \\* FROM tasks WHERE user_id = \\?").WithArgs(1).WillReturnRows(rows)
	records := GetUserTasks(db, "1")
	fmt.Println(records)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
