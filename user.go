package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rwcarlsen/goexif/exif"
	//"gopkg.in/mgo.v2"
	"bytes"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type User struct {
	UserId     bson.ObjectId `bson:"_id"`
	FirstName  string        `bson:"firstname"`
	LastName   string        `bson:"lastname"`
	Email      string        `bson:"email"`
	Password   string        `bson:"password"`
	GoogleId   string        `bson:"gId"`
	FacebookId string        `bson:"fId"`
	TwitterId  string        `bson:"tId"`
	Id         string        `bson:"userId"`
}

type Album struct {
	Id        bson.ObjectId `bson:"_id"`
	AlbumId   string        `bson:"albumId"`
	Owner     string        `bson:"owner"`
	OwnerName string        `bson:"ownerName"`
	Name      string        `bson:"albumname"`
}

type Photo struct {
	Id          bson.ObjectId  `bson:"_id"`
	PhotoId     string         `bson:"photoId"`
	Owner       string         `bson:"owner"`
	OwnerName   string         `bson:"ownerName"`
	AlbumId     string         `bson:"albumId"`
	URL         string         `bson:"url"`
	Description string         `bson:"description"`
	Location    Location       `bson:"location"`
	Timestamp   string         `bson:"timestamp"`
	Views       int            `bson:"views"`
	Tags        []string       `bson:"tags"`
	Comments    []PhotoComment `bson:"comments"`
}

type Video struct {
	Id          bson.ObjectId  `bson:"_id"`
	VideoId     string         `bson:"videoId"`
	Owner       string         `bson:"owner"`
	OwnerName   string         `bson:"ownerName"`
	AlbumId     string         `bson:"albumId"`
	URL         string         `bson:"url"`
	Description string         `bson:"description"`
	Location    Location       `bson:"location"`
	Timestamp   string         `bson:"timestamp"`
	Views       int            `bson:"views"`
	Tags        []string       `bson:"tags"`
	Comments    []PhotoComment `bson:"comments"`
}

type Location struct {
	Name      string `bson:"locationName"`
	Latitude  string `bson:"latitude"`
	Longitude string `bson:"longitude"`
}

type PhotoComment struct {
	User      string `bson:"userName"`
	UserId    string `bson:"userId"`
	Body      string `bson:"comment"`
	Timestamp string `bson:"time"`
}

type PhotoContainer struct {
	Categories []Photo
}

type Tag struct {
	Name   string  `bson:"tag"`
	Photos []Photo `bson:"photos"`
	Videos []Video `bson:"videos"`
}

type DisplayPhotos struct {
	Name   string  `bson:"name"`
	Photos []Photo `bson:"photos"`
	Videos []Video `bson:"videos"`
}

type FlickrTag struct {
	Tags struct {
		Source string `json:"source"`
		Tag    []struct {
			Content string `json:"_content"`
		} `json:"tag"`
	} `json:"tags"`
	Stat string `json:"stat"`
}

type FlickrImage struct {
	//PhotoID     string
	URL         string
	ImageName   string   `bson:"imageName"`
	Description string   `bson:"description"`
	TimeStamp   string   `bson:"timeStamp"`
	Keywords    []string `bson:"keywords"`
}

type News struct {
	Title        string `bson:"title"`
	URL          string `bson:"url"`
	ImageName    string
	ImageCaption string
	ImageUrl     string
	Images       []NewsImage `bson:"images"`
}

type NewsImage struct {
	Name    string `bson:"name"`
	Caption string `bson:"caption"`
}

type Response struct {
	Name    string
	Content string
}

var router = mux.NewRouter()

var authKey = []byte("NCDIUyd78DBCSJBlcsd783")

// Encryption Key
var encKey = []byte("nckdajKBDSY6778FDV891bdf")

var store = sessions.NewCookieStore(authKey, encKey)

var dbConnection *MongoDBConn

//add(dbConnection, name, password) ->add to db
//find(dbConnection, name) ->find in db

