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

	log.Println("Database connected!")

	// Adding schemes to my ORM
	dbmap.AddTableWithName(User{}, "user").SetKeys(false, "userid")
	dbmap.AddTableWithName(Profile{}, "profile").SetKeys(false, "profileid")
	dbmap.AddTableWithName(Pic{}, "pic").SetKeys(false, "picid")
	dbmap.AddTableWithName(Token{}, "token").SetKeys(false, "tokenid")
	dbmap.AddTableWithName(Channel{}, "channel").SetKeys(false, "channelid")
	dbmap.AddTableWithName(Image{}, "image").SetKeys(false, "imageid")
	dbmap.AddTableWithName(Url{}, "url").SetKeys(false, "urlid")
	dbmap.AddTableWithName(Content{}, "content").SetKeys(false, "contentid")

	// Adding to local vairable
	db = &dbmap

	log.Println("Start routine to create the default values of our datas...")

	checkAndCreateDefaultPic(db)

	checkAndCreateDefaultImage(db)

	checkAndCreateAdminUser(db)

	checkAndCreateChannels(db)

	log.Println("All default values has been created.")

	dbmap.TraceOn("[SQL]", log.New(os.Stdout, "[DB]", log.Lmicroseconds))

}

func checkAndCreateDefaultPic(db DB) {
	// Adding default value to pic
	// pic/default.png !
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
			} else {
				log.Println("Default pic created!")
			}
		}
	} else {
		log.Printf("Error searching for default pic. %s\n", err)
	}
}

func checkAndCreateDefaultImage(db DB) {
	// Adding default value to image
	// img/default-small.png
	// img/default-medium.png
	// img/default-large.png
	count, err := db.SelectInt("select count(*) from image where imageid=?", "default")
	if err == nil {
		if count == 0 {
			image := &Image{
				ImageId:  "default",
				Creation: time.Now(),
				Deleted:  false,
			}
			err := db.Insert(image)
			if err != nil {
				log.Printf("Error creating default image. %s\n", err)
			} else {
				log.Printf("Default image created!\n")
			}
		}
	} else {
		log.Printf("Error searching for default image. %s\n", err)
	}
}

func checkAndCreateAdminUser(db DB) {
	count, err := db.SelectInt("select count(*) from user where userid=?", "admin")
	if err == nil {
		if count == 0 {
			admin := &User{
				UserId:     "admin",
				UserName:   "Admin",
				PicId:      "default",
				FullName:   "Admin",
				LikeCount:  0,
				Creation:   time.Now(),
				LastUpdate: time.Now(),
				Deleted:    false,
				Admin:      true,
			}

			err := db.Insert(admin)
			if err != nil {
				log.Printf("Error creating the admin user. %s\n", err)
			} else {
				log.Println("User admin created!")
			}
		}
	} else {
		log.Printf("Error searching for user admin. %s\n", err)
	}
}

func checkAndCreateChannels(db DB) {
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
					log.Printf("Error when creating the channel %s. %s\n", channel.ChannelName, err)
				} else {
					log.Printf("Channel %s created!\n", channel.ChannelName)
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
	"Entretenimento",
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
	"Entretenimento",
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
