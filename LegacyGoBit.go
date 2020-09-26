package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
)
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {

	//file, err := os.Create(path)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		file, err = os.Create(path)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
type ImageStruct struct {
	Url string `json:"url"`
}
type Obit struct {
	Text string `json:"articleBody"`
	ImageObject ImageStruct `json:"image"`
}

func retrieveObit(url string) (Obit, error) {
	var returnObit Obit
	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		log.Fatal(err)
		return returnObit, err
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
		return returnObit, err
	}
	bodyString := string(bodyBytes)
	parts := strings.Split(bodyString, "articleBody\": \"")
	if len(parts) == 2 {
		parts2 := strings.Split(parts[1], "\",")
		if len(parts2) > 1 {
			returnObit.Text = parts2[0]
			//fmt.Println(returnObit.Text)
		}
	}
	photoBase := "https://cache.legacy.net/legacy/images/cobrands/ShreveportTimes/photos/"
	parts = strings.Split(bodyString, photoBase)
	if len(parts) > 1 {
		parts2 := strings.Split(parts[1], ".jpg")
		if len(parts2) > 1 {
			returnObit.ImageObject.Url = photoBase + parts2[0] + ".jpg"
		}
	}
	fmt.Println(returnObit)
	return returnObit, nil
}
func generateHTML(obits *[]Obit) (string){
	message := `<!DOCTYPE html>
	<html>
	<head>
	<style>
		table {
			font-family: arial, sans-serif;
			border-collapse: collapse;
			width: 100%;
		}

	td, th {
		border: 1px solid #dddddd;
		text-align: left;
		padding: 8px;
	}

tr:nth-child(even) {
		background-color: #dddddd;
	}
	</style>
	</head>
	<body>`
	message += "<table>"
	for i := 0; i < len(*obits); i++ {
		message += "<tr><td><img src=\"" + (*obits)[i].ImageObject.Url +
			"\"></td><td>" + (*obits)[i].Text + "</td></tr>"
	}
	message += "</table></html>"
	return message
}
func IsObitAlreadyRetrieved(url string, retrievedObits []string) bool {
	for _, line := range retrievedObits {
		//fmt.Println(i, line)
		if url == line {
			return true
		}
	}

	return false
}
func stripUrl(url string) string{
	//url should strip off  &fhid
	parts := strings.Split(url, "&fhid")
	for _, part := range parts {
		return part
	}
	return url
}
func webScrape(url string, retrievedObits []string) []string {
	var obitlinks []string
	obitlinksmap := make(map[string]struct{})
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			if strings.Contains(href, "obituary.aspx?"){
				href = stripUrl(href)
				//doing a map, because it guarantees uniqueness, and will roll into a slice in a hot minute (below)
				obitlinksmap[href] = struct{}{}
			}
		}
	})
	//roll through the unique links, and shove into the slice (array)
	for link := range(obitlinksmap) {
		//first, find out if we've already retrieved it...
		if !IsObitAlreadyRetrieved(link, retrievedObits) {
			//store, and work with it
			obitlinks = append(obitlinks, link)
			log.Printf("Processing %s", link)
			//StoreObit(db, link)
		} else {
			log.Printf("Already processed, so skipping %s", link)
		}

	}
	return obitlinks
}


func main() {

	to := flag.String("t","","destination Internet mail address")
	from := flag.String("f","","the sender's GMail address")
	pwd := flag.String("p","","the sender's password")
	url := flag.String("u", "", "the URL to use")
	flag.Usage=func() {
		fmt.Printf("Syntax:\n\tObit [flags]\nwhere flags are:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NFlag() != 4 {
		flag.Usage()
		return
	}

	workingDir, _ := os.Getwd()
	log.Print("Creating/using file storage in this directory: " + workingDir)

	retrievedObits, err := readLines(workingDir + "/obits.txt")
	if err != nil {
		log.Printf("readLines: %s", err)
	}

	/*
	db := InitDB(workingDir + "/obits.db")
	CreateTable(db)
	*/
	var obits = []Obit{}
	log.SetOutput(os.Stdout)

	fmt.Printf("Retrieving %s", *url)

	obitlinks := webScrape(*url, retrievedObits)

	if err := writeLines(obitlinks, workingDir + "/obits.txt"); err != nil {
		log.Fatalf("writeLines: %s", err)
	}

	for i := 0; i < len(obitlinks); i++ {
		//real work gets done here
		obit, err := retrieveObit(obitlinks[i])
		if err != nil {
			log.Printf("Not adding due to error: %s", err)
		} else {
			obits = append(obits, obit)
		}
	}
	if len(obits) > 0 {
		var html string
		html = generateHTML(&obits)

		auth := smtp.PlainAuth("", *from, *pwd, "smtp.gmail.com")
		// set headers for html email
		header := textproto.MIMEHeader{}
		header.Set(textproto.CanonicalMIMEHeaderKey("from"), *from)
		header.Set(textproto.CanonicalMIMEHeaderKey("to"), *to)
		header.Set(textproto.CanonicalMIMEHeaderKey("content-type"), "text/html; charset=UTF-8")
		header.Set(textproto.CanonicalMIMEHeaderKey("mime-version"), "1.0")
		header.Set(textproto.CanonicalMIMEHeaderKey("subject"), "Legacy Obits")

		// init empty message
		var buffer bytes.Buffer

		// write header
		for key, value := range header {
			buffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value[0]))
		}

		// write body
		buffer.WriteString(fmt.Sprintf("\r\n%s", html))

		adds := strings.Split(*to, ",")
		for i := 0; i < len(adds); i++ {
			err := smtp.SendMail("smtp.gmail.com:587", auth, *from,
				[]string{adds[i]}, buffer.Bytes())

			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