func main() {
	router.HandleFunc("/", handleIndex)
	router.HandleFunc("/login", handleLogin)
	router.HandleFunc("/logout", handleLogout)
	router.HandleFunc("/register", handleRegister)
	router.HandleFunc("/authenticated", handleAuthenticated)
	router.HandleFunc("/pictures", handlePictures)
	router.HandleFunc("/videos", handleVideos)
	router.HandleFunc("/flickrNews", handleFlickrNews)
	router.HandleFunc("/albums", handleAlbums)
	router.HandleFunc("/upload", handleUpload)
	router.HandleFunc("/uploadPic", uploadHandler)
	router.HandleFunc("/saveComment", handleComments)
	router.HandleFunc("/flickr", handleFlickr)
	router.HandleFunc("/tag", handleTag)
	router.HandleFunc("/tagCloud", createTagCloud)
	router.HandleFunc("/checkLogIn", checkLoggedIn)
	router.HandleFunc("/saveFile", handleSaveImage)
	router.HandleFunc("/createAlbum", handleCreateAlbum)
	router.HandleFunc("/user", handleUserProfile)
	router.HandleFunc("/upvote", handleUpvote)
	router.HandleFunc("/cmsHome", handleCms)
	router.HandleFunc("/delete", handleDelete)
	router.HandleFunc("/retrieveTag", handleMainTag)
	router.HandleFunc("/retrieveUser", handleMainUser)
	authenticateGoogle()
	authenticateFacebook()
	authenticateTwitter()

	dbConnection = NewMongoDBConn()
	_ = dbConnection.connect()

	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	picture := r.FormValue("pic")
	//album := r.FormValue("album")
	//owner := r.FormValue("owner")
	cType := r.FormValue("cType")

	fmt.Println("in delete", picture, cType)
	deleteFromOthers(dbConnection, picture, cType)

	if cType == "image" {
		err := dbConnection.session.DB(db_name).C("photos").Remove(bson.M{"photoId": picture})
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "No")
			return
		}
	} else {
		err := dbConnection.session.DB(db_name).C("videos").Remove(bson.M{"videoId": picture})
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "No")
			return
		}
	}

	deleteFromOthers(dbConnection, picture, cType)
	resp := "Yes_" + picture

	fmt.Fprintf(w, resp)

}
func flickrHelper(flickr string, start int) string {
	s := ""
	var doc bytes.Buffer
	images := getFlickrImages(flickr, start)
	flickrData := struct {
		Tag    string
		PageIP int
		PageIN int
		P      []FlickrImage
	}{
		flickr,
		start - 1,
		start + 1,
		images,
	}
	t, _ := template.ParseFiles("pictureHelper.html")
	t.Execute(&doc, flickrData)
	s = doc.String()
	return s
}

func newsHelper(guardian string, start int) string {
	s := ""
	var doc bytes.Buffer
	news := getNews(guardian, start)
	newsData := struct {
		Tag    string
		PageIP int
		PageIN int
		N      []News
	}{
		guardian,
		start - 1,
		start + 1,
		news,
	}
	t, _ := template.ParseFiles("newsHelper.html")
	t.Execute(&doc, newsData)
	s = doc.String()
	return s
}

func handleFlickrNews(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in handle ")
	r.ParseForm()
	request := r.FormValue("req")
	st := r.FormValue("start")
	cType := r.FormValue("cType")
	s := ""
	var doc bytes.Buffer
	var start int

	response := make([]Response, 2)
	if st == "" {
		start = 0
	} else {
		start, _ = strconv.Atoi(st)
	}

	fmt.Println(request, "in handle flickr 2")

	if request == "start" {
		t, _ := template.ParseFiles("flickrNews.html")
		t.Execute(&doc, nil)
		s = doc.String()
	} else if strings.HasPrefix(request, "getTags") {
		input := request[7:]
		fmt.Println("in else" + input)
		s = "Boxing 5,Tennis 7,Cycling 13,maximum 13"

	} else {

		guardian := request
		flickr := strings.ToLower(request)

		if cType == "image" {

			response[0].Name = "flickr"
			response[0].Content = flickrHelper(flickr, start)
			response[1].Name = "news"
			response[1].Content = ""

		} else if cType == "news" {

			response[1].Name = "news"
			response[1].Content = newsHelper(guardian, start)
			response[0].Name = "flickr"
			response[0].Content = ""
		} else {

			response[1].Name = "news"
			response[1].Content = newsHelper(guardian, start)
			response[0].Name = "flickr"
			response[0].Content = flickrHelper(flickr, start)
		}

		//fmt.Println(s)

		b, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
		}

		//fmt.Printf("%s", b)
		fmt.Fprintf(w, "%s", b)

	}

	fmt.Fprintf(w, s)
}

