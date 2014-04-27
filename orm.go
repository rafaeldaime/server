package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
)

type Orm struct {
	// An anonymous field, all fields of DbMap are promoted into Orm
	gorp.DbMap
}

var orm *Orm = nil

const (
	dbname = "irado"
	dbuser = "irado"
	dbpass = "iradopassword"
)

func OrmMiddleware(c martini.Context) {
	// Inject Orm when requested by haWndlers
	//c.MapTo(orm, (*Orm)(nil))
}

func init() {
	log.Println("Connecting to database...")
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("mysql", dbuser+":"+dbpass+"@/"+dbname)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	dialect := gorp.MySQLDialect{"InnoDB", "UTF8"}

	// construct a gorp DbMap
	gorp := gorp.DbMap{Db: db, Dialect: dialect}
	gorp.TraceOn("[gorp]", log.New(os.Stdout, "Irado!:", log.Lmicroseconds))

	// Adding to local vairable
	orm = &Orm{gorp}
	log.Println("Database connected!")
}

func (orm Orm) GetOrCreateUser(t *Token, profile *Profile) (*User, error) {
	return nil, nil
}

type User struct {
}

type Profile struct {
	ProfileId      string    `db:"profileid" json:"profileid"`           // *PK max: 40
	Source         string    `db:"source" json:"source"`                 // *PK enum: facebook, google, twitter
	UserId         string    `db:"userid" json:"userid"`                 // *FK max: 10
	UserName       string    `db:"username" json:"username"`             // max: 20
	Email          string    `db:"email" json:"email"`                   // max: 40
	FullName       string    `db:"fullname" json:"fullname"`             // max: 40
	Gender         string    `db:"gender" json:"gender"`                 // enum: male, female
	ProfileUrl     string    `db:"profileurl" json:"profileurl"`         // max: 100
	ImageUrl       string    `db:"imageurl" json:"imageurl"`             // max: 100
	Language       string    `db:"language" json:"language"`             // max: 10, default: pt_BR
	Verified       bool      `db:"verified" json:"verified"`             // default: 0
	FirstName      string    `db:"firstname" json:"firstname"`           // max: 20
	LastName       string    `db:"lastname" json:"lastname"`             // max: 20
	LastUpdateTime time.Time `db:"lastupdatetime" json:"lastupdatetime"` // default: CURRENT_TIMESTAMP
	AccessToken    string    `db:"accesstoken" json:"accesstoken"`       // max: 100
	RefreshToken   string    `db:"refreshtoken" json:"refreshtoken"`     // max: 100
	Expiry         time.Time `db:"expiry" json:"expiry"`
	Scope          string    `db:"scope" json:"scope"` // max: 40
}
