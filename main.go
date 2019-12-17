package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
	"gopkg.in/djherbis/stow.v3"
)

var linkStore *stow.Store

const expireTime = time.Hour * 36 //links expire after 36 hours
var checkGex *regexp.Regexp = regexp.MustCompile("[^\\w]")

const minLength = 3
const maxLength = 20

//the shortcuts that can not be used
var blacklist = [...]string{
	"insert",
	"index",
	"about",
}

func main() {
	// Create a boltdb (bbolt fork) database
	db, err := bolt.Open("links.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	linkStore = stow.NewJSONStore(db, []byte("links"))

	//linkStore.Put("hello", Link{Name: "Dustin"})

	http.HandleFunc("/insert", insertHandler)
	http.HandleFunc("/", getHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		http.ServeFile(w, r, "./static/index.html")
	}
	split := strings.Split(r.URL.Path, "/")
	if len(split) <= 1 {
		fmt.Fprint(w, "Error here :(")
		return
	}

	var link Link
	linkStore.Get(split[1], &link)
	//insure that the link actually exists in the database and that it hasn't expired
	if link.Expire.IsZero() || link.Expire.Before(time.Now()) {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, link.URL, http.StatusFound)
}

//Handle insertion
func insertHandler(w http.ResponseWriter, r *http.Request) {
	//NOTE: Eventually this will handle serving the homepage file
	if r.Method != http.MethodPost {
		fmt.Fprint(w, "Can't do that. sorry")
		return
	}

	// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
	if err := r.ParseForm(); err != nil {
		fmt.Printf("ParseForm() err: %v", err)
		fmt.Fprint(w, "Error Parsing Form")
		return
	}

	name := r.FormValue("name")
	iurl := r.FormValue("url")
	if name == "" || iurl == "" {
		fmt.Fprint(w, "You need stuff in the form dumbo")
		return
	}

	//check if name is long enough and short enough
	if len(name) > maxLength || len(name) < minLength {
		fmt.Fprintf(w, "The shortcut has to be between %d and %d characters long", minLength, maxLength)
		return
	}

	//check if name does not contain any invalid chars
	if checkGex.MatchString(name) {
		fmt.Fprint(w, "You can only include numbers and letters in the shortcut")
		return
	}

	//Check that link does not exist and is safe to reasignn
	var t Link
	linkStore.Get(name, &t)
	if !t.Expire.IsZero() || !t.Expire.Add(expireTime).Before(time.Now()) || belongsToBlacklist(t.Name) {
		fmt.Fprint(w, "Sorry, the name is taken already, come back in a bit")
		return
	}

	//check if url is valid
	u, err := url.Parse(iurl)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		fmt.Fprint(w, "Some sort of error in your url. Make sure you are using https!")
		return
	}

	link := Link{Name: name, URL: iurl, Expire: time.Now().Add(expireTime)}

	linkStore.Put(name, link)
	fmt.Fprint(w, "OK")
}

//Link This is the link we will use
type Link struct {
	Name   string
	URL    string
	Expire time.Time
}

func belongsToBlacklist(lookup string) bool {
	for _, val := range blacklist {
		if val == lookup {
			return true
		}
	}
	return false
}
