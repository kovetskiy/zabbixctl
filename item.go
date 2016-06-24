package main

import (
	"strconv"
	"time"
)

type Item struct {
	ID         string `json:"itemid"`
	HostID     string `json:"hostid"`
	Name       string `json:"name"`
	LastValue  string `json:"lastvalue"`
	LastChange string `json:"lastclock"`
}

func (item *Item) DateTime() string {
	if item.LastChange == "0" {
		return "-"
	}

	return item.date().Format("2006-01-02 15:04:05")
}

func (item *Item) date() time.Time {
	date, _ := strconv.ParseInt(item.LastChange, 10, 64)
	return time.Unix(date, 0)
}
