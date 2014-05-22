package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/charset"
	"github.com/dchest/uniuri" // give us random URIs
	"github.com/extemporalgenome/slug"
)

// ^\s+|\s+$ remove white spaces in the begin of the string
// and in the end of the string, | is like an OR operator and with /g (global)
// it will relace all substrings that match in any of these cases
// \s{2,} matches 2 or more white spaces (or newlines)
var (
	htmlNewline         = regexp.MustCompile(`<br\s*[\/]?>`)
	htmlTagsAndEntities = regexp.MustCompile(`<(.*?)>|&(.*?);`)
	spacesTogether      = regexp.MustCompile(` {2,}`)
	newlinesTogether    = regexp.MustCompile(`\n{2,}`)
	beginEndNewLines    = regexp.MustCompile(`^\s+|\s+$`)
	validTitle          = regexp.MustCompile(`^[^a-zA-Zà-úÀ-Ú0-9 \-!?]`)
	validDescription    = regexp.MustCompile(`^[^a-zA-Zà-úÀ-Ú0-9 \-_.,:;!?\n]`)
)

func StripTitle(title string) string {
	// Accept accents and - ! ?
	title = htmlTagsAndEntities.ReplaceAllString(title, "")
	title = validTitle.ReplaceAllString(title, "")
	title = beginEndNewLines.ReplaceAllString(title, "")
	title = spacesTogether.ReplaceAllString(title, " ")
	title = newlinesTogether.ReplaceAllString(title, "\n")
	return title
}

func StripDescription(description string) string {
	// Accept accents and some special characters, including a new-line
	log.Printf("Description1: '%v'", description)
	description = htmlNewline.ReplaceAllString(description, "\n")
	log.Printf("Description2: '%v'", description)
	description = htmlTagsAndEntities.ReplaceAllString(description, "")
	log.Printf("Description3: '%v'", description)
	description = validDescription.ReplaceAllString(description, "")
	log.Printf("Description4: '%v'", description)
	description = beginEndNewLines.ReplaceAllString(description, "")
	log.Printf("Description5: '%v'", description)
	description = spacesTogether.ReplaceAllString(description, " ")
	log.Printf("Description6: '%v'", description)
	description = newlinesTogether.ReplaceAllString(description, "\n")
	log.Printf("Description7: '%v'", description)
	return description
}

func GetContents(db DB, user *User, categoryslug string, limit, page int) ([]Content, error) {
	var contents []Content
	start := page*limit - limit
	query := "select content.* from content"
	var err error
	if categoryslug != "" {
		query += ", category where content.categoryid = category.categoryid"
		query += " and categoryslug = ? order by likecount desc limit ?, ?"
		_, err = db.Select(&contents, query, categoryslug, start, limit)
	} else {
		query += " order by likecount desc limit ?, ?"
		_, err = db.Select(&contents, query, start, limit)
	}
	if err != nil {
		return nil, err
	}

	if user != nil {
		var contentIds []string
		for _, content := range contents {
			contentIds = append(contentIds, content.ContentId)
		}
		var contetLikes []ContentLike
		ids := "'" + strings.Join(contentIds, "', '") + "'"
		query = "select * from contentlike"
		query += " where userid = '" + user.UserId + "' and contentid in ( " + ids + " ) and deleted = false"
		//log.Printf("QUERY= %s", query)
		_, err = db.Select(&contetLikes, query)
		if err != nil {
			return nil, err
		}
		//log.Printf("LIKES: %#v", contetLikes)
		if len(contetLikes) > 0 {
			for i, content := range contents {
				for _, contentLike := range contetLikes {
					log.Printf("%s = %s \n", content.ContentId, contentLike.ContentId)
					if content.ContentId == contentLike.ContentId {
						contents[i].ILike = true // Alter in the list
						log.Printf("CONT liked: %#v\n", contents[i])
					}
				}
			}
		}
	}
	//log.Printf("CONTS: %#v", contents)
	return contents, nil
}

