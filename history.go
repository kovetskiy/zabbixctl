package main

import (
	"fmt"
	"strconv"
	"time"
)

type History struct {
	ItemID string      `json:"itemid"`
	Value  interface{} `json:"value"`
	Clock  string      `json:"clock"`
}

type ItemHistory struct {
	Item
	History
}

func (history *History) String() string {
	return fmt.Sprint(history.Value)
}

func (history *History) date() time.Time {
	date, _ := strconv.ParseInt(history.Clock, 10, 64)
	return time.Unix(date, 0)
}

func (history *History) DateTime() string {
	return history.date().Format("2006-01-02 15:04:05")
}
