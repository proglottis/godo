package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/redis.v1"
)

type ItemsHandler struct {
	Store *ItemStore
}

func (h ItemsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.Index(w)
	case "POST":
		h.Create(w, r)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func (h ItemsHandler) Index(w http.ResponseWriter) {
	items, err := h.Store.All()
	if err != nil {
		panic(err)
	}
	JSON(w, items)
}

func (h ItemsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var item Item
	dec := json.NewDecoder(r.Body)
	for {
		if err := dec.Decode(&item); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	err := h.Store.Persist(&item)
	if err != nil {
		panic(err)
	}
	JSON(w, item)
}

type ItemHandler struct {
	Store *ItemStore
}

func (h ItemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	id := path[len(path)-1]
	switch r.Method {
	case "DELETE":
		h.Delete(w, id)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func (h ItemHandler) Delete(w http.ResponseWriter, id string) {
	err := h.Store.Delete(id)
	if err != nil {
		panic(err)
	}
	JSON(w, nil)
}

func main() {
	client := redis.NewTCPClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()
	store := &ItemStore{Client: client}

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))
	http.Handle("/items", ItemsHandler{Store: store})
	http.Handle("/items/", ItemHandler{Store: store})

	port := os.Getenv("PORT")
	if len(port) < 1 {
		port = "3000"
	}
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		log.Fatal(err)
	}
}

func JSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	w.Write(bytes)
}
