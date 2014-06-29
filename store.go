package main

import (
	"strconv"
	"gopkg.in/redis.v1"
)

type Item struct {
	Id   string `json:"id"`
	Text string `json:"text" binding:"required"`
}

type ItemStore struct {
	Client *redis.Client
}

func (s *ItemStore) All() ([]Item, error) {
	results, err := s.Client.HGetAllMap("godo:items").Result()
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(results))
	for id, text := range results {
		items = append(items, Item{
			Id:   id,
			Text: text,
		})
	}
	return items, nil
}

func (s *ItemStore) Persist(item *Item) error {
	id, err := s.Client.Incr("godo:itemId").Result()
	if err != nil {
		return err
	}
	item.Id = strconv.FormatInt(id, 10)
	err = s.Client.HSet("godo:items", item.Id, item.Text).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *ItemStore) Get(id string) (*Item, error) {
	result, err := s.Client.HGet("godo:items", id).Result()
	if err != nil {
		return nil, err
	}
	item := &Item{
		Id:   id,
		Text: result,
	}
	return item, nil
}

func (s *ItemStore) Delete(id string) error {
	return s.Client.HDel("godo:items", id).Err()
}
