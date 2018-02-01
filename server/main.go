/*
* Code test January 2018
* Author: Landry Monga
**/
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	Clients = make(map[string]*websiteData) //[url]map[sess]Data
	Version = "0.0.1"
)

type websiteData map[string]Data //map[sess]Data

const (
	PORT = ":8080"
)

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
	Height string `json:"heigth"`
}

func main() {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/copyandpaste", Handlerpaste)
	mux.HandleFunc("/new", HandlerNew)
	mux.HandleFunc("/resize", HandlerResize)
	mux.HandleFunc("/submit", HandlerSubmit)

	log.Println("Server starting at", time.Now().Format("15:04:05"), "on", Version)
	log.Println("Listening on port", PORT)
	log.Println(http.ListenAndServe(PORT, MiddleWare(mux)))
}

func MiddleWare(m http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			log.Println("The server only accpets POST requests")
			return
		}
		m.ServeHTTP(w, r)
	})
}

// Handler New
func HandlerNew(w http.ResponseWriter, r *http.Request) {
	resp := getResp(w, r)
	url := resp["websiteURL"].(string)

	if _, ok := Clients[url]; !ok {
		Clients[url] = new(websiteData)
	}
	log.Println("Connexion from", url)
}

// Paste Handler
func Handlerpaste(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	sesData := (*wData)[sess]
	if !ok {
		return
	}

	sesData.CopyAndPaste[formID] = paste
	log.Println((*Clients[url])[sess])
}

// Resize Handler
func HandlerResize(w http.ResponseWriter, r *http.Request) {
	resp := getResp(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	resizeFrom, err := getDimension(resp["resizeFrom"])
	if err != nil {
		log.Println("Original size is missing")
	}

	resizeTo, err := getDimension(resp["resizeTo"])
	if err != nil {
		log.Println("Actual size is missing")
	}

	wData, ok := Clients[url]
	if !ok {
		return
	}

	sesData := (*wData)[sess]
	if !ok {
		return
	}

	sesData.ResizeTo = resizeTo
	sesData.ResizeFrom = resizeFrom
}

// Submit Handler
func HandlerSubmit(w http.ResponseWriter, r *http.Request) {
	resp := getResp(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	time := int(resp["time"].(float64))

	wData, ok := Clients[url]
	if !ok {
		return
	}

	sessData, ok := (*wData)[sess]
	if !ok {
		return
	}

	sessData.FormCompletionTime = time
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
