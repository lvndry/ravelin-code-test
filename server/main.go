/*
* Code test January 2018
* Author: Landry Monga
**/
package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type websiteData map[string]Data //map[sess]Data

type Data struct {
	WebsiteURL         string
	SessionID          string
	ResizeFrom         Dimension
	ResizeTo           Dimension
	CopyAndPaste       map[string]bool
	FormCompletionTime int
}

type Dimension struct {
	Width  string `json:"width"`
	Height string `json:"height"`
}

const (
	PORT = ":8080"
)

var (
	Clients = make(map[string]*websiteData) //[url]map[sess]Data
	Version = "0.0.1"
)

func main() {
	var index bytes.Buffer
	pwd, _ := os.Getwd()
	index.WriteString(pwd)
	index.WriteString("/client/")

	mux := http.NewServeMux()

	// Routes
	mux.Handle("/", http.FileServer(http.Dir(index.String())))
	mux.HandleFunc("/copyandpaste", HandlerPaste)
	mux.HandleFunc("/new", HandlerNew)
	mux.HandleFunc("/resize", HandlerResize)
	mux.HandleFunc("/submit", HandlerSubmit)

	log.Println("Server starting at", time.Now().Format("15:04:05"), "on", Version)
	log.Println("Listening on port", PORT)
	log.Println(http.ListenAndServe(PORT, Middleware(mux)))
}

// Middleware
func Middleware(m http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.ServeHTTP(w, r)
	})
}

// HandlerNew Handler for a new connexion
func HandlerNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("The server only accpets POST requests")
		return
	}

	resp := getResp(w, r)

	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	if _, ok := Clients[url]; !ok {
		c := make(websiteData)
		Clients[url] = &c
	}

	wData, ok := Clients[url]
	if !ok {
		return
	}

	if sessData, ok := (*wData)[sess]; !ok {
		sessData.WebsiteURL = url
		sessData.SessionID = sess

		(*wData)[sess] = sessData
		Clients[url] = wData

		log.Println("Connexion from", url)
		log.Println(sessData)
	}
}

// HandlerPaste handler for the copyAndPaste event
func HandlerPaste(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("The server only accpets POST requests")
		return
	}

	resp := getResp(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	formID, ok := resp["formId"].(string) //Id of the field where the copy/paste append
	if !ok {
		log.Println("FormId is needed")
		return
	}

	paste, ok := resp["paste"].(bool) //Boolean: know if event is a paste or not
	if !ok {
		log.Println("Paste field is needed")
		return
	}

	wData, ok := Clients[url]
	if !ok {
		log.Println("Clients[", url, "] not found")
		return
	}

	sessData, ok := (*wData)[sess]
	if !ok {
		log.Println("sessData:", sessData, "not found")
		return
	}

	sessData.CopyAndPaste = make(map[string]bool)
	sessData.CopyAndPaste[formID] = paste

	(*wData)[sess] = sessData
	Clients[url] = wData

	log.Println(sessData)
}

// HandlerResize handler for the resize event
func HandlerResize(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("The server only accpets POST requests")
		return
	}

	resp := getResp(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	resizeFrom, err := getDimension(resp["resizeFrom"])
	if err != nil {
		log.Println("Original size is missing")
		return
	}

	resizeTo, err := getDimension(resp["resizeTo"])
	if err != nil {
		log.Println("Actual size is missing")
		return
	}

	wData, ok := Clients[url]
	if !ok {
		log.Println("Clients[", url, "] not found")
		return
	}

	sessData, ok := (*wData)[sess]
	if !ok {
		log.Println("sessData:", sessData, "not found")
		return
	}

	sessData.ResizeTo = resizeTo
	sessData.ResizeFrom = resizeFrom

	(*wData)[sess] = sessData
	Clients[url] = wData

	log.Println(sessData)
}

// HandlerSubmit handler called when submit button is clicked
func HandlerSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("The server only accpets POST requests")
		return
	}

	resp := getResp(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	time := int(resp["time"].(float64))

	wData, ok := Clients[url]
	if !ok {
		log.Println("Clients[", url, "] not found")
		return
	}

	sessData, ok := (*wData)[sess]
	if !ok {
		log.Println("sessData:", sessData, "not found")
		return
	}

	sessData.FormCompletionTime = time

	(*wData)[sess] = sessData
	Clients[url] = wData

	log.Println(sessData)
	Print(sessData)
}

// Returns the json response as a map
func getResp(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	var resp map[string]interface{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("err", err)
	}

	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		log.Println("err", err)
	}
	return resp
}

// Returns a Dimension struct
func getDimension(i interface{}) (d Dimension, err error) {
	b, err := json.Marshal(i)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(b, &d)
	if err != nil {
		log.Println(err)
	}
	return d, err
}

// Print prints a Data in the console
func Print(d Data) {
	output :=
		`
Structure for user {{.SessionID}} from {{.WebsiteURL}}

WebsiteURL: {{.WebsiteURL}}
SessionId: {{.SessionID}}
ResizeFrom: Width: {{.ResizeFrom.Width}}, Height: {{.ResizeFrom.Height}}
ResizeTo: Width: {{.ResizeTo.Width}}, Height: {{.ResizeTo.Height}} {{range $key, $value := .CopyAndPaste}}
CopyAndPaste: FormId: {{$key}}, Paste: {{$value}} {{end}}
FormCompletionTime: {{.FormCompletionTime}}

`
	tmpl, err := template.New("").Parse(output)
	if err != nil {
		log.Println(err)
	}
	tmpl.Execute(os.Stdout, d)
}
