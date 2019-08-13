package main

import (
	"time"

	karma "github.com/reconquest/karma-go"
)

func parseDate(date string) (int64, error) {

	var dateUnix int64

	destiny := karma.Describe("method", "parseDate")

	if date == "" {
		timeNow := time.Now()
		dateUnix = timeNow.Unix()
	} else {
		const RFC3339 = "2006-01-02 15:04"
		dateParse, err := time.Parse(RFC3339, date)
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
	const RFC3339 = "2006-01-02 15:04"
	date, err := time.Parse(RFC3339, value)
	if err != nil {
		return 0, karma.Format(err, "can't parse datetime '%s'", value)
	}

	return date.Unix(), nil
}