func handleVideos(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	start, _ := strconv.Atoi(r.FormValue("req"))
	limit := 3
	session, _ := store.Get(r, "cookie")
	currentUser := session.Values["user"].(string)
	u := findUser(dbConnection, currentUser)

	response := make([]Response, 1)
	s := ""
	var doc bytes.Buffer

	var videos []Video
	err := dbConnection.session.DB(db_name).C("videos").Find(bson.M{"owner": u.Id}).Skip(start * limit).Limit(limit).All(&videos)

	if len(videos) > 0 || start == 0 {
		data := struct {
			PageP int
			PageN int
			Video []Video
		}{
			start - 1,
			start + 1,
			videos,
		}

		//fmt.Println(photos)

		t, _ := template.ParseFiles("videosTemplate.html")
		if t == nil {
			fmt.Println("no template******************************************")
		}
		t.Execute(&doc, data)
		s = doc.String()
	} else {
		s = ""
	}

	response[0].Name = "ownVideos"
	response[0].Content = s

	//fmt.Println(s)

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Printf("%s", b)
	fmt.Fprintf(w, "%s", b)
	return

	//authenticated, _ := template.ParseFiles("videos.html")
	//authenticated.Execute(w, u)

}

func handleCms(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie")
	u := &User{}
	if session.Values["user"] == nil {
		u = &User{}
	} else if session.Values["user"].(string) == "" {
		u = &User{}
	} else {
		u = findUser(dbConnection, session.Values["user"].(string))
	}

	var p DisplayPhotos
	c := dbConnection.session.DB(db_name).C("displayPhotos")
	err := c.Find(bson.M{"name": "views"}).One(&p)
	if err != nil {
		fmt.Println("could not get most viewed photos")
	}

	var recent DisplayPhotos
	c = dbConnection.session.DB(db_name).C("displayPhotos")
	err = c.Find(bson.M{"name": "recent"}).One(&recent)
	if err != nil {
		fmt.Println("could not get most viewed photos")
	}

	flickrImages := getFlickrImages("boxing", 0)

	news := getNews("Boxing", 0)

	data := struct {
		P      DisplayPhotos
		R      DisplayPhotos
		Flickr []FlickrImage
		N      []News
		U      User
	}{
		p,
		recent,
		flickrImages,
		news,
		*u,
	}

	authenticated, _ := template.ParseFiles("cmsHome.html")
	authenticated.Execute(w, data)

}

func handleUpvote(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	picId := r.FormValue("picId")
	//albumId := r.FormValue("albumId")
	//owner := r.FormValue("picOwner")
	cType := r.FormValue("cType")

	//user := findUser(dbConnection, owner)

	//var al int
	photo := Photo{}
	video := Video{}

	if cType == "image" {
		err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"photoId": picId}).One(&photo)
		photo.Views = photo.Views + 1
		err = dbConnection.session.DB(db_name).C("photos").Update(bson.M{"photoId": picId}, bson.M{"$set": bson.M{"views": photo.Views}})
		if err != nil {
			fmt.Println("could not update photos in tag db")
			fmt.Println(err)
			fmt.Fprintf(w, "No")
		}
	} else {
		err := dbConnection.session.DB(db_name).C("videos").Find(bson.M{"videoId": picId}).One(&video)
		video.Views = video.Views + 1
		err = dbConnection.session.DB(db_name).C("videos").Update(bson.M{"videoId": picId}, bson.M{"$set": bson.M{"views": video.Views}})
		if err != nil {
			fmt.Println("could not update views in videos db")
			fmt.Println(err)
			fmt.Fprintf(w, "No")
		}
	}

	updateTagDB(photo, video, dbConnection)
	updateMostViewed(photo, video, dbConnection)
	updateMostRecent(photo, video, dbConnection)

	if cType == "image" {
		fmt.Fprintf(w, "Yes_"+strconv.Itoa(photo.Views))
	} else {
		fmt.Fprintf(w, "Yes_"+strconv.Itoa(video.Views))
	}

}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in login")
	r.ParseForm()
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	c := find(dbConnection, email)
	fmt.Println(c == nil)
	if c == nil {
		fmt.Fprintf(w, "No")
	} else {
		if c.Password == pass {
			session, _ := store.Get(r, "cookie")
			session.Values["user"] = c.Id
			session.Save(r, w)
			fmt.Println(c.FirstName)
			fmt.Fprintf(w, "Yes_"+c.FirstName)
		} else {
			fmt.Fprintf(w, "No")
		}
	}
}

