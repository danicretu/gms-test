package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type User struct {
	UserId         bson.ObjectId `bson:"_id"`
	FirstName      string        `bson:"firstname"`
	LastName       string        `bson:"lastname"`
	Email          string        `bson:"email"`
	Password       string        `bson:"password"`
	ProfilePicture string        `bson:"profilepic"`
	Albums         []Album       `bson:"albums"`
	GoogleId       string        `bson:"gId"`
	FacebookId     string        `bson:"fId"`
	Id             string        `bson:"userId"`
}

type Album struct {
	Id         bson.ObjectId `bson:"_id"`
	AlbumId    string        `bson:"albumId"`
	Owner      string        `bson:"owner"`
	OwnerName  string        `bson:"ownerName"`
	Name       string        `bson:"albumname"`
	Desciption string        `bson:"description"`
	Photo      []Photo       `bson:"photos"`
}

type Photo struct {
	Id          bson.ObjectId  `bson:"_id"`
	PhotoId     string         `bson:"photoId"`
	Owner       string         `bson:"owner"`
	OwnerName   string         `bson:"ownerName"`
	URL         string         `bson:"url"`
	Description string         `bson:"description"`
	Location    Location       `bson:"location"`
	Timestamp   string         `bson:"timestamp"`
	Upvote      int            `bson:"upvote"`
	Downvote    int            `bson:"downvote"`
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

var dbConnection *MongoDBConn

var currentUser *User

//add(dbConnection, name, password) ->add to db
//find(dbConnection, name) ->find in db

func login() {

	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/authenticated", handleAuthenticated)
	http.HandleFunc("/pictures", handlePictures)
	http.HandleFunc("/albums", handleAlbums)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/uploadPic", uploadHandler)
	http.HandleFunc("/saveComment", handleComments)
	http.HandleFunc("/auth", authenticate)
	http.HandleFunc("/flickr", handleFlickr)
	http.HandleFunc("/tag", handleTag)
	http.HandleFunc("/tagCloud", createTagCloud)
	authenticateGoogle()
	authenticateFacebook()

	dbConnection = NewMongoDBConn()
	_ = dbConnection.connect()

	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

}

func createTagCloud(w http.ResponseWriter, r *http.Request) {
	result := getAllTags(dbConnection)
	var tags string
	var max = 0
	for tag := range result {
		if len(result[tag].Photos) > max {
			max = len(result[tag].Photos)
		}
		tags += result[tag].Name + " " + strconv.Itoa(len(result[tag].Photos)) + ","
	}
	tags += "maximum " + strconv.Itoa(max)
	fmt.Println(tags)
	fmt.Fprintf(w, tags)

}

func handleTag(w http.ResponseWriter, r *http.Request) {
	url := r.URL.RawQuery
	fmt.Println(url)
	tag := findByTag(dbConnection, url)
	fmt.Println(tag)

	data := struct {
		T Tag
		U User
	}{
		*tag,
		*currentUser,
	}

	fmt.Println("**************************************************")
	fmt.Println(data.T.Photos)

	displaySameTagPhoto, _ := template.ParseFiles("taggedPictures.html")
	displaySameTagPhoto.Execute(w, data)

}

func handleFlickr(w http.ResponseWriter, r *http.Request) {

	url := r.URL.RawQuery

	res, err := http.Get(url)
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

	var tags string
	for tag := range data.Tags.Tag {
		tags = tags + data.Tags.Tag[tag].Content + ","
	}

	fmt.Println(tags)

	fmt.Fprintf(w, tags)

}

func authenticate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("---------------------------------------------------------")
	fmt.Println(currentUser)
	authenticated, _ := template.ParseFiles("authenticated-test.html")
	authenticated.Execute(w, currentUser)

}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	c := find(dbConnection, email)

	if c == nil {
		fmt.Fprintf(w, "No")
	} else {
		if c.Password == pass {
			currentUser = c
			fmt.Fprintf(w, "Yes")
		} else {
			fmt.Fprintf(w, "No")
		}
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {

	fname := r.FormValue("first")
	lname := r.FormValue("last")
	email := r.FormValue("email")
	pass := r.FormValue("password")
	pass2 := r.FormValue("confirmPassword")

	id := bson.NewObjectId()

	albums := createDefaultAlbum(id.Hex(), fname+" "+lname, "")

	newUser := User{id, fname, lname, email, pass, albums[0].Photo[0].URL, albums, "", "", id.Hex()}

	if pass == pass2 {
		fmt.Println(email)
		add(dbConnection, newUser)

		c := find(dbConnection, email)
		currentUser = c
		authenticated, _ := template.ParseFiles("authenticated.html")
		authenticated.Execute(w, currentUser)
	}

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	caption := r.FormValue("caption")
	fmt.Println(caption)

	tags := r.FormValue("tagList")
	fmt.Println(tags)

	file, header, err := r.FormFile("file")
	fmt.Println(header.Filename)

	if err != nil {
		fmt.Println(w, err)
		return
	}

	defer file.Close()

	id := bson.NewObjectId()
	fileName := "./resources/images/userUploaded/" + id.Hex()

	dst, err := os.Create(fileName)
	defer dst.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uploadToAlbum(fileName, id, caption, tags)
	authenticated, _ := template.ParseFiles("authenticated.html")
	authenticated.Execute(w, currentUser)
}

func uploadToAlbum(filename string, id bson.ObjectId, caption string, tags string) {

	user := find(dbConnection, currentUser.Email)
	location := Location{"Glasgow", "1", "2"}

	t := parseTags(tags, filename)

	photo := Photo{id, id.Hex(), currentUser.Id, currentUser.FirstName + " " + currentUser.LastName, filename, caption, location, time.Now().Local().Format("2006-01-02"), 0, 0, t, make([]PhotoComment, 1)}
	addTags(dbConnection, t, photo)

	fmt.Println(user)

	fmt.Println("***********")

	user.Albums[0].Photo = append(user.Albums[0].Photo, photo)
	currentUser.Albums[0].Photo = append(currentUser.Albums[0].Photo, photo)

	fmt.Println(user)
	fmt.Println("***********************")
	fmt.Println(currentUser)
	err := dbConnection.session.DB("gmsTry").C("user").Update(bson.M{"email": user.Email}, bson.M{"$set": bson.M{"albums": user.Albums}})
	if err != nil {

		fmt.Println("***************")
		fmt.Println("error while trying to update2")
	}

}

func parseTags(tags string, filename string) []string {
	tags = strings.ToLower(tags)
	s := strings.Split(tags, ",")
	fmt.Println(s)

	return s
}

func handleAuthenticated(w http.ResponseWriter, r *http.Request) {
	authenticated, _ := template.ParseFiles("authenticated.html")
	authenticated.Execute(w, currentUser)
}

func handlePictures(w http.ResponseWriter, r *http.Request) {
	fmt.Println(currentUser.Albums[0].Photo)

	authenticated, _ := template.ParseFiles("pictures.html")
	authenticated.Execute(w, currentUser)

}

func handleAlbums(w http.ResponseWriter, r *http.Request) {

}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	authenticated, _ := template.ParseFiles("upload.html")
	authenticated.Execute(w, currentUser)
}

