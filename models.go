package main

import (
	"time"
)

type User struct {
	UserId     string    `db:"userid" json:"userid"`       // *PK max: 20
	UserName   string    `db:"username" json:"username"`   // *UQ max: 20
	PicId      string    `db:"picid" json:"picid"`         // *PK max: 20
	FullName   string    `db:"fullname" json:"fullname"`   // *UQ max: 20
	LikeCount  int       `db:"likecount" json:"likecount"` // default: 0
	Creation   time.Time `db:"creation" json:"-"`          // *NN
	LastUpdate time.Time `db:"lastupdate" json:"-"`        // *NN
	Deleted    bool      `db:"deleted" json:"-"`           // default: 0
	Admin      bool      `db:"admin" json:"-"`             // default: 0
}

type Pic struct {
	PicId    string    `db:"picid" json:"picid"`       // *PK max: 20
	Creation time.Time `db:"creation" json:"creation"` // *NN
	Deleted  bool      `db:"deleted" json:"-"`         // default: 0
}

type Profile struct {
	ProfileId    string    `db:"profileid" json:"profileid"`       // *PK max: 40
	UserId       string    `db:"userid" json:"userid"`             // *FK max: 20
	UserName     string    `db:"username" json:"username"`         // *NN max: 20
	Email        string    `db:"email" json:"email"`               // *NN max: 40
	FullName     string    `db:"fullname" json:"fullname"`         // *NN max: 40
	Gender       string    `db:"gender" json:"gender"`             // *NN enum: male, female
	ProfileUrl   string    `db:"profileurl" json:"profileurl"`     // *NN max: 100
	Language     string    `db:"language" json:"language"`         //  max: 10, default: pt_BR
	Verified     bool      `db:"verified" json:"verified"`         // default: 0
	FirstName    string    `db:"firstname" json:"firstname"`       // *NN max: 20
	LastName     string    `db:"lastname" json:"lastname"`         // *NN max: 20
	SourceUpdate time.Time `db:"sourceupdate" json:"sourceupdate"` // *NN
	AccessToken  string    `db:"accesstoken" json:"accesstoken"`   // *NN max: 100
	RefreshToken string    `db:"refreshtoken" json:"refreshtoken"` // *NN max: 100
	Scope        string    `db:"scope" json:"scope"`               // *NN max: 40
	TokenExpiry  time.Time `db:"tokenexpiry" json:"tokenexpiry"`   // *NN Time token spiries
	Creation     time.Time `db:"creation" json:"creation"`         // *NN
	LastUpdate   time.Time `db:"lastupdate" json:"-"`              // *NN
}

type Token struct {
	TokenId  string    `db:"tokenid" json:"tokenid"` // *PK max: 20
	UserId   string    `db:"userid" json:"userid"`   // *FK max: 20
	Creation time.Time `db:"creation" json:"-"`      // *NN
}

type Category struct {
	CategoryId   string `db:"categoryid" json:"categoryid"`     // *PK max: 20
	CategoryName string `db:"categoryname" json:"categoryname"` //  max: 20
	CategorySlug string `db:"categoryslug" json:"categoryslug"` // *UQ max: 20 Index
	LikeCount    int    `db:"likecount" json:"likecount"`       // default: 0
}

// Save in public/img/ImageId-Size.png {"small", "medium", "large"}
type Image struct {
	ImageId  string    `db:"imageid" json:"imageid"`   // *PK max: 20userpicid
	MaxSize  string    `db:"maxsize" json:"maxsize"`   // *NN ENUM('small', 'medium', 'large')
	Creation time.Time `db:"creation" json:"creation"` // *NN
	Deleted  bool      `db:"deleted" json:"-"`         // default: 0
}

type Url struct {
	UrlId     string    `db:"urlid" json:"urlid"`         // *PK max: 5
	FullUrl   string    `db:"fullurl" json:"fullurl"`     //  max: 20
	UserId    string    `db:"userid" json:"userid"`       // *FK max: 20
	Creation  time.Time `db:"creation" json:"creation"`   // *NN
	ViewCount int       `db:"viewcount" json:"viewcount"` // default: 0
}

type Content struct {
	ContentId   string    `db:"contentid" json:"contentid"`     // *PK max: 20
	UrlId       string    `db:"urlid" json:"urlid"`             // *FK max: 5
	CategoryId  string    `db:"categoryid" json:"categoryid"`   // *FK max: 20
	Title       string    `db:"title" json:"title"`             // *NN  max: 255 (250)
	Slug        string    `db:"slug" json:"slug"`               // *NN  max: 255 (250)
	Description string    `db:"description" json:"description"` // *NN  max: 255
	Host        string    `db:"host" json:"host"`               // *NN  max: 20
	UserId      string    `db:"userid" json:"userid"`           // *FK max: 20
	ImageId     string    `db:"imageid" json:"imageid"`         // *FK max: 20
	LikeCount   int       `db:"likecount" json:"likecount"`     // default: 0
	Creation    time.Time `db:"creation" json:"creation"`       // *NN
	LastUpdate  time.Time `db:"lastupdate" json:"lastupdate"`   // *NN
	Deleted     bool      `db:"deleted" json:"-"`               // default: 0
}

// It's a view
type FullContent struct {
	ContentId    string    `db:"contentid" json:"contentid"`       // *PK max: 20
	UrlId        string    `db:"urlid" json:"urlid"`               // *FK max: 5
	CategoryId   string    `db:"categoryid" json:"categoryid"`     // *FK max: 20
	Title        string    `db:"title" json:"title"`               // *NN  max: 255 (250)
	Slug         string    `db:"slug" json:"slug"`                 // *NN  max: 255 (250)
	Description  string    `db:"description" json:"description"`   // *NN  max: 255
	Host         string    `db:"host" json:"host"`                 // *NN  max: 20
	UserId       string    `db:"userid" json:"userid"`             // *FK max: 20
	ImageId      string    `db:"imageid" json:"imageid"`           // *FK max: 20
	LikeCount    int       `db:"likecount" json:"likecount"`       // default: 0
	Creation     time.Time `db:"creation" json:"creation"`         // *NN
	UserFullName string    `db:"userfullname" json:"userfullname"` // *UQ max: 20 *USER
	UserLikes    int       `db:"userlikes" json:"userlikes"`       // default: 0
	UserPicId    string    `db:"userpicid" json:"userpicid"`       // *PK max: 20
	ImageMaxSize string    `db:"imagemaxsize" json:"imagemaxsize"` // *NN ENUM('small', 'medium', 'large')
	ViewCount    int       `db:"viewcount" json:"viewcount"`       // default: 0
	CategoryName string    `db:"categoryname" json:"categoryname"` //  max: 20
	CategorySlug string    `db:"categoryslug" json:"categoryslug"` // *UQ max: 20 Index
	ILike        bool      `db:"ilike" json:"ilike"`               //  *NOT IN THE VIEW
}

type ContentLike struct {
	ContentId  string    `db:"contentid" json:"contentid"`   // *PK max: 20
	UserId     string    `db:"userid" json:"userid"`         // *PK max: 20
	Creation   time.Time `db:"creation" json:"creation"`     // *NN
	LastUpdate time.Time `db:"lastupdate" json:"lastupdate"` // *NN
	Deleted    bool      `db:"deleted" json:"-"`             // default: 0
}
