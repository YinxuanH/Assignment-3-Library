package main

import (
	"fmt"
	"time"

	// mysql connector
	_ "github.com/go-sql-driver/mysql"
	sqlx "github.com/jmoiron/sqlx"
)

const (
	User     = "hyx"
	Password = "123456"
	DBName   = "ass3"

	// student status
	Normal = 0
	Suspend = 1

	// book status
	OnShelf = 0
	Borrowed = 1
	Removed = 2
)

type Library struct {
	db *sqlx.DB
}

func (lib *Library) CreateDB() {
	db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/", User, Password))
	if err != nil {
		panic(err)
	}
	mustExecute(db, []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS %s", DBName),
		fmt.Sprintf("CREATE DATABASE %s", DBName),
		fmt.Sprintf("USE %s", DBName),
	})
	lib.db = db
}

func (lib *Library) ConnectDB() {
	db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", User, Password, DBName))
	if err != nil {
		panic(err)
	}
	lib.db = db
}

func mustExecute(db *sqlx.DB, SQLs []string) {
	for _, s := range SQLs {
		_, err := db.Exec(s)
		if err != nil {
			panic(err)
		}
	}
}

// CreateTables create the tables in MySQL
func (lib *Library) CreateTables() error {
	mustExecute(lib.db, []string{
		"CREATE TABLE administrator (id CHAR(5) NOT NULL, name VARCHAR(32) NOT NULL, password VARCHAR(10) NOT NULL, PRIMARY KEY(id))",
		"CREATE TABLE student (id CHAR(10) NOT NULL, name VARCHAR(32) NOT NULL, password VARCHAR(10) NOT NULL, status SMALLINT NOT NULL, PRIMARY KEY(id))",
		"CREATE TABLE book (id INT NOT NULL, ISBN CHAR(13) NOT NULL, title VARCHAR(50) NOT NULL, author VARCHAR(32) NOT NULL, status SMALLINT NOT NULL, remark VARCHAR(100), PRIMARY KEY(id))",
		"CREATE TABLE borrow_return (student_id CHAR(10) NOT NULL, book_id INT NOT NULL, borrow_date DATE NOT NULL, due_date DATE NOT NULL, return_date DATE, extend_num SMALLINT NOT NULL, FOREIGN KEY(student_id) REFERENCES student(id), FOREIGN KEY(book_id) REFERENCES book(id))",
	})
	return nil
}

// InitializeDB initialize the Library Management System
func (lib *Library) InitializeDB() error {
	mustExecute(lib.db, []string{
		"INSERT INTO administrator (id, name, password) VALUES (\"00000\", \"Andy\", \"00000\"), (\"00001\", \"Harry\", \"12345\")",
		"INSERT INTO student (id, name, password, status) VALUES (\"0000000000\", \"Alina\", \"A0000\", 0), (\"0000000001\", \"Eira\", \"E1111\", 0), (\"0000000002\", \"Gaia\", \"G2222\", 0)",
		"INSERT INTO book (id, ISBN, title, author, status) VALUES (1, \"9787802204423\", \"Journey_to_the_West\", \"Wu_Chengen\", 1), (2, \"9787802204423\", \"Journey_to_the_West\", \"Wu_Chengen\", 0)",
		"INSERT INTO book (id, ISBN, title, author, status) VALUES (3, \"9787510136740\", \"Romance_of_the_Three_Kingdoms\", \"Luo_Guanzhong\", 0), (4, \"9787510136740\", \"Romance_of_the_Three_Kingdoms\", \"Luo_Guanzhong\", 0)",
		"INSERT INTO book (id, ISBN, title, author, status) VALUES (5, \"9787020015016\", \"The_Water_Margin\", \"Shi_Naian\", 1)",
		"INSERT INTO borrow_return (student_id, book_id, borrow_date, due_date, extend_num) VALUES (\"0000000000\", 1, '2020-04-03', '2020-04-17', 1), (\"0000000002\", 5, '2020-05-04', '2020-05-25', 2)",
	})
	return nil
}

// Login check the account ID and password
func (lib *Library) Login(usr_id, password string, account int) (bool, string, error) {
	var table string
	if account == 1 {
		table = "student"
	} else {
		table = "administrator"
	}
	rows, err := lib.db.Query(fmt.Sprintf("SELECT password, name FROM %s WHERE id = \"%s\"", table, usr_id))
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		var pwd, name string
		err = rows.Scan(&pwd, &name)
		if err != nil {
			panic(err)
		}
		if password == pwd {
			return true, name, nil
		}
	}
	return false, "", nil
}

