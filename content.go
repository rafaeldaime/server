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

	//bs, _ := ioutil.ReadAll(resp.Body)
	//log.Printf("\n\nBODY: %s \n\n", bs)

	imageUrl := ""

	openTitle := false // Indicates if title's tag has open

	// This function create a Tokenizer for an io.Reader, obs. HTML should be UTF-8
	z := html.NewTokenizer(resp.Body)
	for {
		tokenType := z.Next()

		if tokenType == html.ErrorToken {
			if z.Err() == io.EOF { // EVERTHINGS WORKS WELL!
				break
			}
			// Ops, we've got something wrong
			return nil, "", errors.New(
				fmt.Sprintf("Desculpe, mas ocorreu um erro ao extrair as tags HTML da pagina passada. %s.", z.Err()))
		}

		switch tokenType {
		case html.SelfClosingTagToken:
			token := z.Token()
			if token.Data == "meta" {
				// Extracting this meta data information
				key := ""
				value := ""
				for _, attr := range token.Attr {
					switch attr.Key {
					case "property", "name":
						key = attr.Val
					case "content":
						value = attr.Val
					}
				}

				switch key {
				case "og:title", "twitter:title":
					log.Println("Title:" + strings.TrimSpace(value) + "....")
					if strings.TrimSpace(value) != "" {
						newContent.Title = strings.TrimSpace(value)
					}

				case "description", "og:description", "twitter:description":
					log.Println("Description:" + strings.TrimSpace(value) + "....")
					if strings.TrimSpace(value) != "" {
						newContent.Description = strings.TrimSpace(value)
					}
				case "og:image", "twitter:image":
					log.Println("imageUrl:" + strings.TrimSpace(value) + "....")
					if strings.TrimSpace(value) != "" {
						imageUrl = strings.TrimSpace(value)
					}
					// case "og:url", "twitter:url":
					// 	fullUrl = value
				}

			}
		case html.TextToken:
			// If title aren't set by any meta tags, so let's use the page's title
			if openTitle && newContent.Title == "" {
				token := z.Token()
				newContent.Title = strings.TrimSpace(token.Data)
				log.Println("FIADAPUTAAAAAAAAAAAAAAAAAAAAA" + strings.TrimSpace(token.Data) + "!!!")
			}
		case html.StartTagToken, html.EndTagToken:
			tn, _ := z.TagName()
			// Check if it is an title tag opennig, it's the fastest way to compare bytes
			if len(tn) == 5 && tn[0] == 't' && tn[1] == 'i' && tn[2] == 't' && tn[3] == 'l' && tn[4] == 'e' {
				if tokenType == html.StartTagToken {
					openTitle = true
				} else {
					openTitle = false
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

	// NOT SAVING JUST FOR DEBUGGING
	//err = db.Insert(content)

	if err != nil {
		return nil, err
	}

	return content, nil
}