func handleComments(w http.ResponseWriter, r *http.Request) {
	comment := r.FormValue("comment")
	picture := r.FormValue("pictureNumber")
	album := r.FormValue("albumNumber")
	owner := r.FormValue("owner")

	var user *User
	user = findUser(dbConnection, owner)

	fmt.Println(comment, picture, owner, album)

	fmt.Println(owner)
	fmt.Println(comment, picture, owner, album)
	fmt.Println(user)
	fmt.Println(comment, picture, owner, album)

	var al int

	for i := range user.Albums {
		if user.Albums[i].AlbumId == album {
			al = i
			break
		}
	}

	var pic int

	for i := range user.Albums[al].Photo {
		if user.Albums[al].Photo[i].PhotoId == picture {
			pic = i
			break
		}
	}

	fmt.Println(al, pic)

	com := PhotoComment{currentUser.FirstName + " " + currentUser.LastName, currentUser.Id, comment, time.Now().Local().Format("2006-01-02")}

	fmt.Println(com)

	user.Albums[al].Photo[pic].Comments = append(user.Albums[al].Photo[pic].Comments, com)

	fmt.Println(user)
	err := dbConnection.session.DB("gmsTry").C("user").Update(bson.M{"_id": user.UserId}, bson.M{"$set": bson.M{"albums": user.Albums}})
	if err != nil {
		panic(err)
	}

	authenticated, _ := template.ParseFiles("pictures.html")
	authenticated.Execute(w, currentUser)
}
