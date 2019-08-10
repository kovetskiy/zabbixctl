package main

import (
	"time"

	karma "github.com/reconquest/karma-go"
	"github.com/simplereach/timeutils"
)

func parseDate(date string) (int64, error) {

	var dateUnix int64

	destiny := karma.Describe("method", "parseDate")

	if date == "" {
		timeNow := time.Now()
		dateUnix = timeNow.Unix()
	} else {
		dateParse, err := timeutils.ParseDateString(date)
		if err != nil {
			return dateUnix, destiny.Describe(
				"error", err,
			).Describe(
				"date", date,
			).Reason(
				"can't convert date to unixtime",
			)
		}
		dateUnix = dateParse.Unix()
	}
	return dateUnix, nil
}

func parseDateTime(value string) (int64, error) {
	date, err := timeutils.ParseDateString(value)
	if err != nil {
		return 0, karma.Format(err, "can't parse datetime '%s'", value)
	}

	return date.Unix(), nil
}
