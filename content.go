package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go.net/html"
	"github.com/dchest/uniuri" // give us random URIs
	"github.com/extemporalgenome/slug"
)

func GetContent(newContent *Content) (*Content, string, error) {
	resp, err := http.Get(newContent.FullUrl)
	if err != nil {
		return nil, "", errors.New(
			fmt.Sprintf("Desculpe, ocorreu ao tentar recuperar a pagina referente a URL passada. %s.", err))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", errors.New(
			fmt.Sprintf("Desculpe, mas a pagina passada respondeu indevidamente. O Status Code recebido foi: %d.", resp.StatusCode))
	}

	imageUrl := ""

	// This function create a Tokenizer for an io.Reader, obs. HTML should be UTF-8
	z := html.NewTokenizer(resp.Body)
	for {
		tokenType := z.Next()

		if tokenType == html.ErrorToken {
			if z.Err() == io.EOF { // EVERTHINGS WORKS WELL!
				break
			}
			// Ops, we've got something wrong, it isn't an EOF token
			return nil, "", errors.New(
				fmt.Sprintf("Desculpe, mas ocorreu um erro ao extrair as tags HTML da pagina passada. %s.", z.Err()))
		}

		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken:

			token := z.Token()
			// Check if it is an title tag opennig, it's the fastest way to compare bytes
			if token.Data == "title" {
				// log.Printf("TAG: '%v'\n", token.Data)
				nextTokenType := z.Next()
				if nextTokenType == html.TextToken {
					nextToken := z.Token()
					newContent.Title = strings.TrimSpace(nextToken.Data)
					// log.Println("<title> = " + newContent.Title)
				}

			} else if token.Data == "meta" {
				key := ""
				value := ""

				// log.Printf("NewMeta: %s : ", token.String())

				// Extracting this meta data information
				for _, attr := range token.Attr {
					switch attr.Key {
					case "property", "name":
						key = attr.Val
					case "content":
						value = attr.Val
					}
				}

				switch key {
				case "title", "og:title", "twitter:title":
					if strings.TrimSpace(value) != "" {
						newContent.Title = strings.TrimSpace(value)
						// log.Printf("Title: %s\n", strings.TrimSpace(value))
					}

				case "og:site_name", "twitter:domain":
					if strings.TrimSpace(value) != "" {
						// newContent.Title = strings.TrimSpace(value)
						log.Printf("Site Name: %s\n", strings.TrimSpace(value))
					}

				case "description", "og:description", "twitter:description":
					if strings.TrimSpace(value) != "" {
						newContent.Description = strings.TrimSpace(value)
						// log.Printf("Description: %s\n", strings.TrimSpace(value))
					}
				case "og:image", "twitter:image", "twitter:image:src":
					if strings.TrimSpace(value) != "" {
						imageUrl = strings.TrimSpace(value)
						// log.Printf("Image: %s\n", strings.TrimSpace(value))
					}
				case "og:url", "twitter:url":
					if strings.TrimSpace(value) != "" {
						newContent.FullUrl = strings.TrimSpace(value)
						// log.Printf("Url: %s\n", strings.TrimSpace(value))
					}
				}
			}
		}
	}
	// Limiting the size of Title and Description to 250 characters
	if len(newContent.Title) > 250 {
		newContent.Title = newContent.Title[0:250]
	}
	if len(newContent.Description) > 250 {
		newContent.Description = newContent.Description[0:250]
	}

	log.Printf("Title: %s\n description: %s\n imageUrl:%s\n",
		newContent.Title, newContent.Description, imageUrl)

	return newContent, imageUrl, nil
}

func consumeTag(content *Content, key, value string) {

}

func CreateContent(db DB, user *User, url *Url, img *Image, newContent *Content) (*Content, error) {
	contentId := uniuri.NewLen(20)
	u, err := db.Get(Content{}, contentId)
	for err == nil && u != nil {
		contentId := uniuri.NewLen(20)
		u, err = db.Get(Content{}, contentId)
	}

	if err != nil {
		return nil, err
	}

	s := slug.Slug(newContent.Title)

	// Let's check if this slug already exists,
	// if existis, we will increment a sulfix to it
	newSlug := s
	increment := 1
	count, err := db.SelectInt("select count(*) from content where slug=?", newSlug)
	for err == nil && count != 0 {
		increment += 1
		newSlug = fmt.Sprintf("%s-%d", s, increment)
		count, err = db.SelectInt("select count(*) from content where slug=?", newSlug)
	}

	log.Printf("SLUG: %s, inc: %d, count: %d\n", newSlug, increment, count)

	if err != nil {
		return nil, err
	}

	content := &Content{
		ContentId:   contentId,
		UrlId:       url.UrlId,
		ChannelId:   newContent.ChannelId,
		Title:       newContent.Title,
		Slug:        newSlug,
		Description: newContent.Description,
		UserId:      user.UserId,
		ImageId:     "default", // We will check the if there is an image below
		MaxSize:     "",        // It sinalizes that there isn't a image for this content
		LikeCount:   0,
		Creation:    time.Now(),
		LastUpdate:  time.Now(),
		Deleted:     false,
		FullUrl:     newContent.FullUrl,
	}

	if img != nil {
		content.ImageId = img.ImageId
		content.MaxSize = img.MaxSize
	}

	err = db.Insert(content)

	if err != nil {
		return nil, err
	}

	return content, nil
}
