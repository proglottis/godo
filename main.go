package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"gopkg.in/redis.v1"
	"net/http"
	"strconv"
)

type Item struct {
	Id   string `json:"id"`
	Text string `json:"text" binding:"required"`
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
		Layout:     "layout",
	}))
	client := redis.NewTCPClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()
	m.Map(client)

	m.Get("/", Root)

	// Items
	m.Get("/items", GetItems)
	m.Post("/items", binding.Bind(Item{}), CreateItem)
	m.Delete("/items/:id", DeleteItem)

	m.Run()
}

func Root(r render.Render) {
	r.HTML(http.StatusOK, "index", nil)
}

func GetItems(r render.Render, c *redis.Client) {
	results, err := c.HGetAllMap("godo:items").Result()
	if err != nil {
		panic(err)
	}
	items := make([]Item, 0, len(results))
	for id, text := range results {
		items = append(items, Item{
			Id:   id,
			Text: text,
		})
	}
	r.JSON(http.StatusOK, items)
}

func CreateItem(r render.Render, c *redis.Client, item Item) {
	id, err := c.Incr("godo:itemId").Result()
	if err != nil {
		panic(err)
	}
	item.Id = strconv.FormatInt(id, 10)
	err = c.HSet("godo:items", item.Id, item.Text).Err()
	if err != nil {
		panic(err)
	}
	r.JSON(http.StatusOK, item)
}

func DeleteItem(r render.Render, params martini.Params, c *redis.Client) {
	err := c.HDel("godo:items", params["id"]).Err()
	if err != nil {
		panic(err)
	}
	r.JSON(http.StatusOK, nil)
}
