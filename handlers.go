package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/dchest/uniuri" // give us random URIs
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

func getAllCategories(db DB, r render.Render, req *http.Request) {
	qs := req.URL.Query()
	order := qs.Get("order")
	log.Printf("CATEGORIES ORDER: %v \n", order)
	var categories []Category
	query := "select * from category"

	if order == "categoryname" {
		query += " order by categoryname asc"
	} else if order == "-categoryname" {
		query += " order by categoryname desc"
	}

	_, err := db.Select(&categories, query)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao selecionar a lista de categorias %s.", err)))
		return
	}
	r.JSON(http.StatusOK, categories)
}

func getContents(db DB, auth Auth, r render.Render, req *http.Request) {
	user, err := auth.GetUser()
	qs := req.URL.Query()
	order := qs.Get("order")
	log.Printf("CONTENTS ORDER: %v \n", order)
	// limit := qs.Get("limit")
	// page := qs.Get("page")

	fullContents, err := GetAllFullContent(db, user, 30, 1)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao buscar os conteudos completos no banco %s.", err)))
		return
	}
	r.JSON(http.StatusOK, fullContents)
}

func meHandler(auth Auth, r render.Render, req *http.Request) {
	user, err := auth.GetUser()
	if err != nil {
		// Can't continue without the user
		// Abort, so AuthMiddleware could alert the user for credentials requirement
		return
	}
	r.JSON(http.StatusOK, user)

}

type LikeReturn struct {
	ContentId string `json:"contentid"`
	LikeCount int    `json:"likecount"`
	ILike     bool   `json:"ilike"`
}

func AddLikeHandler(db DB, auth Auth, r render.Render, req *http.Request, params martini.Params) {
	user, err := auth.GetUser()
	if err != nil {
		return // AuthMiddleware will response user
	}

	contentId := params["contentid"]
	incr := false // Is it necessary to update content's likecount?

	// Let's select if this contentid really exists
	contentobj, err := db.Get(Content{}, contentId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao se buscar o conteudo desejado. %s.", err)))
		return
	}

	if contentobj == nil {
		r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas o contentid %s informado nao foi encontrado.", contentId)))
		return
	}

	content := contentobj.(*Content)

	likeobj, err := db.Get(ContentLike{}, contentId, user.UserId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao se buscar o contentlike desejado. %s.", err)))
		return
	}

	if likeobj == nil {
		contentLike := &ContentLike{
			ContentId:  contentId,
			UserId:     user.UserId,
			Creation:   time.Now(),
			LastUpdate: time.Now(),
			Deleted:    false,
		}
		err = db.Insert(contentLike)
		if err != nil {
			r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas ocorreu um erro adicionar seu like. %s.", err)))
			return
		}
		incr = true // Updates contents like
	} else {
		contentLike := likeobj.(*ContentLike)
		if contentLike.Deleted == true {
			contentLike.Deleted = false
			contentLike.LastUpdate = time.Now()
			count, err := db.Update(contentLike)
			if err != nil {
				r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
					"Desculpe, mas ocorreu um erro atualizar seu like. %s.", err)))
				return
			}
			if count == 0 {
				r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
					"Desculpe, mas seu like nao foi atualizado.")))
				return
			}
			incr = true // Updates contents like
		}
	}

	// Like added, lets update the content likecount number if it's necessary
	if incr {
		content.LikeCount += 1
		count, err := db.Update(content)
		if err != nil {
			r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas ocorreu um erro atualizar o centeudo. %s.", err)))
			return
		}
		if count == 0 {
			r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas seu o conteudo nao foi atualizado.")))
			return
		}
	}

	r.JSON(http.StatusOK, LikeReturn{
		ContentId: content.ContentId,
		LikeCount: content.LikeCount,
		ILike:     true,
	})
}

