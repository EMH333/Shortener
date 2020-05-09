package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

func TestMostRecentRetrival(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "testingdb*.db")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	InitAnalytics(tmpfile.Name())
	defer CloseAnalytics()

	var link1 Link = Link{
		Name:   "test",
		URL:    "https://incorrect.com",
		Expire: time.Now(),
	}
	var link2 Link = Link{
		Name:   "test",
		URL:    "https://correct.com",
		Expire: time.Now(),
	}

	LogURLInsert(&link1)
	LogURLInsert(&link2)

	result, _ := GetMostRecentDetailsFromName("test")

	if result.URL != link2.URL {
		t.Error("Incorrect link record selected")
	}
}

func TestURLHits(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "testingdb*.db")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	InitAnalytics(tmpfile.Name())
	defer CloseAnalytics()

	var link = Link{
		Expire: time.Now(),
		Name:   "test",
		URL:    "https://example.com",
	}

	var expectedHits uint32 = 10
	for i := 0; i < int(expectedHits); i++ {
		LogURLHit(&link)
	}

	if GetURLHits(&link) != expectedHits {
		t.Error("Incorrect number of hits recorded")
	}
}