func handleAuthenticated(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie")
	currentUser := session.Values["user"].(string)
	u := findUser(dbConnection, currentUser)
	var photos []Photo

	err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"owner": u.Id}).Skip(0).Limit(9).All(&photos)
	if err != nil {
		fmt.Println("could not get images from DB")
	}

	photoData := struct {
		FirstName string
		PageN     int
		PageP     int
		Photo     []Photo
	}{
		u.FirstName,
		1,
		1,
		photos,
	}

	authenticated, _ := template.ParseFiles("pictures2.html")
	authenticated.Execute(w, photoData)
}

func tagAlgo(u string) string {
	grepCmd, err := exec.Command("/bin/sh", "run.sh", u).Output()
	if err != nil {
		fmt.Println(err)
		fmt.Println("error")
	}

	return string(grepCmd)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie")
	if session.Values["user"] == nil {
		session.Values["user"] = ""
		session.Save(r, w)
	}

	authenticated, _ := template.ParseFiles("index.html")
	authenticated.Execute(w, session.Values["user"].(string))

}

func handleMainUser(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie")
	currentUser := session.Values["user"].(string)
	user := findUser(dbConnection, currentUser)

	u := r.URL.RawQuery

	var photos []Photo
	var videos []Video

	dbConnection.session.DB(db_name).C("photos").Find(bson.M{"owner": u}).Skip(0).Limit(3).All(&photos)
	dbConnection.session.DB(db_name).C("videos").Find(bson.M{"owner": u}).Skip(0).Limit(3).All(&videos)

	photoData := struct {
		FirstName string
		PageIN    int
		PageIP    int
		PageVN    int
		PageVP    int
		User      string
		Photo     []Photo
		Video     []Video
	}{
		user.FirstName,
		1,
		1,
		1,
		1,
		u,
		photos,
		videos,
	}

	authenticated, _ := template.ParseFiles("otherUsers.html")
	authenticated.Execute(w, photoData)

}

func handleUserProfile(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	t := r.FormValue("user")
	start := r.FormValue("start")
	cType := r.FormValue("cType")
	nModP := r.FormValue("nModP")
	//nModR := r.FormValue("nModR")
	st, _ := strconv.Atoi(start)
	nMod, _ := strconv.Atoi(nModP)
	nMod += 1
	limit := 3
	flag := true

	var photos []Photo
	var videos []Video
	var sti int
	var stv int
	var doc bytes.Buffer
	s := ""

	if t == "" {
		dbConnection.session.DB(db_name).C("photos").Find(bson.M{"owner": t}).Skip(0).Limit(3).All(&photos)
		dbConnection.session.DB(db_name).C("videos").Find(bson.M{"owner": t}).Skip(0).Limit(3).All(&videos)
		sti = 0
		stv = 0

	} else {

		if cType == "" {
			err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"owner": t}).Skip(st * limit).Limit(limit).All(&photos)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(photos)

			err = dbConnection.session.DB(db_name).C("videos").Find(bson.M{"owner": t}).Skip(st * limit).Limit(limit).All(&videos)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(photos)
			sti = 0
			stv = 0
		} else if cType == "image" {
			err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"owner": t}).Skip(st * limit).Limit(limit).All(&photos)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(photos)

			if len(photos) == 0 {
				flag = false
			}

			err = dbConnection.session.DB(db_name).C("videos").Find(bson.M{"owner": t}).Skip(nMod * limit).Limit(limit).All(&videos)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(photos)
			sti = st
			stv = nMod
		} else {
			err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"owner": t}).Skip(nMod * limit).Limit(limit).All(&photos)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(photos)

			err = dbConnection.session.DB(db_name).C("videos").Find(bson.M{"owner": t}).Skip(st * limit).Limit(limit).All(&videos)
			if err != nil {
				fmt.Println(err)
			}

			if len(videos) == 0 {
				flag = false
			}

			fmt.Println(photos)
			sti = nMod
			stv = st
		}
	}

	fmt.Println(t, " ", start, " ", cType, " ", nModP)

	if flag == true {

		photoData := struct {
			PageIN int
			PageIP int
			PageVN int
			PageVP int
			User   string
			Photo  []Photo
			Video  []Video
		}{
			sti + 1,
			sti - 1,
			stv + 1,
			stv - 1,
			t,
			photos,
			videos,
		}

		temp, _ := template.ParseFiles("photoVideoTemplate.html")
		if temp == nil {
			fmt.Println("no template******************************************")
		}

		temp.Execute(&doc, photoData)
		s = doc.String()
	} else {
		s = ""
	}

	fmt.Fprintf(w, s)

}

