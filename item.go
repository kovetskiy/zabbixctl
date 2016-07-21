package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reItemKeyParams = regexp.MustCompile(`\[([^\]]+)\]`)
)

type Item struct {
	ID         string `json:"itemid"`
	HostID     string `json:"hostid"`
	Name       string `json:"name"`
	ValueType  string `json:"value_type"`
	LastValue  string `json:"lastvalue"`
	LastChange string `json:"lastclock"`
	Key        string `json:"key_"`
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

func (item *Item) Format() string {
	name := item.Name

	match := reItemKeyParams.FindStringSubmatch(item.Key)
	if len(match) == 0 {
		return name
	}

	args := strings.Split(match[1], ",")
	for index, arg := range args {
		name = strings.Replace(name, fmt.Sprintf(`$%d`, index+1), arg, -1)
	}

	return name
}
