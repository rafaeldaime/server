package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/dchest/uniuri"
	"github.com/extemporalgenome/slug"
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
	dbmap.AddTableWithName(Token{}, "token").SetKeys(false, "tokenid", "userid")
	dbmap.AddTableWithName(Category{}, "category").SetKeys(false, "categoryid")
	dbmap.AddTableWithName(Image{}, "image").SetKeys(false, "imageid")
	dbmap.AddTableWithName(Url{}, "url").SetKeys(false, "urlid")
	dbmap.AddTableWithName(Content{}, "content").SetKeys(false, "contentid")
	dbmap.AddTableWithName(FullContent{}, "fullcontent").SetKeys(false, "contentid")
	dbmap.AddTableWithName(ContentLike{}, "contentlike").SetKeys(false, "contentid", "userid")
	dbmap.AddTableWithName(Access{}, "access").SetKeys(false, "accessid")

	// Adding to local vairable
	db = &dbmap

	log.Println("Start routine to create the default values of our datas...")

	checkAndCreateDefaultPic(db)

	checkAndCreateDefaultImage(db)

	checkAndCreateAnonymousUser(db)

	checkAndCreateCategories(db)

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

func checkAndCreateAnonymousUser(db DB) {
	count, err := db.SelectInt("select count(*) from user where userid=?", "anonymous")
	if err == nil {
		if count == 0 {
			anonymous := &User{
				UserId:     "anonymous",
				UserName:   "Anonimo",
				PicId:      "default",
				FullName:   "Usuario Anonimo",
				LikeCount:  0,
				Creation:   time.Now(),
				LastUpdate: time.Now(),
				Deleted:    false,
				Admin:      false,
			}

			err := db.Insert(anonymous)
			if err != nil {
				log.Printf("Error creating the anonymous user. %s\n", err)
			} else {
				log.Println("User anonymous created!")
			}
		}
	} else {
		log.Printf("Error searching for user anonymous. %s\n", err)
	}
}

func checkAndCreateCategories(db DB) {
	for i, categoryName := range categoryList {
		categorySlug := slug.Slug(categoryName)
		count, err := db.SelectInt("select count(*) from category where categoryslug=?", categorySlug)
		if err != nil {
			log.Printf("Error searching for the category with categorySlug %s\n", categorySlug)
			log.Println("Stopping the creation of categories")
			return
		}
		if count == 0 {
			category := &Category{
				CategoryId:   uniuri.NewLen(20),
				CategoryName: categoryName,
				CategorySlug: categorySlug,
				LikeCount:    0,
			}
			if i == 0 { // "Sem Categoria" is my default category
				category.CategoryId = "default"
			}
			err := db.Insert(category)
			if err != nil {
				log.Printf("Error when creating the category %s. %s\n", category.CategoryName, err)
			} else {
				log.Printf("Category %s created!\n", category.CategoryName)
			}
		}
	}
}

var categoryList = []string{
	"Sem categoria",
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