func handleCreateAlbum(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.FormValue("name")
	//description := r.FormValue("description")

	session, _ := store.Get(r, "cookie")
	currentUser := session.Values["user"].(string)
	c := findUser(dbConnection, currentUser)

	albumId := createAlbum(name, c.Id, c.FirstName+" "+c.LastName, dbConnection)

	fmt.Fprintf(w, albumId)
}

func handleSaveImage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	file, _, err := r.FormFile("uploadData")

	if err != nil {
		fmt.Println(w, err)
		fmt.Fprintf(w, "No")
		return
	}

	id := bson.NewObjectId()
	fileName := "./resources/images/userUploaded/" + id.Hex()

	dst, err := os.Create(fileName)
	defer dst.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprintf(w, "No")
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprintf(w, "No")
		return
	}

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Yes_"+fileName+"_nil_nil")
		return
	}

	x, err := exif.Decode(f)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Yes_"+fileName+"_nil_nil")
		return
	}

	if x == nil {
		fmt.Println("x is nil")
		fmt.Fprintf(w, "Yes_"+fileName+"_nil_nil")

	} else {

		lat, long, err := x.LatLong()
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Yes_"+fileName+"_nil_nil")
		} else {

			fmt.Fprintf(w, "Yes_"+fileName+"_"+strconv.FormatFloat(lat, 'f', -1, 64)+"_"+strconv.FormatFloat(long, 'f', -1, 64))
		}
	}

}

func handleLogout(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "cookie")
	session.Values["user"] = ""
	session.Save(r, w)
	u := findUser(dbConnection, session.Values["user"].(string))

	if u == nil {
		u = &User{}
	}
	http.Redirect(w, r, "/cmsHome", http.StatusFound)
}

func checkLoggedIn(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie")

	if session.Values["user"] == nil {
		fmt.Fprintf(w, "No")
	} else if session.Values["user"].(string) == "" {
		fmt.Fprintf(w, "No")
	} else {
		message := "Yes," + findUser(dbConnection, session.Values["user"].(string)).FirstName
		fmt.Fprintf(w, message)
	}
}

func createTagCloud(w http.ResponseWriter, r *http.Request) {
	result := getAllTags(dbConnection)
	var tags string
	var max = 0
	for tag := range result {
		if len(result[tag].Photos)+len(result[tag].Videos) > max {
			max = len(result[tag].Photos) + len(result[tag].Videos)
		}

		tags += result[tag].Name + " " + strconv.Itoa(len(result[tag].Photos)+len(result[tag].Videos)) + ","
	}
	tags += "maximum " + strconv.Itoa(max)
	fmt.Fprintf(w, tags)

}

func handleTag(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	t := r.FormValue("tag")
	start := r.FormValue("start")
	cType := r.FormValue("cType")
	nModP := r.FormValue("nModP")
	//nModR := r.FormValue("nModR")
	st, _ := strconv.Atoi(start)
	nMod, _ := strconv.Atoi(nModP)
	nMod += 1
	limit := 3

	var photos []Photo
	var videos []Video
	var sti int
	var stv int
	var doc bytes.Buffer
	s := ""
	flag := true

	if cType == "" {
		tag := findByTag(dbConnection, t)
		photos = tag.Photos
		if len(tag.Photos) == 0 {
			photos = nil
		}
		videos = tag.Videos
		if len(tag.Videos) == 0 {
			videos = nil
		}

		sti = 0
		stv = 0
	} else if cType == "image" {
		err := dbConnection.session.DB(db_name).C("tags").Find(bson.M{"tag": t}).Skip(st * limit).Limit(limit).All(&photos)
		if err != nil {
			fmt.Println(err)
		}

		if len(photos) == 0 {
			flag = false
		}
		//fmt.Println(photos)

		err = dbConnection.session.DB(db_name).C("tags").Find(bson.M{"tag": t}).Skip(nMod * limit).Limit(limit).All(&videos)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(photos)
		sti = st
		stv = nMod
	} else {
		err := dbConnection.session.DB(db_name).C("tags").Find(bson.M{"tag": t}).Skip(nMod * limit).Limit(limit).All(&photos)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(photos)

		err = dbConnection.session.DB(db_name).C("tags").Find(bson.M{"tag": t}).Skip(st * limit).Limit(limit).All(&videos)
		if err != nil {
			fmt.Println(err)
		}

		if len(videos) == 0 {
			flag = false
		}
		fmt.Println(photos)
		sti = nMod
		stv = st

	}

	fmt.Println(t, " ", start, " ", cType, " ", nModP)

	if flag == true {

		photoData := struct {
			PageIN int
			PageIP int
			PageVN int
			PageVP int
			Tag    string
			Photo  []Photo
			Video  []Video
		}{
			sti + 1,
			sti - 1,
			stv + 1,
			stv - 1,
			t,
			photos,
			videos,
		}
		fmt.Println(photoData)

		temp, _ := template.ParseFiles("tagContentTemplate.html")
		if temp == nil {
			fmt.Println("no template******************************************")
		}

		temp.Execute(&doc, photoData)
		s = doc.String()
	} else {
		s = ""
	}

	fmt.Println(s)

	fmt.Fprintf(w, s)
}