func GetSlug(content *Content) (string, error) {
	s := slug.Slug(content.Title)

	// Let's check if this slug already exists,
	// if existis, we will increment a sulfix to it
	newSlug := s
	increment := 1
	count, err := db.SelectInt("select count(*) from content where contentid!=? AND slug=?", content.ContentId, newSlug)
	for err == nil && count != 0 {
		increment += 1
		newSlug = fmt.Sprintf("%s-%d", s, increment)
		count, err = db.SelectInt("select count(*) from content where contentid!=? AND slug=?", content.ContentId, newSlug)
	}
	if err != nil {
		return "", err
	}

	log.Printf("SLUG: %s, inc: %d, count: %d\n", newSlug, increment, count)
	return newSlug, nil
}

func GetContent(fullUrl string) (*Content, string, error) {
	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, "", errors.New(
			fmt.Sprintf("Desculpe, ocorreu ao tentar recuperar a pagina referente a URL passada. %s.", err))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", errors.New(
			fmt.Sprintf("Desculpe, mas a pagina passada respondeu indevidamente. O Status Code recebido foi: %d.", resp.StatusCode))
	}

	reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, "", errors.New(
			fmt.Sprintf("Erro ao decodificar o charset da pagina. %s.", err))
	}

	content := &Content{}
	imageUrl := ""

	// This function create a Tokenizer for an io.Reader, obs. HTML should be UTF-8
	z := html.NewTokenizer(reader)
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
					content.Title = strings.TrimSpace(nextToken.Data)
					// log.Println("<title> = " + content.Title)
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
						content.Title = strings.TrimSpace(value)
						// log.Printf("Title: %s\n", strings.TrimSpace(value))
					}

				case "og:site_name", "twitter:domain":
					if strings.TrimSpace(value) != "" {
						//content.SiteName = strings.TrimSpace(value)
						//log.Printf("Site Name: %s\n", strings.TrimSpace(value))
					}

				case "description", "og:description", "twitter:description":
					if strings.TrimSpace(value) != "" {
						content.Description = strings.TrimSpace(value)
						// log.Printf("Description: %s\n", strings.TrimSpace(value))
					}
				case "og:image", "twitter:image", "twitter:image:src":
					if strings.TrimSpace(value) != "" {
						imageUrl = strings.TrimSpace(value)
						// log.Printf("Image: %s\n", strings.TrimSpace(value))
					}
				case "og:url", "twitter:url":
					if strings.TrimSpace(value) != "" {
						// Not used, cause user could use a redirect service
						// fullUrl = strings.TrimSpace(value)
						// log.Printf("Url: %s\n", strings.TrimSpace(value))
					}
				}
			}
		}
	}

	// Limiting the size of Title and Description to 250 characters
	if len(content.Title) > 250 {
		content.Title = content.Title[0:250]
	}
	if len(content.Description) > 250 {
		content.Description = content.Description[0:250]
	}
	// If content description is empty, lets full fill with something
	if len(content.Description) == 0 {
		content.Description = "Veja o conteudo completo..."
	}

	// Adding the host of this content
	content.Host = resp.Request.URL.Host

	log.Printf("Title: %s\n description: %s\n host:%s\n imageUrl:%s\n",
		content.Title, content.Description, content.Host, imageUrl)

	return content, imageUrl, nil
}

func CreateContent(db DB, user *User, url *Url, img *Image, content *Content) (*Content, error) {
	contentId := uniuri.NewLen(20)
	u, err := db.Get(Content{}, contentId)
	for err == nil && u != nil {
		contentId := uniuri.NewLen(20)
		u, err = db.Get(Content{}, contentId)
	}

	if err != nil {
		return nil, err
	}

	s := slug.Slug(content.Title)

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

	newContent := &Content{
		ContentId:    contentId,
		UrlId:        url.UrlId,
		CategoryId:   "default", // First category will be "Sem categoria"
		Title:        content.Title,
		Slug:         newSlug,
		Description:  content.Description,
		Host:         content.Host,
		UserId:       user.UserId,
		ImageId:      "default", // We will check the if there is an image below
		ImageMaxSize: "small",   // We will check the if there is an image below
		LikeCount:    0,
		Creation:     time.Now(),
		LastUpdate:   time.Now(),
		Deleted:      false,
	}

	if img != nil {
		newContent.ImageId = img.ImageId
		newContent.ImageMaxSize = img.MaxSize
	}

	err = db.Insert(newContent)

	if err != nil {
		return nil, err
	}

	return newContent, nil
}
