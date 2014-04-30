package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
)

type DB interface {
	// An anonymous field, all fields of DbMap are promoted into Orm
	//gorp.DbMap
	gorp.SqlExecutor
	// It's the iterface for querying DB
}

// The only one instance of db
var db DB

const (
	dbname = "database"
	dbuser = "app"
	dbpass = "SecretPassword!"
)

func OrmMiddleware(c martini.Context, w http.ResponseWriter, r *http.Request) {
	// Inject our db when it is requested by handlers
	c.MapTo(db, (*DB)(nil))
}

func init() {
	log.Println("Connecting to database...")
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	dbopen, err := sql.Open("mysql", dbuser+":"+dbpass+"@/"+dbname+"?charset=utf8&parseTime=true")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	//defer db.Close() // I DUNNO IF IT WORKS HERE, LETS TEST

	dialect := gorp.MySQLDialect{"InnoDB", "UTF8"}

	// construct a gorp DbMap
	dbmap := gorp.DbMap{Db: dbopen, Dialect: dialect}
	dbmap.TraceOn("[gorp]", log.New(os.Stdout, "[DB]", log.Lmicroseconds))

	// Adding schemes to my ORM
	dbmap.AddTableWithName(User{}, "user").SetKeys(false, "userid")
	dbmap.AddTableWithName(Profile{}, "profile").SetKeys(false, "profileid", "source")
	dbmap.AddTableWithName(Pic{}, "pic").SetKeys(false, "picid")
	dbmap.AddTableWithName(Token{}, "token").SetKeys(false, "tokenid")

	// Adding to local vairable
	db = &dbmap
	log.Println("Database connected!")
}

type User struct {
	UserId     string    `db:"userid" json:"userid"`         // *PK max: 20
	UserName   string    `db:"username" json:"username"`     // *UQ max: 20
	LikeCount  int       `db:"likecount" json:"likecount"`   // default: 0
	Creation   time.Time `db:"creation" json:"creation"`     // *NN
	LastUpdate time.Time `db:"lastupdate" json:"lastupdate"` // *NN
	Deleted    bool      `db:"deleted" json:"deleted"`       // default: 0
	Admin      bool      `db:"admin" json:"admin"`           // default: 0
}

type Pic struct {
	PicId    string    `db:"picid" json:"picid"`       // *PK max: 20
	UserId   string    `db:"userid" json:"userid"`     // *FK max: 20
	Creation time.Time `db:"creation" json:"creation"` // *NN
	Deleted  bool      `db:"deleted" json:"deleted"`   // default: 0
}

type Profile struct {
	ProfileId    string    `db:"profileid" json:"profileid"`       // *PK max: 40
	Source       string    `db:"source" json:"source"`             // *PK enum: facebook, google, twitter
	UserId       string    `db:"userid" json:"userid"`             // *FK max: 20
	UserName     string    `db:"username" json:"username"`         // *NN max: 20
	Email        string    `db:"email" json:"email"`               // *NN max: 40
	FullName     string    `db:"fullname" json:"fullname"`         // *NN max: 40
	Gender       string    `db:"gender" json:"gender"`             // *NN enum: male, female
	ProfileUrl   string    `db:"profileurl" json:"profileurl"`     // *NN max: 100
	Language     string    `db:"language" json:"language"`         // max: 10, default: pt_BR
	Verified     bool      `db:"verified" json:"verified"`         // default: 0
	FirstName    string    `db:"firstname" json:"firstname"`       // *NN max: 20
	LastName     string    `db:"lastname" json:"lastname"`         // *NN max: 20
	SourceUpdate time.Time `db:"sourceupdate" json:"sourceupdate"` // *NN
	AccessToken  string    `db:"accesstoken" json:"accesstoken"`   // *NN max: 100
	RefreshToken string    `db:"refreshtoken" json:"refreshtoken"` // *NN max: 100
	Scope        string    `db:"scope" json:"scope"`               // *NN max: 40
	TokenExpiry  time.Time `db:"tokenexpiry" json:"tokenexpiry"`   // *NN Time token spiries
	Creation     time.Time `db:"creation" json:"creation"`         // *NN
	LastUpdate   time.Time `db:"lastupdate" json:"lastupdate"`     // *NN
}

type Token struct {
	TokenId  string    `db:"tokenid" json:"tokenid"` // *PK max: 20
	UserId   string    `db:"userid" json:"userid"`   // *FK max: 20
	Creation time.Time `db:"creation" json:"-"`      // *NN
}