func handleMainTag(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie")
	currentUser := session.Values["user"].(string)
	user := findUser(dbConnection, currentUser)

	u := r.URL.RawQuery
	tag := findByTag(dbConnection, u)

	photoData := struct {
		FirstName string
		PageIN    int
		PageIP    int
		PageVN    int
		PageVP    int
		Tag       string
		Photo     []Photo
		Video     []Video
	}{
		user.FirstName,
		1,
		1,
		1,
		1,
		tag.Name,
		tag.Photos,
		tag.Videos,
	}

	authenticated, _ := template.ParseFiles("taggedPictures2.html")
	authenticated.Execute(w, photoData)
}

func handleFlickr(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url1 := r.FormValue("url1")
	url2 := r.FormValue("url2")
	tag := r.FormValue("tags")
	var tags = ""

	tagList := strings.Split(tag, ",")
	for tag := range tagList {

		res, err := http.Get(url1 + tagList[tag] + url2)
		if err != nil {
			fmt.Println(err.Error())
		}

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			fmt.Println(err.Error())
		}

		resp := string(body)
		resp = resp[14 : len(resp)-1]

		var data FlickrTag
		err = json.Unmarshal([]byte(resp), &data)
		if err != nil {
			fmt.Println("error when unmarshalling JSON response from Flickr" + err.Error())
		}

		for tag := 0; tag < 4; tag++ {
			tags = tags + data.Tags.Tag[tag].Content + ","
		}

	}

	if tags == "" {
		tags = tagAlgo(tag)
	}

	fmt.Fprintf(w, tags)

}

func handleRegister(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	fname := r.FormValue("first")
	lname := r.FormValue("last")
	email := r.FormValue("email")
	pass := r.FormValue("pass")

	id := bson.NewObjectId()

	createDefaultAlbum(dbConnection, id.Hex(), fname+" "+lname)

	newUser := User{id, fname, lname, email, pass, "", "", "", id.Hex()}
	add(dbConnection, newUser)

	c := find(dbConnection, email)

	if c == nil {
		fmt.Fprintf(w, "No")
	} else {

		session, _ := store.Get(r, "cookie")
		session.Values["user"] = c.Id
		session.Save(r, w)
		fmt.Fprintf(w, "Yes")
	}

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	image := r.FormValue("imageURL")
	caption := r.FormValue("caption")
	cType := r.FormValue("contentType")
	album := r.FormValue("albumSelect")
	loc := r.FormValue("location")
	lng := r.FormValue("lng")
	lat := r.FormValue("lat")
	locationN := r.FormValue("locality")
	/*if loc == "" {
		lng = ""
		lat = ""
		locationN = ""
	} */

	fmt.Println(loc, " loc")
	fmt.Println(lng, " lng")
	fmt.Println(lat, " lat")
	fmt.Println(locationN, " locN")

	streetN := r.FormValue("formatted_address")
	streetN = strings.Split(streetN, ",")[0]
	tags := r.FormValue("tagList")

	var location = *new(Location)
	if lat != "" && lng != "" {
		location = Location{streetN + ", " + locationN, lat, lng}
	}

	t := make([]string, 0)
	if tags != "" {
		t = parseTags(tags, image)
	}

	id := bson.NewObjectId()
	p := Photo{}
	v := Video{}

	session, _ := store.Get(r, "cookie")
	user := session.Values["user"].(string)
	currentUser := findUser(dbConnection, user)

	//c := uploadToAlbum(cType, image, caption, album, lng, lat, streetN+", "+locationN, tags, currentUser)

	if cType == "image" {
		p = Photo{id, id.Hex(), currentUser.Id, currentUser.FirstName + " " + currentUser.LastName, album, image, caption, location, time.Now().Local().Format("2006-01-02"), 0, t, make([]PhotoComment, 1)}
		addTags(dbConnection, t, p, Video{})
		c := dbConnection.session.DB(db_name).C("photos")
		err := c.Insert(p)
		if err != nil {
			panic(err)
		}
	} else {
		v = Video{id, id.Hex(), currentUser.Id, currentUser.FirstName + " " + currentUser.LastName, album, image, caption, location, time.Now().Local().Format("2006-01-02"), 0, t, make([]PhotoComment, 1)}
		addTags(dbConnection, t, Photo{}, v)
		c := dbConnection.session.DB(db_name).C("videos")
		err := c.Insert(v)
		if err != nil {
			panic(err)
		}

	}

	insertInMostRecent(p, v, dbConnection)

}

