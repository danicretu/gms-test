package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type MongoDBConn struct {
	session *mgo.Session
}

func NewMongoDBConn() *MongoDBConn {
	return &MongoDBConn{}
}

func (m *MongoDBConn) connect() *mgo.Session {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}

	fmt.Println("connect")
	m.session = session
	return m.session
}

func add(m *MongoDBConn, user User) {

	c := m.session.DB("gmsTry").C("user")
	err := c.Insert(user)
	if err != nil {
		panic(err)
	}

}

func addTags(m *MongoDBConn, tags []string, photo Photo) {

	c := m.session.DB("gmsTry").C("tags")
	for tag := range tags {
		result := Tag{}
		err := c.Find(bson.M{"tag": tags[tag]}).One(&result)
		if err != nil {
			fmt.Println("error while finding tag", tags[tag])
			result.Name = tags[tag]
			result.Photos = make([]Photo, 1)
			result.Photos[0] = photo
			err2 := c.Insert(result)
			if err2 != nil {
				fmt.Println("error while adding tag ", result.Name)
			}
		} else {
			result.Photos = append(result.Photos, photo)
			err = c.Update(bson.M{"tag": result.Name}, bson.M{"$set": bson.M{"photos": result.Photos}})
			if err != nil {
				fmt.Println("error while trying to update tag ", result.Name)
			}
		}

	}

}

func findByTag(m *MongoDBConn, tag string) *Tag {
	c := m.session.DB("gmsTry").C("tags")
	result := Tag{}
	err := c.Find(bson.M{"tag": tag}).One(&result)
	if err != nil {
		fmt.Println("Error finding tag")
		fmt.Println(err)
		return nil
	}
	fmt.Println(result)
	return &result
}

func getAllTags(m *MongoDBConn) []Tag {
	c := m.session.DB("gmsTry").C("tags")
	var result []Tag
	err := c.Find(nil).All(&result)
	if err != nil {
		fmt.Println("Error finding tag")
		fmt.Println(err)
		return nil
	}

	fmt.Println(result)
	return result
}

func find(m *MongoDBConn, email string) *User {
	result := User{}
	c := m.session.DB("gmsTry").C("user")
	fmt.Println(email)
	err := c.Find(bson.M{"email": email}).One(&result)
	if err != nil {
		return nil
	}

	return &result
}

func findUser(m *MongoDBConn, id string) *User {
	result := User{}
	c := m.session.DB("gmsTry").C("user")
	fmt.Println(id)
	err := c.Find(bson.M{"userId": id}).One(&result)
	if err != nil {
		return nil

	}

	return &result
}

func createDefaultAlbum(ownerId string, ownerName string, picture string) []Album {
	albums := make([]Album, 1)
	id := bson.NewObjectId()

	photos := make([]Photo, 1)
	photos[0] = createDefaultPhoto(ownerId, ownerName, picture)

	album := Album{id, id.Hex(), ownerId, ownerName, "Default Album", "", photos}
	albums[0] = album

	return albums
}

func createDefaultPhoto(ownerId string, ownerName string, picture string) Photo {
	id := bson.NewObjectId()
	var loc Location
	loc = Location{"Glasgow", "", ""}
	var photo Photo
	var url string

	if picture == "" {
		url = "./resources/images/userUploaded/default.gif"
	} else {
		url = picture
	}
	photo = Photo{id, id.Hex(), ownerId, ownerName, url, "Default Picture", loc, time.Now().Local().Format("2006-01-02"), 0, 0, make([]string, 1), make([]PhotoComment, 1)}

	return photo
}
