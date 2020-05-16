package main

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/saied74/toychat/pkg/forms"
	"golang.org/x/crypto/bcrypt"
)

const (
	superadmin = "superadmin"
)

type profile struct {
	firstMsg  string
	secondMsg string
	input     string
}

type profiles map[string]profile

type userModel struct {
	dB *sql.DB
}

var errDuplicateEmail = errors.New("models: duplicate email")

func main() {
	var name, email, password1, password2 string
	printIntro()
	var p = getProfile()

	// database password flag is required so we don't save it in the program.
	dsn := flag.String("dsn", "toy:password@/toychat?parseTime=true",
		"MySQL data source name")
	pw := flag.String("pw", "password", "database password is always required")
	flag.Parse()
	dbAddress := strings.Replace(*dsn, "password", *pw, 1)

	db, err := openDB(dbAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	m := userModel{dB: db}

	reader := bufio.NewReader(os.Stdin)

	name = p["name"].getInput(reader)
	email = p["email"].getInput(reader)
	for {
		password1 = p["password1"].getInput(reader)
		password2 = p["password2"].getInput(reader)
		if password1 == password2 {
			break
		}
	}

	data := url.Values{
		"name":     []string{name},
		"email":    []string{email},
		"password": []string{password1},
	}
	suForm := forms.NewForm(data)
	suForm.FieldRequired("name", "email", "password")
	suForm.MaxLength("name", 255)
	suForm.MaxLength("email", 255)
	suForm.MatchPattern("email", forms.EmailRX)
	suForm.MinLength("password", 10)
	if !suForm.Valid() {
		for key, value := range suForm.Errors {
			check := strings.Replace(value, ";", "\n", -1)
			fmt.Println(key, ":", check)
			fmt.Println("database did not update, run the program again")
			// os.Exit(0)
		}
	}
	err = m.insertUser("admins", name, email, password1, superadmin)
	if err != nil {
		log.Println(err)
		printBadConclusion()
	}
	printGoodConclusion()
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func getProfile() profiles {
	return profiles{
		"name": profile{
			firstMsg:  "Enter your first name followed by last name: ",
			secondMsg: "Your name is: ",
			input:     "",
		},
		"email": profile{
			firstMsg:  "Enter your email address: ",
			secondMsg: "Your email is: ",
			input:     "",
		},
		"password1": {
			firstMsg:  "Enter your password: ",
			secondMsg: "Your password is: ",
			input:     "",
		},
		"password2": {
			firstMsg:  "Enter your password for a second time: ",
			secondMsg: "Your second entry does does not match the first: ",
			input:     "",
		},
	}
}

func (p profile) getInput(reader *bufio.Reader) string {
	for {
		fmt.Println(p.firstMsg)
		input, _ := reader.ReadString('\n')
		fmt.Println(p.secondMsg, input)
		fmt.Println("Is that correct (Y or N):")
		test, _ := reader.ReadString('\n')
		test = strings.TrimSuffix(test, "\n")
		if test == "Y" {
			return strings.TrimSuffix(input, "\n")
		}
	}
}

func (m *userModel) insertUser(table, name, email, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err //note we are not returning any words so we can check for the error
	}
	stmt := `INSERT INTO ` + table +
		` (name, email, hashed_password, created, role) VALUES(?, ?, ?, UTC_TIMESTAMP(), ?)`

	_, err = m.dB.Exec(stmt, name, email, string(hashedPassword), role)
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 &&
				strings.Contains(mySQLError.Message, "users_uc_email") {
				return errDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func printIntro() {
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("+ This program is desgined to create the super admin for the +")
	fmt.Println("+ toychat application.  It must be run when the admin table  +")
	fmt.Println("+ does not have any superadmin roles in it.  If there is a   +")
	fmt.Println("+ superadmin role currently in the table, remove it using    +")
	fmt.Println("+ MySQL tools and then run this progream.                    +")
	fmt.Println("+ For now, I am just permitting one superadmin role and this +")
	fmt.Println("+ set up program.  I will change it in the future.           +")
	fmt.Println("+ You should run this program on a secure machine and in a   +")
	fmt.Println("+ and in a secure environment.                               +")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
}

func printGoodConclusion() {
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("+++++++++++++++++++++ CONGRATULATIONS ++++++++++++++++++++++++")
	fmt.Println("+ You have successfully updated the superadmin profile.      +")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
}

func printBadConclusion() {
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("++++++++++++++++++++++++++ SORRY +++++++++++++++++++++++++++++")
	fmt.Println("+ Inspect and clean the database admins table and try again  +")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
}
