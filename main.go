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
const maxURLLength = 2000

//the shortcuts that can not be used
//Note that this doesn't require the .html because dots aren't allowed
var blacklist = [...]string{
	"insert",
	"index",
	"about",
}

func main() {
	// Create a boltdb (bbolt fork) database
	db, err := bolt.Open("links.db", 0600, nil)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	linkStore = stow.NewJSONStore(db, []byte("links"))

	http.HandleFunc("/insert", insertHandler)
	http.HandleFunc("/", getHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	//serve home page if requested
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		http.ServeFile(w, r, "./static/index.html")
		return
	}

	//serve static files as well
	if r.URL.Path == "/script.js" {
		http.ServeFile(w, r, "./static/script.js")
		return
	}
	if r.URL.Path == "/style.css" {
		http.ServeFile(w, r, "./static/style.css")
		return
	}

	//else try to serve the link
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
	if r.Method != http.MethodPost {
		fmt.Fprint(w, "Can't do that. sorry")
		return
	}

	// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
	if err := r.ParseForm(); err != nil {
		fmt.Printf("ParseForm() err: %v", err) //NOTE remove for production
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
	if err != nil || u.Scheme != "https" || u.Host == "" || len(iurl) > maxURLLength {
		fmt.Fprint(w, "Some sort of error in your url. Make sure you are using https!")
		return
	}

	//TODO add forever links that permananetly point to a url

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
