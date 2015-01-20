package main

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io/ioutil"
	"net/http"
)

var (
	codeG  = ""
	tokenG = ""
)

var oauthCfgG = &oauth.Config{
	//TODO: put your project's Client Id here.  To be got from https://code.google.com/apis/console
	ClientId: "410632888135-3joclvqh2br18gq8mqa64klfo1vnae0k.apps.googleusercontent.com",

	//TODO: put your project's Client Secret value here https://code.google.com/apis/console
	ClientSecret: "-XsCO0Rlpd7EawBFLdlRs-aQ",

	//For Google's oauth2 authentication, use this defined URL
	AuthURL: "https://accounts.google.com/o/oauth2/auth",

	//For Google's oauth2 authentication, use this defined URL
	TokenURL: "https://accounts.google.com/o/oauth2/token",

	//To return your oauth2 code, Google will redirect the browser to this page that you have defined
	//TODO: This exact URL should also be added in your Google API console for this project within "API Access"->"Redirect URIs"
	RedirectURL: "http://localhost:8080/oauth2callback",

	//This is the 'scope' of the data that you are asking the user's permission to access. For getting user's info, this is the url that Google has defined.
	Scope: "https://www.googleapis.com/auth/userinfo.profile",
}

type UserG struct {
	Id          string
	Name        string
	Given_Name  string
	Family_Name string
	Link        string
	Picture     string
	Locale      string
}

//This is the URL that Google has defined so that an authenticated application may obtain the user's info in json format
const profileInfoURLG = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"

func authenticateGoogle() {
	http.HandleFunc("/authorizeGoogle", handleAuthorizeG)

	//Google will redirect to this page to return your code, so handle it appropriately
	http.HandleFunc("/oauth2callback", handleOAuth2CallbackG)
}

func handleAuthorizeG(w http.ResponseWriter, r *http.Request) {
	//Get the Google URL which shows the Authentication page to the user
	url := oauthCfgG.AuthCodeURL("")

	fmt.Print(url)

	//redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

// Function that handles the callback from the Google server
func handleOAuth2CallbackG(w http.ResponseWriter, r *http.Request) {
	//Get the code from the response
	code := r.FormValue("code")

	t := &oauth.Transport{Config: oauthCfgG}

	// Exchange the received code for a token
	t.Exchange(code)

	//now get user data based on the Transport which has the token
	resp, err := t.Client().Get(profileInfoURLG)

	if err != nil {
		panic(err.Error())
	}

	var user UserG

	body, err := ioutil.ReadAll(resp.Body)

	fmt.Print(string(body))

	json.Unmarshal(body, &user)
	fmt.Println(user.Id)
	fmt.Println(user.Given_Name)
	fmt.Println(user.Family_Name)
	fmt.Println(user.Picture)
	fmt.Println(user.Locale)
	fmt.Println(user.Name)

	var existing *User
	dbConnection.session.DB("gmsTry").C("user").Find(bson.M{"gId": user.Id}).One(&existing)

	fmt.Println(existing)

	if existing != nil {
		currentUser = existing
	} else {

		id := bson.NewObjectId()
		albums := createDefaultAlbum(id.Hex(), user.Given_Name+" "+user.Family_Name, user.Picture)

		newUser := User{id, user.Given_Name, user.Family_Name, "", "", user.Picture, albums, user.Id, "", id.Hex()}
		add(dbConnection, newUser)
		currentUser = &newUser

	}

	authenticated, _ := template.ParseFiles("authenticated.html")
	authenticated.Execute(w, currentUser)
}