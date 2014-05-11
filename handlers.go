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
	"github.com/martini-contrib/render"
)

func getAllChannels(db DB, r render.Render) {
	var channels []Channel
	_, err := db.Select(&channels, "select * from channel")
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao selecionar a lista de canais %s.", err)))
		return
	}
	r.JSON(http.StatusOK, channels)
}

func createNewContent(db DB, auth Auth, r render.Render, req *http.Request) {
	var newContent Content
	err := json.NewDecoder(req.Body).Decode(&newContent)
	if err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		r.JSON(http.StatusNotAcceptable, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Nao foi possivel decodificar o objeto Json: %s! %s.", body, err)))
		return
	}

	if newContent.FullUrl == "" || newContent.ChannelId == "" {
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
	// WE FUCK CANT DO TE LINE BELOW, CAUSE WE WILL FUCK THE YOUTUBE ID's, FOR EXAMPLE
	//newContent.FullUrl = strings.ToLower(newContent.FullUrl)
	hasHtml := regexp.MustCompile(`^https?:\/\/`)
	if !hasHtml.MatchString(newContent.FullUrl) {
		newContent.FullUrl = "http://" + newContent.FullUrl
	}

	// Ver se ja existe um conteudo DESTE usuario com a FullUrl passada
	query := "select content.* from content, url where content.urlid=url.urlid and content.userid=? and url.fullurl=?"
	var contents []Content
	// !!! I'VE CHANGED HERE JUST FOR DEBBUG
	_, err = db.Select(&contents, query, user.UserId+"foda", newContent.FullUrl)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao se verificar se ja existe um conteudo seu com essa URL. %s.", err)))
		return
	}

	if len(contents) > 0 {
		contents[0].FullUrl = newContent.FullUrl
		r.JSON(http.StatusOK, contents[0])
		return
	}

	// Getting the content on the WEB based on the newContent.FullUrl given by user
	gettedContent, imageUrl, err := GetContent(&newContent)
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas ocorreu um erro ao tentar baixar o conteudo da URL informada. %s.", err)))
		return
	}

	// Lets save our new URL
	url, err := saveUrl(db, user, gettedContent.FullUrl)
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

	content, err := CreateContent(db, user, url, img, gettedContent)
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
