package main

import "testing"

func TestBlacklist(t *testing.T) {
	if belongsToBlacklist("test") {
		t.Fail()
	}
	if !belongsToBlacklist("stats") {
		t.Fail()
	}

	if checkGex.MatchString("lol") {
		t.Fail()
	}
	var invalidNames = [...]string{
		"./",
		";BREAK; PRINT lol;",
		"lol%",
		"123@12532efsd",
	}
	for _, val := range invalidNames {
		if !checkGex.MatchString(val) {
			t.Errorf("Missed %s", val)
		}
	}
}