func DeleteLikeHandler(db DB, auth Auth, r render.Render, req *http.Request, params martini.Params) {
	user, err := auth.GetUser()
	if err != nil {
		return // AuthMiddleware will response user
	}

	contentId := params["contentid"]
	decr := false // Is it necessary to update content's likecount?

	// Let's select if this contentid really exists
	contentobj, err := db.Get(Content{}, contentId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao se buscar o conteudo desejado. %s.", err)))
		return
	}

	if contentobj == nil {
		r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas o contentid %s informado nao foi encontrado.", contentId)))
		return
	}

	content := contentobj.(*Content)

	likeobj, err := db.Get(ContentLike{}, content.ContentId, user.UserId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao se buscar o contentlike desejado. %s.", err)))
		return
	}

	if likeobj == nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas voce ainda nao deu like neste conteudo %s.", content.ContentId)))
		return
	} else {
		// Like really exists for this content
		contentLike := likeobj.(*ContentLike)
		if contentLike.Deleted == false {
			contentLike.Deleted = true
			contentLike.LastUpdate = time.Now()
			count, err := db.Update(contentLike)
			if err != nil {
				r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
					"Desculpe, mas ocorreu um erro atualizar seu like. %s.", err)))
				return
			}
			if count == 0 {
				r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
					"Desculpe, mas seu like nao foi atualizado.")))
				return
			}
			decr = true // Updates contents like
		}
	}

	// Like added, lets update the content likecount number if it's necessary
	if decr {
		content.LikeCount -= 1
		count, err := db.Update(content)
		if err != nil {
			r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas ocorreu um erro atualizar o centeudo. %s.", err)))
			return
		}
		if count == 0 {
			r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas seu o conteudo nao foi atualizado.")))
			return
		}
	}

	r.JSON(http.StatusOK, LikeReturn{
		ContentId: content.ContentId,
		LikeCount: content.LikeCount,
		ILike:     false,
	})
}

// 5MB
const MAX_MEMORY = 5 * 1024 * 1024

func changeContentImage(db DB, auth Auth, params martini.Params, r render.Render, req *http.Request) {
	contentId := params["contentid"]

	// Get user in this session
	user, err := auth.GetUser()
	if err != nil {
		return // AuthMiddleware will response user
	}

	// Checks if the session's user really is the content's owner
	query := "select * from content where content.contentid=? and content.userid=?"
	var contents []*Content
	_, err = db.Select(&contents, query, contentId, user.UserId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao buscar seu conteudo no banco. %s.", err)))
		return
	}

	if len(contents) == 0 {
		r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas nao foi encontrado um conteudo seu com o ContentId %s informado.", contentId)))
		return
	}

	oldContent := contents[0]

	if err := req.ParseMultipartForm(MAX_MEMORY); err != nil {
		r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, nao foi possivel receber seu arquivo.", err)))
		return
	}

	for key, value := range req.MultipartForm.Value {
		log.Printf("FILE: %s:%s", key, value)
	}

	var image *Image
	for _, fileHeaders := range req.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			defer file.Close()
			if err != nil {
				r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
					"Desculpe, ocorreu um erro abrir o arquio enviado. %s.", err)))
				return
			}

			// SAVE IMAGE ()

			image, err = SaveImage(file)
			if err != nil {
				r.JSON(http.StatusForbidden, NewError(ErrorCodeDefault, fmt.Sprintf(
					"Desculpe, ocorreu um erro abrir o arquio enviado. %s.", err)))
				return
			}

			// Stop on the first file!
			break

			// path := fmt.Sprintf("public/%s", fileHeader.Filename)
			// buf, _ := ioutil.ReadAll(file)
			// ioutil.WriteFile(path, buf, os.ModePerm)
		}
	}

	// Updating fields on saved content... to appoint the new image
	oldContent.ImageId = image.ImageId

	log.Printf("oldContent: %#v\n", oldContent)

	count, err := db.Update(oldContent)
	if err != nil {
		log.Printf("Erro: %#v", err)
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao tentar atualizar seu conteudo. %s.", err)))
		return
	}
	if count == 0 {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas seu conteudo nao foi atualizado.")))
		return
	}

	r.JSON(http.StatusOK, oldContent)
	return
}

