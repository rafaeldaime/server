package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/dchest/uniuri"
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
	dbmap.AddTableWithName(Profile{}, "profile").SetKeys(false, "profileid")
	dbmap.AddTableWithName(Pic{}, "pic").SetKeys(false, "picid")
	dbmap.AddTableWithName(Token{}, "token").SetKeys(false, "tokenid")
	dbmap.AddTableWithName(Channel{}, "channel").SetKeys(false, "channelid")

	// Adding to local vairable
	db = &dbmap

	checkInsertDefaultPic(db)

	checkInsertChannels(db)

	log.Println("Database connected!")
}

func checkInsertDefaultPic(db DB) {
	// Adding default value to pic
	// Never exclude pic default.png !
	count, err := db.SelectInt("select count(*) from pic where picid=?", "default")
	if err == nil {
		if count == 0 {
			pic := &Pic{
				PicId:    "default",
				Creation: time.Now(),
				Deleted:  false,
			}
			err := db.Insert(pic)
			if err != nil {
				log.Printf("Error creating default pic. %s\n", err)
			}
		}
	} else {
		log.Printf("Error searching for default pic. %s\n", err)
	}
}

func checkInsertChannels(db DB) {
	for i, channelName := range channelList {
		channelSlug := slugList[i]
		count, err := db.SelectInt("select count(*) from channel where channelslug=?", channelSlug)
		if err == nil {
			if count == 0 {
				channel := &Channel{
					ChannelId:   uniuri.NewLen(20),
					ChannelName: channelName,
					ChannelSlug: channelSlug,
					LikeCount:   0,
				}
				err := db.Insert(channel)
				if err != nil {
					log.Printf("Error when inserting the channel. %v\n", channel)
				}
			}
		} else {
			log.Printf("Error searching for the channel with channelslug %s\n", channelSlug)
		}
	}
}

var channelList = []string{
	"Animais",
	"Arte e Cultura",
	"Beleza e Estilo",
	"Carros e Motos",
	"Casa e Decoração",
	"Ciência e Tecnologia",
	"Comidas e Bebidas",
	"Crianças",
	"Curiosidades",
	"Downloads",
	"Educação",
	"Esporte",
	"Eventos",
	"Família",
	"Filmes",
	"Fotos",
	"Futebol",
	"Humor",
	"Internacional",
	"Internet",
	"Jogos",
	"Livro",
	"Meio ambiente",
	"Mulher",
	"Música",
	"Negócios",
	"Notícias",
	"Pessoas e Blogs",
	"Política",
	"Saúde",
	"Turismo e Viagem",
	"Vídeos",
}

var slugList = []string{
	"Animais",
	"Arte-e-Cultura",
	"Beleza-e-Estilo",
	"Carros-e-Motos",
	"Casa-e-Decoracao",
	"Ciencia-e-Tecnologia",
	"Comidas-e-Bebidas",
	"Criancas",
	"Curiosidades",
	"Downloads",
	"Educacao",
	"Esporte",
	"Eventos",
	"Familia",
	"Filmes",
	"Fotos",
	"Futebol",
	"Humor",
	"Internacional",
	"Internet",
	"Jogos",
	"Livro",
	"Meio-ambiente",
	"Mulher",
	"Musica",
	"Negocios",
	"Noticias",
	"Pessoas-e-Blogs",
	"Politica",
	"Saude",
	"Turismo-e-Viagem",
	"Videos",
}
