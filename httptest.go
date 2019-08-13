package main

import (
	"fmt"
	"strconv"
	"time"
)

// HTTPTestStep represents single step in the web scenario.
type HTTPTestStep struct {
	ID     string `json:"httpstepid"`
	TestID string `json:"httptestid"`
	URL    string `json:"url"`
}

// HTTPTest represents web scenario, which often used for simple step-by-step
// external monitoring of websites via HTTP.
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
	date, err := strconv.ParseInt(check.NextCheck, 10, 64)
	if err != nil {
		debugf("Error: %+v", err)
	}
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
