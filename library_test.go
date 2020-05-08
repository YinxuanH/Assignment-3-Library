package main

import (
	"testing"
	"fmt"
)

func TestCreateTables(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	err := lib.CreateTables()
	if err != nil {
		t.Errorf("can't create tables")
	}
}

func TestInitializeDB(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	err := lib.InitializeDB()
	if err != nil {
		t.Errorf("can't initialize the Library Management System")
	}
}

func TestLogin(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	tables := []struct {
		usr_id, password, name string
		account int
		login bool
	}{
		{"0000000000", "A0000", "Alina", 1, true},
		{"0000000000", "B0000", "", 1, false},
		{"00000", "00000", "Andy", 2, true},
		{"00000", "10000", "", 2, false},
	}
	for _, table := range tables {
		login, name, err := lib.Login(table.usr_id, table.password, table.account)
		if err != nil {
			t.Errorf("can't log in the library Management System")
		}
		if table.login != login || table.name != name {
			t.Errorf("Login was incorrect, got: %t, %s, want: %t, %s", login, name, table.login, table.name)
		}
	}
}

func TestAddBook(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	title := "Romance_of_the_Three_Kingdoms"
	author := "Luo_Guanzhong"
	ISBN := "9787510136740"
	err := lib.AddBook(title, author, ISBN)
	if err != nil {
		t.Errorf("can't add a book into the library")
	}
	rows, err := lib.db.Query(fmt.Sprintf("SELECT title, author, ISBN FROM book WHERE id = %d", 6))
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		var _title, _author, _ISBN string
		err = rows.Scan(&_title, &_author, &_ISBN)
		if err != nil {
			panic(err)
		}
		if title != _title || author != _author || ISBN != _ISBN {
			t.Errorf("AddBook was incorrect, got: %s, %s, %s, want: %s, %s, %s", _title, _author, _ISBN, title, author, ISBN)
		}
	} else {
		t.Errorf("can't add a book into the library")
	}
}

func TestRemoveBook(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	tables := []struct {
		book_id, status int
		explanation string
	}{
		{1, 2, "The book is lost."},
		{2, 2, "The book is damaged."},
	}
	for _, table := range tables {
		err := lib.RemoveBook(table.book_id, table.explanation)
		if err != nil {
			t.Errorf("can't remove a book from the library with explanation")
		}
		rows, err := lib.db.Query(fmt.Sprintf("SELECT status FROM book WHERE id = %d", table.book_id))
		if err != nil {
			panic(err)
		}
		if rows.Next() {
			var status int
			err = rows.Scan(&status)
			if err != nil {
				panic(err)
			}
			if table.status != status {
				t.Errorf("RemoveBook was incorrect, got: %d want: %d", status, table.status)
			}
		}
	}
}

func TestAddStudent(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	id := "0000000000"
	name := "Alina"
	password := "A0000"
	err := lib.AddStudent(id, name, password)
	if err != nil {
		t.Errorf("can't add a student account into the Library Management System")
	}
	login, _name, err := lib.Login(id, password, 1)
	if login == false || name != _name{
		t.Errorf("AddStudent was incorrect")
	}
}

func TestQueryBook(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	tables := []struct {
		t, v string
	}{
		{"title", "Journey_to_the_West"},
		{"author", "Jennie"},
		{"ISBN", "9787020015016"},
	}
	for _, table := range tables {
		err := lib.QueryBook(table.t, table.v)
		if err != nil {
			t.Errorf("can't query a book by title, author or ISBN")
		}
	}
}

func TestQueryBorrowHistory(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	student_id := "0000000000"
	book_id := 1
	lib.ReturnBook(student_id, book_id)
	err := lib.QueryBorrowHistory(student_id)
	if err != nil {
		t.Errorf("can't query the borrow history of a student account")
	}
}

func TestQueryBorrowedBook(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	tables := []struct {
		student_id string
	}{
		{"0000000000"},
		{"0000000001"},
	}
	for _, table := range tables {
		err := lib.QueryBorrowedBook(table.student_id)
		if err != nil {
			t.Errorf("can't query the books a student has borrowed and not returned yet")
		}
	}
}

func TestCheckDeadline(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	tables := []struct {
		book_id int
	}{
		{3},
		{5},
	}
	for _, table := range tables {
		err := lib.CheckDeadline(table.book_id)
		if err != nil {
			t.Errorf("can't check the deadline of returning a borrowed book")
		}
	}
}

func TestCheckOverdueBook(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	student_id := "0000000000"
	num := 1
	_num, err := lib.CheckOverdueBook(student_id, 0)
	if err != nil {
		t.Errorf("can't check if a student has any overdue books")
	}
	if num != _num {
		t.Errorf("CheckOverdueBook was incorrect, got: %d, want: %d", _num, num)
	}
}

func TestCheckAccountStatus(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	student_id := "0000000000"
	status := true
	_status, err := lib.CheckAccountStatus(student_id)
	if err != nil {
		t.Errorf("can't check if the account has more than 3 overdue books")
	}
	if status != _status {
		t.Errorf("CheckAccountStatus was incorrect, got: %t, want: %t", _status, status)
	}
}

func TestBorrowBook(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	tables := []struct {
		student_id, ISBN string
	}{
		{"0000000000", "9787020015016"},
		{"0000000001", "9787802204423"},
	}
	for _, table := range tables {
		err := lib.BorrowBook(table.student_id, table.ISBN)
		if err != nil {
			t.Errorf("can't borrow a book from the library")
		}
	}
}

func TestReturnBook(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	tables := []struct {
		student_id string
		book_id int
	}{
		{"0000000000", 1},
		{"0000000001", 2},
	}
	for _, table := range tables {
		err := lib.ReturnBook(table.student_id, table.book_id)
		if err != nil {
			t.Errorf("can't return a borrowed book to the library")
		}
	}
}

func TestExtendDeadline(t *testing.T) {
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	tables := []struct {
		student_id string
		book_id int
	}{
		{"0000000000", 1},
		{"0000000002", 5},
		{"0000000002", 5},
	}
	for _, table := range tables {
		err := lib.ExtendDeadline(table.student_id, table.book_id)
		if err != nil {
			t.Errorf("can't extend the deadline of returning a borrowed book")
		}
	}
}