// AddBook add a book into the library
func (lib *Library) AddBook(title, author, ISBN string) error {
	rows, err := lib.db.Query("SELECT MAX(id) FROM book")
	if err != nil {
		panic(err)
	}
	var id int
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			panic(err)
		}
	} else {
		id = 0
	}
	mustExecute(lib.db, []string{
		fmt.Sprintf("INSERT INTO book (id, ISBN, title, author, status) VALUES (%d, '%s', '%s', '%s', 0)", id + 1, ISBN, title, author),
	})
	fmt.Println("The new book has been successfully added to the library!")
	return nil
}

// RemoveBook remove a book from the library with explanation
func (lib *Library) RemoveBook(book_id int, explanation string) error {
	rows, err := lib.db.Query(fmt.Sprintf("SELECT status FROM book WHERE id = %d", book_id))
	if err != nil {
		panic(err)
	}
	var status int
	if rows.Next() {
		err = rows.Scan(&status)
		if err != nil {
			panic(err)
		}
		if status == OnShelf {
			mustExecute(lib.db, []string{
				fmt.Sprintf("UPDATE book SET status = %d, remark = \"%s\" WHERE id = %d", Removed, explanation, book_id),
			})
			fmt.Println("The old book has been successfully removed from the library!")
		} else if status == Borrowed {
			date := time.Now().Format("2006-01-02")
			rows, err = lib.db.Query(fmt.Sprintf("SELECT student_id FROM borrow_return WHERE book_id = %d AND return_date IS NULL", book_id))
			if err != nil {
				panic(err)
			}
			var student_id string
			if rows.Next() {
				err = rows.Scan(&student_id)
				if err != nil {
					panic(err)
				}
				mustExecute(lib.db, []string{
					fmt.Sprintf("UPDATE book SET status = %d, remark = \"%s\" WHERE id = %d", Removed, explanation, book_id),
					fmt.Sprintf("UPDATE borrow_return SET return_date = '%s' WHERE book_id = %d AND return_date IS NULL", date, book_id),
				})
				fmt.Println("The old book has been successfully removed from the library!")
			}
		} else {
			fmt.Println("The old book has already been removed from the library before!")
		}
	} else {
		fmt.Println("The book doesn't exit in the library!")
	}
	return nil
}

// AddStudent add a student account into the Library Management System
func (lib *Library) AddStudent(id, name, password string) error {
	rows, err := lib.db.Query(fmt.Sprintf("SELECT * FROM student WHERE id = \"%s\"", id))
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		fmt.Println("The ID of the student has already been added!")
	} else {
		mustExecute(lib.db, []string{
			fmt.Sprintf("INSERT INTO student (id, name, password, status) VALUES (\"%s\", \"%s\", \"%s\", %d)", id, name, password, Normal),
		})
		fmt.Println("The student account has been successfully added to the library Management System!")
	}
	return nil
}

// QueryBook query a book by title, author or ISBN
func (lib *Library) QueryBook(t, v string) error {
	var i int
	rows, err := lib.db.Query(fmt.Sprintf("SELECT id, title, author, ISBN, status FROM book WHERE %s = \"%s\" AND status <> %d", t, v, Removed))
	if err != nil {
		panic(err)
	}
	i = 0
	for rows.Next() {
		var title, author, ISBN, book_status string
		var id, status int
		i++
		err = rows.Scan(&id, &title, &author, &ISBN, &status)
		if err != nil {
			panic(err)
		}
		if status == 0 {
			book_status = "OnShelf"
		} else {
			book_status = "Borrowed"
		}
		fmt.Printf("Book_id: %d, Title: %s, Author: %s, ISBN: %s, Status: %s\n", id, title, author, ISBN, book_status)
	}
	if i == 0 {
		fmt.Println("No such book was found!")
	}
	return nil
}

// QueryBorrowHistory query the borrow history of a student account (exclude books has been borrowed and not returned yet)
func (lib *Library) QueryBorrowHistory(id string) error {
	rows, err := lib.db.Query(fmt.Sprintf("SELECT book_id, title, author, ISBN, borrow_date, return_date, extend_num FROM borrow_return, book WHERE book_id = id AND student_id = \"%s\" AND return_date IS NOT NULL", id))
	if err != nil {
		panic(err)
	}
	i := 0
	for rows.Next() {
		var title, author, ISBN, borrow_date, return_date string
		var book_id, extend_num int
		i++
		err = rows.Scan(&book_id, &title, &author, &ISBN, &borrow_date, &return_date, &extend_num)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Book_id: %d, Title: %s, Author: %s, ISBN: %s, Borrow_date: %s, Return_date: %s, Extend_num = %d\n", book_id, title, author, ISBN, borrow_date, return_date, extend_num)
	}
	if i == 0 {
		fmt.Println("No borrow history of this student account!")
	}
	return nil
}

