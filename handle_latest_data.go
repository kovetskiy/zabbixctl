package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/zazab/hierr"
)

func handleLatestData(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {
	var (
		hostnames, pattern = parseSearchQuery(args["<pattern>"].([]string))
		graphs             = args["--graph"].(bool)
		table              = tabwriter.NewWriter(os.Stdout, 1, 4, 2, ' ', 0)
	)

	if len(hostnames) == 0 {
		return errors.New("no hostname specified")
	}

	var hosts []Host
	var err error

	err = withSpinner(
		":: Requesting information about hosts",
		func() error {
			hosts, err = zabbix.GetHosts(Params{
				"monitored_hosts":         "1",
				"with_items":              "1",
				"with_monitored_items":    "1",
				"with_monitored_triggers": "1",
				"search": Params{
					"name": hostnames,
				},
				"searchWildcardsEnabled": "1",
				"output": []string{
					"host",
				},
			})
			return err
		},
	)

	if err != nil {
		return hierr.Errorf(
			err,
			"can't obtain zabbix hosts",
		)
	}

	var (
		identifiers = []string{}
		hash        = map[string]Host{}
	)

	for _, host := range hosts {
		identifiers = append(identifiers, host.ID)
		hash[host.ID] = host
	}

	debugf("* hosts identifiers: %s", identifiers)

	params := Params{
		"hostids": identifiers,
	}

	var items []Item
	err = withSpinner(
		":: Requesting information about hosts items",
		func() error {
			items, err = zabbix.GetItems(params)
			return err
		},
	)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't obtain zabbix items",
		)
	}

	for _, item := range items {
		line := fmt.Sprintf(
			"%s\t%s\t%s\t%-10s",
			hash[item.HostID].Name, item.Format(),
			item.DateTime(), item.LastValue,
		)

		if pattern != "" && !matchPattern(pattern, line) {
			continue
		}

		if graphs {
			line = line + " " + zabbix.GetGraphURL(item.ID)
		}

		fmt.Fprintln(table, line)
	}

	table.Flush()

	return nil
}
