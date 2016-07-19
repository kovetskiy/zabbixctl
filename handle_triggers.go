package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/simplereach/timeutils"
	"github.com/zazab/hierr"
)

type ExtendedOutput int

const (
	ExtendedOutputNone ExtendedOutput = iota
	ExtendedOutputValue
	ExtendedOutputDate
	ExtendedOutputAll
)

func handleTriggers(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {
	var (
		acknowledge    = args["--acknowledge"].(bool)
		words, pattern = parseSearchQuery(args["<pattern>"].([]string))
		confirmation   = !args["--noconfirm"].(bool)
		extended       = ExtendedOutput(args["--extended"].(int))

		table = tabwriter.NewWriter(os.Stdout, 1, 4, 2, ' ', 0)
	)

	if len(words) > 0 {
		return fmt.Errorf(
			"unexpected command line token '%s', "+
				"use '/%s' for searching triggers",
			words[0], words[0],
		)
	}

	params, err := parseParams(args)
	if err != nil {
		return err
	}

	var triggers []Trigger

	err = withSpinner(
		":: Requesting information about statuses of triggers",
		func() error {
			triggers, err = zabbix.GetTriggers(params)
			return err
		},
	)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't obtain zabbix triggers",
		)
	}

	var history = make(map[string]ItemHistory)

	if extended != ExtendedOutputNone {
		history, err = getTriggerItemsHistory(zabbix, triggers)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't obtain history for items of triggers`,
			)
		}
	}

	debugln("* showing triggers table")
	if pattern != "" {
		debugf("** searching %s", pattern)
	}

	identifiers := []string{}
	for _, trigger := range triggers {
		if pattern != "" && !matchPattern(pattern, trigger.String()) {
			continue
		}

		fmt.Fprintf(
			table,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s",
			trigger.LastEvent.ID, trigger.DateTime(),
			trigger.Severity(),
			trigger.StatusProblem(),
			trigger.StatusAcknowledge(),
			trigger.Hostname,
			trigger.Description,
		)

		if len(trigger.Functions) > 0 {
			if last, ok := history[trigger.Functions[0].ItemID]; ok {
				if extended >= ExtendedOutputValue {
					fmt.Fprintf(table, "\t%s", last.History.String())
				}

				if extended >= ExtendedOutputDate {
					fmt.Fprintf(table, "\t%s", last.History.DateTime())
				}

				if extended >= ExtendedOutputAll {
					fmt.Fprintf(table, "\t%s", last.Item.Format())
				}
			}
		}

		fmt.Fprint(table, "\n")

		identifiers = append(identifiers, trigger.LastEvent.ID)
	}

	err = table.Flush()
	if err != nil {
		return err
	}

	if !acknowledge || len(identifiers) == 0 {
		return nil
	}

	if confirmation {
		if !confirmAcknowledge() {
			return nil
		}
	}

	err = withSpinner(
		":: Acknowledging specified triggers",
		func() error {
			return zabbix.Acknowledge(identifiers)
		},
	)

	if err != nil {
		return hierr.Errorf(
			err,
			"can't acknowledge triggers %s",
			identifiers,
		)
	}

	fmt.Fprintln(os.Stderr, ":: Acknowledged")

	return nil
}

func getTriggerItemsHistory(
	zabbix *Zabbix,
	triggers []Trigger,
) (map[string]ItemHistory, error) {
	history := map[string]ItemHistory{}

	itemIDs := []string{}
	for _, trigger := range triggers {
		if len(trigger.Functions) > 0 {
			itemIDs = append(itemIDs, trigger.Functions[0].ItemID)
		}
	}

	items, err := zabbix.GetItems(Params{
		"itemids": itemIDs,
	})
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't obtain items of triggers`,
		)
	}

	err = withSpinner(
		":: Requesting history for items of triggers",
		func() error {
			for _, item := range items {
				lastValues, err := zabbix.GetHistory(Params{
					"history": item.ValueType,
					"itemids": item.ID,
					"limit":   1,
				})
				if err != nil {
					return err
				}

				if len(lastValues) == 0 {
					continue
				}

				history[item.ID] = ItemHistory{
					Item:    item,
					History: lastValues[0],
				}
			}

			return nil
		},
	)

	return history, err
}

func parseParams(args map[string]interface{}) (Params, error) {
	var (
		severity    = args["--severity"].(int)
		onlyNotAck  = args["--only-nack"].(bool)
		maintenance = args["--maintenance"].(bool)
		problem     = args["--problem"].(bool)
		recent      = args["--recent"].(bool)
		since, _    = args["--since"].(string)
		until, _    = args["--until"].(string)
		sort        = strings.Split(args["--sort"].(string), ",")
		order       = args["--order"].(string)
		limit       = args["--limit"].(string)
	)

	params := Params{
		"sortfield":    sort,
		"sortorder":    order,
		"min_severity": severity,
		"limit":        limit,
	}

	if onlyNotAck {
		params["withLastEventUnacknowledged"] = "1"
	}

	if maintenance {
		params["maintenance"] = "1"
	}

	if recent {
		params["only_true"] = "1"
	}

	if problem {
		params["filter"] = Params{"value": "1"}
	}

	var err error
	if until != "" {
		params["lastChangeTill"], err = parseDateTime(until)
	} else if since != "" {
		params["lastChangeSince"], err = parseDateTime(since)
	}

	return params, err
}

func parseDateTime(value string) (int64, error) {
	date, err := timeutils.ParseDateString(value)
	if err != nil {
		return 0, hierr.Errorf(err, "can't parse datetime '%s'", value)
	}

	return date.Unix(), nil
}

func confirmAcknowledge() bool {
	var value string
	fmt.Fprintf(os.Stderr, "\n:: Proceed with acknowledge? [Y/n]: ")
	fmt.Scanln(&value)
	return value == "" || value == "Y" || value == "y"
}