// QueryBorrowedBook query the books a student has borrowed and not returned yet
// show borrow_date, due_date, extend_num
func (lib *Library) QueryBorrowedBook(id string) error {
	rows, err := lib.db.Query(fmt.Sprintf("SELECT book_id, title, author, ISBN, borrow_date, due_date, extend_num FROM borrow_return, book WHERE book_id = id AND student_id = \"%s\" AND status = %d", id, Borrowed))
	if err != nil {
		panic(err)
	}
	i := 0
	for rows.Next() {
		var title, author, ISBN, borrow_date, due_date string
		var book_id, extend_num int
		i++
		err = rows.Scan(&book_id, &title, &author, &ISBN, &borrow_date, &due_date, &extend_num)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Book_id: %d, Title: %s, Author: %s, ISBN: %s, Borrow_date: %s, Due_date: %s, Extend_num = %d\n", book_id, title, author, ISBN, borrow_date, due_date, extend_num)
	}
	if i == 0 {
		fmt.Println("No book has been borrowed and not returned yet of this student account!")
	}
	return nil
}

// CheckDeadline check the deadline of returning a borrowed book
// return due_date
func (lib *Library) CheckDeadline(id int) error {
	rows, err := lib.db.Query(fmt.Sprintf("SELECT due_date FROM borrow_return WHERE book_id = %d AND return_date IS NULL", id))
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		var due_date string
		err = rows.Scan(&due_date)
		if err != nil {
			panic(err)
		}
		fmt.Printf("The deadline of this book is %s\n", due_date)
		return nil
	}
	fmt.Println("No such book was found!")
	return nil
}

// CheckOverdueBook check if a student has any overdue books
// return the number of the overdue books
func (lib *Library) CheckOverdueBook(id string, op int) (int, error) {
	date := time.Now().Format("2006-01-02")
	rows, err := lib.db.Query(fmt.Sprintf("SELECT DISTINCT book_id, title, author, ISBN, due_date, extend_num FROM borrow_return, book WHERE student_id = \"%s\" AND book_id = id AND due_date < '%s' AND return_date IS NULL", id, date))
	if err != nil {
		panic(err)
	}
	i := 0
	for rows.Next() {
		var title, author, ISBN, due_date string
		var book_id, extend_num int
		i++
		err = rows.Scan(&book_id, &title, &author, &ISBN, &due_date, &extend_num)
		if err != nil {
			panic(err)
		}
		if op == 0 {
			fmt.Printf("Book_id: %d, Title: %s, Author: %s, ISBN: %s, Due_date: %s, Extend_num = %d\n", book_id, title, author, ISBN, due_date, extend_num)
		}
	}
	if i == 0 && op == 0 {
		fmt.Println("No overdue book of this student account!")
	}
	return i, nil
}

// CheckAccountStatus check if the account has more than 3 overdue books
func (lib *Library) CheckAccountStatus(id string) (bool, error) {
	num, err := lib.CheckOverdueBook(id, 1)
	if err != nil {
		panic(err)
	}
	if num > 3 {
		mustExecute(lib.db, []string{
			fmt.Sprintf("UPDATE student SET status = %d WHERE id = %s", Suspend, id),
		})
		return false, nil
	}
	mustExecute(lib.db, []string{
		fmt.Sprintf("UPDATE student SET status = %d WHERE id = \"%s\"", Normal, id),
	})
	return true, nil
}

// BorrowBook borrow a book from the library
// A book can be borrowed for a period of 7 days
func (lib *Library) BorrowBook(student_id, ISBN string) error {
	rows, err := lib.db.Query(fmt.Sprintf("SELECT id, status FROM book WHERE ISBN = \"%s\" AND status <> %d", ISBN, Removed))
	if err != nil {
		panic(err)
	}
	i := false
	for rows.Next() {
		var book_id, status int
		err = rows.Scan(&book_id, &status)
		if err != nil {
			panic(err)
		}
		if status == OnShelf {
			borrow_date := time.Now().Format("2006-01-02")
			due_date := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
			mustExecute(lib.db, []string{
				fmt.Sprintf("UPDATE book SET status = %d WHERE id = %d", Borrowed, book_id),
				fmt.Sprintf("INSERT INTO borrow_return (student_id, book_id, borrow_date, due_date, extend_num) VALUES (\"%s\", %d, \"%s\", \"%s\", 0)", student_id, book_id, borrow_date, due_date),
			})
			fmt.Println("The book has been successfully borrowed!")
			return nil
		}
		i = true
	}
	if i {
		fmt.Println("The book has all been borrowed by other students!")
	} else {
		fmt.Println("No such book was found!")
	}
	return nil	
}

