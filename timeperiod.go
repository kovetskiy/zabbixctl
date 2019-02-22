package main

import (
	"strconv"
	"time"
)

// Timeperiod struct
type Timeperiod struct {
	ID        string `json:"timeperiodid"`
	TypeID    string `json:"timeperiod_type"`
	Every     string `json:"every"`
	Month     string `json:"month"`
	DayOfWeek string `json:"dayofweek"`
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	Period    string `json:"period"`
	StartDate string `json:"start_date"`
}

// https://www.zabbix.com/documentation/3.4/manual/api/reference/maintenance/object#time_period
func (timeperiod *Timeperiod) GetType() string {
	switch timeperiod.TypeID {
	case "2":
		return "DAILY"
	case "3":
		return "WEEKLY"
	case "4":
		return "MONTHLY"
	default:
		return "ONCE"

	}
}

func (timeperiod *Timeperiod) GetStartDate() string {
	date, _ := strconv.ParseInt(timeperiod.StartDate, 10, 64)
	return time.Unix(date, 0).Format("2006-01-02 15:04:05")
}

func (timeperiod *Timeperiod) GetPeriodMinute() int64 {
	period, _ := strconv.ParseInt(timeperiod.Period, 10, 64)
	return (period / 60)
}
