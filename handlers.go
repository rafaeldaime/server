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

func getAllChannels(db DB, r render.Render, req *http.Request) {
	qs := req.URL.Query()
	order := qs.Get("order")
	log.Printf("ORDER: %v \n", order)
	var channels []Channel
	query := "select * from channel"

	if order == "channelname" {
		query += " order by channelname asc"
	} else if order == "-channelname" {
		query += " order by channelname desc"
	}

	_, err := db.Select(&channels, query)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao selecionar a lista de canais %s.", err)))
		return
	}
	r.JSON(http.StatusOK, channels)
}

func meHandler(auth Auth, r render.Render) {
	user, err := auth.GetUser()
	if err != nil {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Voce nao esta logado! %s.", err)))
	} else {
		r.JSON(http.StatusOK, user)
	}

}

// 5MB
const MAX_MEMORY = 5 * 1024 * 1024

func changeContentImage(db DB, auth Auth, params martini.Params, r render.Render, req *http.Request) {
	contentId := params["contentid"]

	// Get user in this session
	user, err := auth.GetUser()
	if err != nil || user == nil {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpa, voce precisa estar logado para adicionar um conteudo! %s.", err)))
		return
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
	if err != nil || user == nil {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpa, voce precisa estar logado para adicionar um conteudo! %s.", err)))
		return
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

	// Let's check if ChannelId passed by user really exists
	obj, err := db.Get(Channel{}, content.ChannelId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao buscar o canal fornecido com esse id. %s.", err)))
		return
	}
	if obj == nil {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas o channelid %s fornecido nao existe.", content.ChannelId)))
		return
	}

	content.Title = StripTitle(content.Title)
	content.Description = StripDescription(content.Description)

	if content.Title == "" || content.Description == "" {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
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
	oldContent.ChannelId = content.ChannelId
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

type newContentDataStruct struct {
	ChannelId string
	FullUrl   string
}

func createNewContent(db DB, auth Auth, r render.Render, req *http.Request) {
	var newContentData newContentDataStruct
	err := json.NewDecoder(req.Body).Decode(&newContentData)
	if err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		r.JSON(http.StatusNotAcceptable, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Nao foi possivel decodificar o objeto Json: %s! %s.", body, err)))
		return
	}

	if newContentData.FullUrl == "" || newContentData.ChannelId == "" {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpa, mas os campos fullurl e/ou channelid nao foram informados!")))
		return
	}

	// Get user in this session
	user, err := auth.GetUser()
	if err != nil || user == nil {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpa, voce precisa estar logado para adicionar um conteudo! %s.", err)))
		return
	}

	// If user sent us an URL without http, we will put it in the begin of URL
	hasHtml := regexp.MustCompile(`^https?:\/\/`)
	if !hasHtml.MatchString(newContentData.FullUrl) {
		newContentData.FullUrl = "http://" + newContentData.FullUrl
	}

	// Ver se ja existe um conteudo DESTE usuario com a FullUrl passada
	query := "select content.* from content, url where content.urlid=url.urlid and content.userid=? and url.fullurl=?"
	var contents []Content
	_, err = db.Select(&contents, query, user.UserId, newContentData.FullUrl)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao se verificar se ja existe um conteudo seu com essa URL. %s.", err)))
		return
	}

	if len(contents) > 0 {
		r.JSON(http.StatusOK, contents[0])
		return
	}

	// Let's check if ChannelId passed by user really exists
	obj, err := db.Get(Channel{}, newContentData.ChannelId)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao buscar o canal fornecido com esse id. %s.", err)))
		return
	}
	if obj == nil {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas o channelid %s fornecido nao existe.", newContentData.ChannelId)))
		return
	}

	// Getting the content on the WEB based on the newContent.FullUrl given by user
	content, imageUrl, err := GetContent(newContentData.FullUrl)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao tentar baixar o conteudo da URL informada. %s.", err)))
		return
	}

	// Lets save our new URL
	url, err := saveUrl(db, user, newContentData.FullUrl)
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

	// Adding the channel id passed by user
	content.ChannelId = newContentData.ChannelId
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
