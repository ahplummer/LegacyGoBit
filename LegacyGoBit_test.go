package main

import "testing"

func TestStripUrl(t *testing.T) {
	shorturl := "https://www.legacy.com/obituaries/shreveporttimes/obituary.aspx?n=emilie-scott-wallace&pid=193667567"
	longurl := "https://www.legacy.com/obituaries/shreveporttimes/obituary.aspx?n=emilie-scott-wallace&pid=193667567&fhid=blahblah"
	res := stripUrl(shorturl)
	if res != shorturl {
		t.Errorf("Failed to return just the basic Url")
	}
	res = stripUrl(longurl)
	if res != shorturl {
		t.Errorf("Failed to pull out the base Url")
	}
}


func TestIsObitAlreadyRetrieved(t *testing.T) {
	testline := "this is my test line"
	var testslice = []string{}
	testslice = append(testslice, testline)
	testslice = append(testslice, "random string")
	res := IsObitAlreadyRetrieved(testline, testslice)
	if !res {
		t.Errorf("Failed to find item at top of list")
	}

	testslice = []string{}
	testslice = append(testslice, "random string")
	testslice = append(testslice, testline)
	res = IsObitAlreadyRetrieved(testline, testslice)
	if !res {
		t.Errorf("Failed to find item at bottom of list")
	}

	testslice = []string{}
	testslice = append(testslice, "random string")
	testslice = append(testslice, "another random string")
	res = IsObitAlreadyRetrieved(testline, testslice)
	if res {
		t.Errorf("Failed: test line is supposedly in list, but not.")
	}
}
