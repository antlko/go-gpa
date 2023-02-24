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

func (u User) String() string {
	return fmt.Sprintf("{%d %s %+v}\n", u.ID, u.Name, u.GetRoles())
}

func (u User) GetRoles() []Role {
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

// GPAConfigure method where could be provided configs for GPAEntity, for ex. custom table name
func (d UserRole) GPAConfigure(o *gpa.Engine) {
	o.SetTableName(d, "user_roles")
}

func destroy() {
	DB := db.NewPGInstance(db.PGConfig{Host: host, Port: port, User: user, Password: password, DBName: dbname})
	if _, err := DB.Exec(`
	 DROP TABLE IF EXISTS user_roles;
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
	gpa.NewEngine(DB, gpa.Config{IsLazy: true})
	var err error

	gpa.From[UserRole]()
	gpa.From[Role]()

	err = gpa.From[User]().Inserts([]User{{Name: "Ann"}, {Name: "San"}, {Name: "Vi"}})
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
	roleUser, err := gpa.From[Role]().FindOneBy([]gpa.F{{FieldName: "name", Sign: gpa.Equal, Value: "USER"}}, nil)
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

	err = gpa.From[UserRole]().Inserts([]UserRole{
		{Role: roleAdmin.ID, User: 3},
		{Role: roleUser.ID, User: 3},
	})
	if err != nil {
		panic(err)
	}

	userWithRoles, err := gpa.From[User]().FindByID(1)
	if err != nil {
		panic(err)
	}
	println(fmt.Sprintf("%+v", userWithRoles.GetRoles()))

	usersWithRoles, err := gpa.From[User]().FindAll(nil)
	if err != nil {
		panic(err)
	}
	println(fmt.Sprintf("%+v\n", usersWithRoles))
}
