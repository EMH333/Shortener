package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var aDB *bolt.DB

var history = []byte("history")
var hits = []byte("hits")

//InitAnalytics initalizes analyitics database and creates buckets as needed
func InitAnalytics() {
	db, err := bolt.Open("analytics.db", 0600, nil)
	aDB = db
	if err != nil {
		log.Fatal(err)
	}

	//create the hitory and hits buckets
	err = aDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(hits)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists(history)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Initalized Analytics")
}

//CloseAnalytics closes stats database
func CloseAnalytics() {
	aDB.Close()
}

/*LogURLInsert inserts shortcut into history database. Format for value is creation_time#URL
 */
func LogURLInsert(link *Link) {
	err := aDB.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(history)

		//put history into bucket
		b.Put(getKeyFromLink(link), []byte(time.Now().Format(time.RFC3339)+"#"+link.URL))
		return nil
	})
	if err != nil {
		log.Print(err)
	}
}

//LogURLHit records a hit to a given link
func LogURLHit(link *Link) {
	err := aDB.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(hits)
		key := getKeyFromLink(link)

		//grab key, reset to 0 if it doesn't exist
		v := b.Get(key)
		if v == nil {
			v = make([]byte, 4)
			binary.LittleEndian.PutUint32(v, 0)
		}

		num := binary.LittleEndian.Uint32(v)
		num++

		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, num)

		b.Put(key, out)
		return nil
	})
	if err != nil {
		log.Print(err)
	}
}

//GetURLHits gets number of hits for a specific link (only needs accurate name and expiration)
func GetURLHits(link *Link) uint32 {
	var h uint32
	aDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(hits)
		key := getKeyFromLink(link)

		//grab key, reset to 0 if it doesn't exist
		v := b.Get(key)
		if v == nil {
			h = 0
		} else {
			h = binary.LittleEndian.Uint32(v)
		}
		return nil
	})
	return h
}

//GetMostRecentDetailsFromName parse together a basic link and creation time from history
func GetMostRecentDetailsFromName(name string) (Link, time.Time) {
	var link Link
	var creationTime time.Time
	aDB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(history).Cursor()

		prefix := []byte(name)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			e, _ := time.Parse(time.RFC3339, string(bytes.Split(k, []byte("_"))[1]))
			link = Link{
				Expire: e,
				Name:   name,
				URL:    string(bytes.Split(v, []byte("#"))[1]),
			}
			creationTime, _ = time.Parse(time.RFC3339, string(bytes.Split(v, []byte("#"))[0]))
		}
		return nil
	})
	return link, creationTime
}

//StatsHandler Serve a very simple stats page to confirm analytics are working
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	split := strings.Split(r.URL.Path, "/")
	if len(split) <= 1 {
		fmt.Fprint(w, "Error here :(")
		return
	}

	var name string = split[2]

	if checkGex.MatchString(name) {
		http.Error(w, "Invalid link name", http.StatusBadRequest)
		return
	}

	link, creationTime := GetMostRecentDetailsFromName(name)
	h := GetURLHits(&link)

	fmt.Fprintf(w, "Link has %d hits with a url of %s and a creation date of %s", h, link.URL, creationTime.Format(time.RFC822))
}

func getKeyFromLink(link *Link) []byte {
	return []byte(link.Name + "_" + link.Expire.Format(time.RFC3339))
}