func parseTags(tags string, filename string) []string {
	tags = strings.ToLower(tags)
	s := strings.Split(tags, ",")

	return s
}

func getPictures(collName string, field string, userId string, templateName string, start int) string {

	s := ""
	var doc bytes.Buffer
	var photos []Photo
	limit := 3
	err := dbConnection.session.DB(db_name).C(collName).Find(bson.M{field: userId}).Skip(start * limit).Limit(limit).All(&photos)

	if err != nil {
		fmt.Println(err)
	}
	if len(photos) > 0 || start == 0 {
		photoData := struct {
			PageN int
			PageP int
			Photo []Photo
		}{
			start + 1,
			start - 1,
			photos,
		}

		t, _ := template.ParseFiles(templateName)
		if t == nil {
			fmt.Println("no template******************************************")
		}

		t.Execute(&doc, photoData)
		s = doc.String()
	} else {
		s = ""
	}
	return s

}

func handlePictures(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	start := r.FormValue("req")
	s, _ := strconv.Atoi(start)
	fmt.Println(s)
	session, _ := store.Get(r, "cookie")
	currentUser := session.Values["user"].(string)
	response := make([]Response, 1)

	response[0].Name = "ownPictures"
	response[0].Content = getPictures("photos", "owner", currentUser, "pictureTemplate.html", s)

	//fmt.Println(s)

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Printf("%s", b)
	fmt.Fprintf(w, "%s", b)
	return

}