// ReturnBook return a borrowed book to the library
func (lib *Library) ReturnBook(student_id string, book_id int) error {
	rows, err := lib.db.Query(fmt.Sprintf("SELECT * FROM borrow_return WHERE student_id = \"%s\" AND book_id = %d AND return_date IS NULL", student_id, book_id))
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		return_date := time.Now().Format("2006-01-02")
		mustExecute(lib.db, []string{
			fmt.Sprintf("UPDATE book SET status = %d WHERE id = %d", OnShelf, book_id),
			fmt.Sprintf("UPDATE borrow_return SET return_date = '%s' WHERE student_id = \"%s\" AND book_id = %d AND return_date IS NULL", return_date, student_id, book_id),
		})
		fmt.Println("The book has been successfully returned!")
		return nil
	}
	fmt.Println("No such book was found!")
	return nil
}

// ExtendDeadline extend the deadline of returning a borrowed book
// Extend the deadline for 7 days each time
// Refuse to extend if the deadline has been extended for 3 times or the book is overdue
func (lib *Library) ExtendDeadline(student_id string, book_id int) error {
	rows, err := lib.db.Query(fmt.Sprintf("SELECT due_date, extend_num FROM borrow_return WHERE student_id = \"%s\" AND book_id = %d AND return_date IS NULL", student_id, book_id))
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		date := time.Now().Format("2006-01-02")
		var due_date string
		var extend_num int
		err = rows.Scan(&due_date, &extend_num)
		if err != nil {
			panic(err)
		}
		if date < due_date {
			if extend_num < 3 {
				mustExecute(lib.db, []string{
					fmt.Sprintf("UPDATE borrow_return SET due_date = date_add(due_date, interval 7 day), extend_num = extend_num + 1 WHERE student_id = \"%s\" AND book_id = %d AND return_date IS NULL", student_id, book_id),
				})
				fmt.Println("The deadline of the book has been successfully extended!")
			} else {
				fmt.Println("Refuse to extend because the deadline of the book has been extended to many times!")
			}
		} else {
			fmt.Println("Refuse to extend because the book is overdue!")
		}
		return nil
	}
	fmt.Println("No such book was found!")
	return nil
}

