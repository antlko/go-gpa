package main

import (
	"fmt"
	"github.com/antlko/go-gpa/db"
	"github.com/antlko/go-gpa/gpa"
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
	ID    int64   `db:"id"`
	Name  string  `db:"name"`
	Roles *[]Role `join:"user_roles" fetchBy:"role_id" mappedBy:"user_id" fetch:"lazy"`
}

func (u User) GetRoles() []Role {
	if u.Roles != nil {
		return *u.Roles
	}
	return []Role{}
}

func (u User) GetTypes() []Role {
	if u.Roles != nil {
		return *u.Roles
	}
	return []Role{}
}

// Role GPA entity
type Role struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// UserRole GPA entity
type UserRole struct {
	Role interface{} `db:"role_id" join:"roles" mappedBy:"id" fetch:"lazy"`
	User interface{} `db:"user_id" join:"users" mappedBy:"id"`
}

// GPAConfigure method where should be provided configs for GPAEntity
func (d UserRole) GPAConfigure(o *gpa.Engine) {
	o.SetTableName(d, "user_roles")
}

type UserType struct {
	ID          int64  `db:"id"`
	IsAvailable bool   `db:"is_available"`
	TypeName    string `db:"type_name"`
}

// GPAConfigure method where should be provided configs for GPAEntity
func (d UserType) GPAConfigure(o *gpa.Engine) {
	o.SetTableName(d, "user_types")
}

func destroy() {
	DB := db.NewPGInstance(db.PGConfig{Host: host, Port: port, User: user, Password: password, DBName: dbname})
	if _, err := DB.Exec(`
	 DROP TABLE IF EXISTS user_roles;
	 DROP TABLE IF EXISTS user_types;
	 DROP TABLE IF EXISTS roles;
	 DROP TABLE IF EXISTS users;
	`); err != nil {
		panic(err)
	}
}

func main() {
	defer destroy()

	DB := db.NewPGInstance(db.PGConfig{Host: host, Port: port, User: user, Password: password, DBName: dbname})

	// Global initialization go-gpa Engine
	gpa.NewEngine(DB)
	var err error

	gpa.From[UserRole]()
	gpa.From[UserType]()
	gpa.From[Role]()

	err = gpa.From[User]().Inserts([]User{{Name: "Ann"}, {Name: "San"}})
	if err != nil {
		panic(err)
	}
	err = gpa.From[Role]().Inserts([]Role{{Name: "ADMIN"}, {Name: "USER"}})
	if err != nil {
		panic(err)
	}

	_, err = gpa.From[UserRole]().FindAll(nil)
	if err != nil {
		panic(err)
	}

	roleAdmin, err := gpa.From[Role]().FindOneBy([]gpa.F{{FieldName: "name", Sign: gpa.Equal, Value: "ADMIN"}}, nil)
	if err != nil {
		panic(err)
	}

	err = gpa.From[UserRole]().Insert(UserRole{
		Role: roleAdmin.ID,
		User: 1,
	})
	if err != nil {
		panic(err)
	}

	userWithRoles, err := gpa.From[User]().FindByID(1)
	if err != nil {
		panic(err)
	}
	println(fmt.Sprintf("%+v", userWithRoles.GetRoles()))
}
