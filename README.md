<h1>go-gpa</h1>
<h3>Golang Persistence API</h3>

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

```

ORM initialization:
```
// Global initialization go-gpa Engine
gpa.NewEngine(DB)
```

Api examples:

```go
// Save data to DB
id, err := gpa.From[User]().Insert(User{
    Name: "John",
})

// Update data in DB
err := gpa.From[User]().Update(User{ID: id, Name: "Doe"})

// Find all Data from DB
users, err := gpa.From[User]().FindAll()

// Find Data from DB by ID
user, err := gpa.From[User]().FindByID(id)

// Insert array of data
err := gpa.From[Document]().Inserts([]Document{
    {Text: "doc1", Title: "some text", Views: 11},
    {Text: "doc2", Title: "some text 2", Views: 67},
});

// Finding by conditions with filters
err := gpa.From[Document]().FindBy([]gpa.F{
    {FieldName: "title", Sign: gpa.Equal, Value: "doc1", Cond: gpa.OR},
    {FieldName: "views", Sign: gpa.More, Value: 10},
})

// Custom selecting with sqlx
_, err := gpa.From[Document]().Select("title = $1", "hellotext")

// Custom getting by sqlx
_, err := gpa.From[Document]().Get("views >= $1", 10)

// Remove Data from DB
err := gpa.From[User]().Delete(user.ID);
```