func handleAlbums(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	query := r.FormValue("albumId")
	start := r.FormValue("start")
	cType := r.FormValue("cType")
	nModP := r.FormValue("nModP")
	//nModR := r.FormValue("nModR")
	st, _ := strconv.Atoi(start)
	nMod, _ := strconv.Atoi(nModP)
	nMod += 1
	limit := 3

	fmt.Println(query, " ", start, " ", cType, " ", nModP)

	session, _ := store.Get(r, "cookie")
	user := session.Values["user"].(string)
	currentUser := findUser(dbConnection, user)
	response := make([]Response, 1)
	s := ""
	var doc bytes.Buffer
	flag := true

	if query == "" {
		var albums []Album
		err := dbConnection.session.DB(db_name).C("albums").Find(bson.M{"owner": currentUser.Id}).All(&albums)
		if err != nil {
			fmt.Println(err)
		}
		data := struct {
			Page   string
			Albums []Album
		}{
			"0",
			albums,
		}

		fmt.Println(data)
		t, _ := template.ParseFiles("albumTemplate.html")
		if t == nil {
			fmt.Println("no template******************************************")
		}
		t.Execute(&doc, data)
		s = doc.String()
		response[0].Name = "ownAlbums"
	} else {
		fmt.Println("in eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", query)
		var photos []Photo
		var videos []Video
		var sti int
		var stv int

		if cType == "" {
			err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"albumId": query}).Skip(st * limit).Limit(limit).All(&photos)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(photos)

			err = dbConnection.session.DB(db_name).C("videos").Find(bson.M{"albumId": query}).Skip(st * limit).Limit(limit).All(&videos)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(photos)
			sti = 0
			stv = 0
		} else if cType == "image" {
			err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"albumId": query}).Skip(st * limit).Limit(limit).All(&photos)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(photos)

			if len(photos) == 0 {
				flag = false
			}

			err = dbConnection.session.DB(db_name).C("videos").Find(bson.M{"albumId": query}).Skip(nMod * limit).Limit(limit).All(&videos)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(photos)
			sti = st
			stv = nMod
		} else {
			err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"albumId": query}).Skip(nMod * limit).Limit(limit).All(&photos)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(photos)

			err = dbConnection.session.DB(db_name).C("videos").Find(bson.M{"albumId": query}).Skip(st * limit).Limit(limit).All(&videos)
			if err != nil {
				fmt.Println(err)
			}

			if len(videos) == 0 {
				flag = false
			}
			fmt.Println(photos)
			sti = nMod
			stv = st
		}

		fmt.Println(query, " ", start, " ", cType, " ", nModP)

		if flag == true {

			photoData := struct {
				PageIN  int
				PageIP  int
				PageVN  int
				PageVP  int
				AlbumId string
				Photo   []Photo
				Video   []Video
			}{
				sti + 1,
				sti - 1,
				stv + 1,
				stv - 1,
				query,
				photos,
				videos,
			}

			temp, _ := template.ParseFiles("albumDetailTemplate.html")
			if temp == nil {
				fmt.Println("no template******************************************")
			}

			temp.Execute(&doc, photoData)
			s = doc.String()
		} else {
			s = ""
		}

		response[0].Name = "albumDetail"
	}

	response[0].Content = s

	//fmt.Println(s)

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Printf("%s", b)
	fmt.Fprintf(w, "%s", b)
	return

}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie")
	u := session.Values["user"].(string)
	currentUser := findUser(dbConnection, u)
	response := make([]Response, 1)
	s := ""
	var doc bytes.Buffer

	var albums []Album
	err := dbConnection.session.DB(db_name).C("albums").Find(bson.M{"owner": currentUser.Id}).All(&albums)
	if err != nil {
		fmt.Println(err)
	}

	data := struct {
		Albums []Album
	}{
		albums,
	}
	t, _ := template.ParseFiles("upload2.html")
	t.Execute(&doc, data)
	s = doc.String()

	fmt.Println(s)
	response[0].Name = "upload"
	response[0].Content = s

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "%s", b)
}

func handleComments(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	comment := r.FormValue("comment")
	picture := r.FormValue("pic")
	cType := r.FormValue("cType")

	fmt.Println("comment", comment)
	fmt.Println("comment", picture)
	fmt.Println("comment", cType)

	session, _ := store.Get(r, "cookie")
	user2 := session.Values["user"].(string)

	currentUser := findUser(dbConnection, user2)
	com := PhotoComment{currentUser.FirstName + " " + currentUser.LastName, currentUser.Id, comment, time.Now().Local().Format("2006-01-02")}

	photo := Photo{}
	video := Video{}

	if cType == "image" {
		err := dbConnection.session.DB(db_name).C("photos").Find(bson.M{"photoId": picture}).One(&photo)
		photo.Comments = append(photo.Comments, com)
		err = dbConnection.session.DB(db_name).C("photos").Update(bson.M{"photoId": picture}, bson.M{"$set": bson.M{"comments": photo.Comments}})
		if err != nil {
			fmt.Println("could not update photos in tag db")
			fmt.Println(err)
			fmt.Fprintf(w, "No")
		}
	} else {
		err := dbConnection.session.DB(db_name).C("videos").Find(bson.M{"videoId": picture}).One(&video)
		video.Comments = append(video.Comments, com)
		err = dbConnection.session.DB(db_name).C("videos").Update(bson.M{"videoId": picture}, bson.M{"$set": bson.M{"comments": video.Comments}})
		if err != nil {
			fmt.Println("could not update views in videos db")
			fmt.Println(err)
			fmt.Fprintf(w, "No")
		}
	}

	updateTagDB(photo, video, dbConnection)
	updateMostRecent(photo, video, dbConnection)
	updateMostViewed(photo, video, dbConnection)

	response := com.Body + "_" + com.User + "_" + com.Timestamp
	fmt.Fprintf(w, "Yes_"+response)
}
