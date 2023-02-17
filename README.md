<h1>go-gpa</h1>
<h3>GPA - Golang Persistence API / Golang ORM / Golang Database / Generics </h3>

<h3>Working with:</h3><h4>Postgres</h2>

<h3>Idea:</h3>

The idea of this library is to make developers life much more easier
and provide more quick development, at least for starting projects.

If you not have initialized table previously  `orm.From[Entity]()` method
will initialize table automatically.

<h3>Usage:</h3>

Entity initializations:
```go
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
```

ORM initialization:
```
// Global initialization go-gpa Engine
gpa.NewEngine(DB)

// If used lazy fetch type, 
// structures should be initialized manually
// gpa.From[UserRole]()
// gpa.From[Role]()
```

Api examples:

```go
// Save data to DB
err := gpa.From[User]().Insert(User{
    Name: "John",
})

// Update data in DB
err := gpa.From[User]().Update(User{ID: id, Name: "Doe"})

// Find all Data from DB
users, err := gpa.From[User]().FindAll(nil) // could be added pagination

// Find Data from DB by ID
user, err := gpa.From[User]().FindByID(id)

// Insert array of data
err := gpa.From[Document]().Inserts([]Document{
    {Text: "doc1", Title: "some text", Views: 11},
    {Text: "doc2", Title: "some text 2", Views: 67},
});

// Finding by conditions with filters
docs, err := gpa.From[Document]().FindBy([]gpa.F{
    {FieldName: "title", Sign: gpa.Equal, Value: "doc1", Cond: gpa.OR},
    {FieldName: "views", Sign: gpa.More, Value: 10},
}, nil)

// Find One By custom filter
roleAdmin, err := gpa.From[Role]().FindOneBy([]gpa.F{{FieldName: "name", Sign: gpa.Equal, Value: "ADMIN"}}, nil)

// Custom selecting with sqlx
_, err := gpa.From[Document]().Select("title = $1", "hellotext")

// Custom getting by sqlx
_, err := gpa.From[Document]().Get("views >= $1", 10)

// Remove Data from DB
err := gpa.From[User]().Delete(user.ID);

// Get Lazy User with Roles
userWithRoles, err := gpa.From[User]().FindByID(1)
userWithRoles.GetRoles()
```


