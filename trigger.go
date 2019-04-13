package main

import (
	"strconv"
	"strings"
	"time"
)

type Trigger struct {
	ID          string     `json:"triggerid"`
	Description string     `json:"description"`
	Hostname    string     `json:"host"`
	Value       string     `json:"value"`
	Comments    string     `json:"comments"`
	Functions   []Function `json:"functions"`
	LastChange  string     `json:"lastchange"`
	LastEvent   struct {
		ID           string `json:"eventid"`
		Acknowledged string `json:"acknowledged"`
	} `json:"lastEvent"`
	Hosts []struct {
		Hostid string `json:"hostid"`
		Name   string `json:"name"`
	} `json:"hosts"`
	Priority string `json:"priority"`
}

func (trigger *Trigger) String() string {
	return trigger.LastEvent.ID + " " +
		trigger.Hostname + " " + trigger.Description
}

func (trigger *Trigger) GetHostName() string {
	if len(trigger.Hosts) > 0 {
		return trigger.Hosts[0].Name
	}
	return "<missing>"
}

func (trigger *Trigger) StatusAcknowledge() string {
	if trigger.LastEvent.Acknowledged == "1" {
		return "ACK"
	}

	return "NACK"
}

func (trigger *Trigger) StatusProblem() string {
	if trigger.Value == "1" {
		return "PROBLEM"
	}

	return "OK"
}

func (trigger *Trigger) Severity() Severity {
	value, _ := strconv.Atoi(trigger.Priority)
	return Severity(value)
}

func (trigger *Trigger) DateTime() string {
	return trigger.date().Format("2006-01-02 15:04:05")
}

func (trigger *Trigger) Age() string {
	date := time.Since(trigger.date())

	var (
		seconds = int(date.Seconds()) % 60
		minutes = int(date.Minutes()) % 60
		hours   = int(date.Hours())
		days    = hours / 24
		months  = days / 30.
	)

	var units []string

	units = addUnit(units, months, "mon")
	units = addUnit(units, days%7, "d")
	units = addUnit(units, hours%24, "h")
	units = addUnit(units, minutes, "m")
	units = addUnit(units, seconds, "s")

	return strings.Join(units, " ")
}

func (trigger *Trigger) date() time.Time {
	date, _ := strconv.ParseInt(trigger.LastChange, 10, 64)
	return time.Unix(date, 0)
}

func addUnit(units []string, value int, unit string) []string {
	if value > 1 {
		units = append(units, strconv.Itoa(value)+unit)
	}

	return units
}