func updateContent(db DB, auth Auth, params martini.Params, r render.Render, req *http.Request) {
	content := &Content{}
	err := json.NewDecoder(req.Body).Decode(content)
	if err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		r.JSON(http.StatusNotAcceptable, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Nao foi possivel decodificar o objeto Json: %s! %s.", body, err)))
		return
	}

	content.ContentId = params["contentid"]

	// Get user in this session
	user, err := auth.GetUser()
	if err != nil {
		return // AuthMiddleware will response user
	}

	// Checks if the session's user really is the content's owner
	query := "select * from content where content.contentid=? and content.userid=?"
	var contents []*Content
	_, err = db.Select(&contents, query, content.ContentId, user.UserId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao buscar seu conteudo no banco. %s.", err)))
		return
	}

	if len(contents) == 0 {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas nao foi encontrado um conteudo seu com o ContentId %s informado.", content.ContentId)))
		return
	}

	oldContent := contents[0]

	// Let's check if CategoryId passed by user really exists
	obj, err := db.Get(Category{}, content.CategoryId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao buscar a categoria fornecida com esse id. %s.", err)))
		return
	}
	if obj == nil {
		r.JSON(http.StatusMethodNotAllowed, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas a categoryId %s fornecida nao existe.", content.CategoryId)))
		return
	}

	content.Title = StripTitle(content.Title)
	content.Description = StripDescription(content.Description)

	if content.Title == "" || content.Description == "" {
		r.JSON(http.StatusMethodNotAllowed, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpa, mas o titulo e a descricao nao podem ficar vazios!")))
		return
	}

	slug, err := GetSlug(content)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao tentar criar o slug do seu conteudo. %s.", err)))
		return
	}

	// Updating fields on saved content
	oldContent.CategoryId = content.CategoryId
	oldContent.Title = content.Title
	oldContent.Slug = slug
	oldContent.Description = content.Description
	oldContent.LastUpdate = time.Now()

	log.Printf("Content: %#v\n\n", content)
	log.Printf("oldContent: %#v\n", oldContent)

	count, err := db.Update(oldContent)
	if err != nil {
		log.Printf("Erro: %#v", err)
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao tentar atualizar seu conteudo. %s.", err)))
		return
	}
	if count == 0 {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas seu conteudo nao foi atualizado.")))
		return
	}

	r.JSON(http.StatusOK, oldContent)
	return
}

type addContentDataStruct struct {
	FullUrl string
}

func addContent(db DB, auth Auth, r render.Render, req *http.Request) {
	var addContentData addContentDataStruct
	err := json.NewDecoder(req.Body).Decode(&addContentData)
	if err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		r.JSON(http.StatusNotAcceptable, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Nao foi possivel decodificar o objeto Json: %s! %s.", body, err)))
		return
	}

	if addContentData.FullUrl == "" {
		r.JSON(http.StatusMethodNotAllowed, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpa, mas os campo da URL nao foi informado!")))
		return
	}

	// Get user in this session
	user, err := auth.GetUser()
	if err != nil || user == nil {
		return // AuthMiddleware will response user
	}

	// If user sent us an URL without http, we will put it in the begin of URL
	hasHtml := regexp.MustCompile(`^https?:\/\/`)
	if !hasHtml.MatchString(addContentData.FullUrl) {
		addContentData.FullUrl = "http://" + addContentData.FullUrl
	}

	// Ver se ja existe um conteudo DESTE usuario com a FullUrl passada
	query := "select content.* from content, url where content.urlid=url.urlid and content.userid=? and url.fullurl=?"
	var contents []Content
	_, err = db.Select(&contents, query, user.UserId, addContentData.FullUrl)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao se verificar se ja existe um conteudo seu com essa URL. %s.", err)))
		return
	}

	// If this content was already added by this user, lets return it
	if len(contents) > 0 {
		r.JSON(http.StatusOK, contents[0])
		return
	}

	// Getting the content on the WEB based on the FullUrl given by user
	content, imageUrl, err := GetContent(addContentData.FullUrl)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao tentar baixar o conteudo da URL informada. %s.", err)))
		return
	}

	// Lets save our new URL
	url, err := saveUrl(db, user, addContentData.FullUrl)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao tentar adicionar sua URL. %s.", err)))
		return
	}

	var img *Image
	if imageUrl != "" {
		img, err = GetImage(db, imageUrl)
		if err != nil {
			// If we can't get this img for some reason, it doesn't metter
			log.Printf("!!! Something got wrong when getting the image in %s. %s", imageUrl, err)
		}
	}

	content, err = CreateContent(db, user, url, img, content)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao tentar adicionar seu conteudo. %s.", err)))
		return
	}

	r.JSON(http.StatusOK, content)
	return
}

func saveUrl(db DB, user *User, fullUrl string) (*Url, error) {
	urlId := uniuri.NewLen(5)
	u, err := db.Get(Url{}, urlId)
	for err == nil && u != nil {
		urlId := uniuri.NewLen(5)
		u, err = db.Get(Url{}, urlId)
	}

	if err != nil {
		return nil, err
	}

	url := &Url{
		UrlId:     urlId,
		FullUrl:   fullUrl,
		UserId:    user.UserId,
		Creation:  time.Now(),
		ViewCount: 0,
	}

	err = db.Insert(url)
	if err != nil {
		return nil, err
	}

	return url, nil
}