func main() {
	fmt.Println("Welcome to the Library Management System!")
	lib := Library{}
	lib.CreateDB()
	lib.ConnectDB()
	lib.CreateTables()
	lib.InitializeDB()
	for true {
		var account int
		fmt.Println("To log in, please select your account type:")
		fmt.Println("1.Student 2.Administrator 0.Exit")
		fmt.Scanln(&account)
		if account == 0 {
			break
		}
		if account == 1 || account == 2 {
			var usr_id, password string
			fmt.Println("Please enter your account ID:")
			fmt.Scanln(&usr_id)
			fmt.Println("Please enter your password:")
			fmt.Scanln(&password)
			login, name, err := lib.Login(usr_id, password, account)
			if err != nil {
				panic(err)
			}
			if !login {
				fmt.Println("Sorry, the account ID with this password was not found!")
				continue
			}
			fmt.Printf("Hello, %s!\n", name)
			for true {
				fmt.Println("Please select the function you need:")
				var funct int
				if account == 1 {
					fmt.Println("1.Query a book by title, author or ISBN")
					fmt.Println("2.Borrow a book from the library")
					fmt.Println("3.Return a book to the library")
					fmt.Println("4.Extend the deadline of returning a book you borrowed")
					fmt.Println("5.Query the books you have borrowed but not returned yet") 
					fmt.Println("6.Query the borrow history of your account")
					fmt.Println("0.Log out")
					fmt.Scanln(&funct)
					if funct == 0 {
						fmt.Printf("Bye, %s!\n", name)
						break
					}
					status, err := lib.CheckAccountStatus(usr_id)
					if err != nil {
						panic(err)
					}
					switch funct {
						case 1: {
							var t, v string
							var i int
							fmt.Println("Query a book by?")
							fmt.Println("1.Title 2.Author 3.ISBN")
							fmt.Scanln(&i)
							if i > 3 || i < 1 {
								fmt.Println("Undefined type!")
								break
							}
							switch i {
								case 1: t = "title"
								case 2: t = "author"
								case 3: t = "ISBN"
							}
							fmt.Printf("Please enter the %s of the book:\n", t)
							fmt.Scanln(&v)
							lib.QueryBook(t, v)
						}
						case 2: {
							if status {
								var ISBN string
								fmt.Println("Please enter the ISBN of the book you want to borrow:")
								fmt.Scanln(&ISBN)
								lib.BorrowBook(usr_id, ISBN)
							} else {
								fmt.Println("You have more than 3 books overdue. Please return overdue books first!")
							}
						}
						case 3: {
							var book_id int
							fmt.Println("Please enter the ID of the book you want to return:")
							fmt.Scanln(&book_id)
							lib.ReturnBook(usr_id, book_id)
						}
						case 4: {
							var book_id int
							fmt.Println("Please enter the ID of the book you want to extend the deadline:")
							fmt.Scanln(&book_id)
							lib.ExtendDeadline(usr_id, book_id)
						}
						case 5: lib.QueryBorrowedBook(usr_id)
						case 6: lib.QueryBorrowHistory(usr_id)
						default: fmt.Println("Undefined type. Please try again!")
					}				
				} else {
					fmt.Println("1.Add a book to the library")
					fmt.Println("2.Remove a book from the library with explanation")
					fmt.Println("3.Add a student account")
					fmt.Println("4.Query a book by title, author or ISBN")
					fmt.Println("5.Query the borrow history of a student account")
					fmt.Println("6.Query the books a student has borrowed and not returned yet")
					fmt.Println("7.Check the deadline of returning a borrowed book")
					fmt.Println("8.Check if a student has any overdue books")
					fmt.Println("0.Log out")
					fmt.Scanln(&funct)
					if funct == 0 {
						fmt.Printf("Bye, %s!\n", name)
						break
					}
					switch funct {
						case 1: {
							var title, author, ISBN string
							fmt.Println("Please enter the title of the new book:")
							fmt.Scanln(&title)
							fmt.Println("Please enter the author of the new book:")
							fmt.Scanln(&author)
							fmt.Println("Please enter the ISBN of the new book:")
							fmt.Scanln(&ISBN)
							lib.AddBook(title, author, ISBN)
						}
						case 2: {
							var book_id int
							var explanation string
							fmt.Println("Please enter the ID of the book to be removed:")
							fmt.Scanln(&book_id)
							fmt.Println("Please enter the explanation:")
							fmt.Scanln(&explanation)
							lib.RemoveBook(book_id, explanation)
						}
						case 3: {
							var student_id, student_name, pwd string
							fmt.Println("Please enter the ID of the student:")
							fmt.Scanln(&student_id)
							fmt.Println("Please enter the name of the student:")
							fmt.Scanln(&student_name)
							fmt.Println("Please enter the password of the account:")
							fmt.Scanln(&pwd)
							lib.AddStudent(student_id, student_name, pwd) 
						}
						case 4: {
							var t, v string
							var i int
							fmt.Println("Query a book by?")
							fmt.Println("1.Title 2.Author 3.ISBN")
							fmt.Scanln(&i)
							if i > 3 || i < 1 {
								fmt.Println("Undefined type!")
								break
							}
							switch i {
								case 1: t = "title"
								case 2: t = "author"
								case 3: t = "ISBN"
							}
							fmt.Printf("Please enter the %s of the book:\n", t)
							fmt.Scanln(&v)
							lib.QueryBook(t, v)
						}
						case 5: {
							var student_id string
							fmt.Println("Please enter the ID of the student you want to query:")
							fmt.Scanln(&student_id)
							lib.QueryBorrowHistory(student_id)
						}
						case 6: {
							var student_id string
							fmt.Println("Please enter the ID of the student you want to query:")
							fmt.Scanln(&student_id)
							lib.QueryBorrowedBook(student_id)
						}
						case 7: {
							var book_id int
							fmt.Println("Please enter the ID of the borrowed book you want to check:")
							fmt.Scanln(&book_id)
							lib.CheckDeadline(book_id)
						}
						case 8: {
							var student_id string
							fmt.Println("Please enter the ID of the student you want to check:")
							fmt.Scanln(&student_id)
							lib.CheckOverdueBook(student_id, 0)
						}
						default: fmt.Println("Undefined type. Please try again!")
					}
				}
			}
		} else {
			fmt.Println("Undefined type. Please try again!")
			continue
		}
	}
	fmt.Println("Exit the Library Management System!")
}