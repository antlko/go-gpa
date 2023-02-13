package main

import (
	"database/sql"
	"fmt"
	"github.com/antlko/go-gpa/db"
	"github.com/antlko/go-gpa/gpa"
	"github.com/jmoiron/sqlx"
)

const (
	host     = "localhost"
	port     = "5432"
	user     = "postgres"
	password = "example"
	dbname   = "pg_experiments"
)

// User GPA entity
type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// GPAConfigure method where should be provided configs for GPAEntity
func (d User) GPAConfigure(o *gpa.Engine) {
	o.SetTableName(d, "users")
}

// Document GPA entity
type Document struct {
	ID       int64        `db:"id"`
	Title    string       `db:"title"`
	Text     string       `db:"text"`
	Views    int64        `db:"views"`
	TimeSlot sql.NullTime `db:"time_slot"`
}

/*
	If config method was not implemented and was not defined SetTableName(...) configuration,
	then engine will try to define by pluralize library table in the DB:

	F.e.: Entity -> entities, User -> users, Document -> documents
*/
//func (d Document) GPAConfigure(o *gpa.Engine) {
//	o.SetTableName(d, "documents")
//}

var DB *sqlx.DB

func init() {
	DB = db.NewPGInstance(db.PGConfig{Host: host, Port: port, User: user, Password: password, DBName: dbname})

	// Global initialization go-gpa Engine
	gpa.NewEngine(DB)

	// Set up testing data
	if _, err := DB.Exec(`
      DROP TABLE IF EXISTS users;
      DROP TABLE IF EXISTS documents;

      CREATE TABLE IF NOT EXISTS users 
      (
          id serial,
          name text
      );
   `); err != nil {
		panic(err)
	}
}

func destroy() {
	if _, err := DB.Exec(`
     DROP TABLE documents;
     DROP TABLE users;
   `); err != nil {
		panic(err)
	}
}

func main() {
	defer destroy()

	// Save data to DB
	id, err := gpa.From[User]().Insert(User{
		Name: "John",
	})
	if err != nil {
		panic(err)
	}

	// Update entity
	if err := gpa.From[User]().Update(User{
		ID:   id,
		Name: "Doe",
	}); err != nil {
		panic(err)
	}

	// Find all Data from DB
	users, err := gpa.From[User]().FindAll()
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	fmt.Println(len(users))

	// Find Data from DB by ID
	user, err := gpa.From[User]().FindByID(id)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	// Insert array of data
	if err := gpa.From[Document]().Inserts([]Document{
		{Text: "doc1", Title: "some text", Views: 11},
		{Text: "doc2", Title: "some text 2", Views: 67},
	}); err != nil {
		panic(err)
	}

	// Finding by conditions with filters
	docs, err := gpa.From[Document]().FindBy([]gpa.F{
		{FieldName: "title", Sign: gpa.Equal, Value: "doc1", Cond: gpa.OR},
		{FieldName: "views", Sign: gpa.More, Value: 10},
	})
	if err != nil {
		panic(err)
	}
	if len(docs) != 2 {
		panic("wrong orm work")
	}

	// Custom selecting with sqlx
	_, err = gpa.From[Document]().Select("title = $1", "hellotext")
	if err != nil {
		panic(err)
	}

	// Custom getting by sqlx
	_, err = gpa.From[Document]().Get("views >= $1", 10)
	if err != nil {
		panic(err)
	}

	// Remove Data from DB
	if err := gpa.From[User]().Delete(user.ID); err != nil {
		panic(err)
	}
}
