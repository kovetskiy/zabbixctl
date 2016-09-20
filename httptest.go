package main

import (
	"fmt"
	"strconv"
	"time"
)

type HTTPTestStep struct {
	ID     string `json:"httpstepid"`
	TestID string `json:"httptestid"`
	URL    string `json:"url"`
}

type HTTPTest struct {
	ID         string `json:"httptestid"`
	HostID     string `json:"hostid"`
	Name       string `json:"name"`
	Delay      string `json:"delay"`
	NextCheck  string `json:"nextcheck"`
	TemplateID string `json:"templateid"`

	Steps []HTTPTestStep `json:"steps"`
}

func (check *HTTPTest) DateTime() string {
	if check.NextCheck == "0" {
		return "-"
	}

	return check.date().Format("2006-01-02 15:04:05")
}

func (check *HTTPTest) date() time.Time {
	date, _ := strconv.ParseInt(check.NextCheck, 10, 64)
	return time.Unix(date, 0)
}

func (check *HTTPTest) Format() string {
	return fmt.Sprintf(
		"%s (%d steps every %s seconds)\t%s (next)\t",
		check.Name,
		len(check.Steps),
		check.Delay,
		check.DateTime(),
	)
}
