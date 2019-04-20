package main

import (
	"strconv"
	"time"
)

// Maintenance struct
type Maintenance struct {
	ID          string       `json:"maintenanceid"`
	Name        string       `json:"name"`
	Type        string       `json:"maintenance_type"`
	Description string       `json:"description"`
	Since       string       `json:"active_since"`
	Till        string       `json:"active_till"`
	Hosts       []Host       `json:"hosts"`
	Timeperiods []Timeperiod `json:"timeperiods"`
	Groups      []Group      `json:"groups"`
}

// Maintenances struct
type Maintenances struct {
	ID []string `json:"maintenanceids"`
}

func (maintenance *Maintenance) GetString() string {
	return maintenance.ID + " " +
		maintenance.Name + " " + maintenance.Description
}

// https://www.zabbix.com/documentation/3.4/manual/api/reference/maintenance/object
func (maintenance *Maintenance) GetTypeCollect() string {
	if maintenance.Type == "0" {
		return "COLLECT"
	}
	return "NO COLLECT"
}

func (maintenance *Maintenance) GetStatus() string {
	now := time.Now().Format("2006-01-02 15:04:05")
	maintenanceSince := maintenance.GetDateTime(maintenance.Since)
	maintenanceTill := maintenance.GetDateTime(maintenance.Till)

	if now >= maintenanceSince && now < maintenanceTill {
		return "ACTIVE"
	}
	return "EXPIRED"
}

func (maintenance *Maintenance) GetDateTime(unixtime string) string {
	date, _ := strconv.ParseInt(unixtime, 10, 64)
	return time.Unix(date, 0).Format("2006-01-02 15:04:05")
}
