package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/reconquest/karma-go"
)

func handleLatestData(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {
	var (
		hostnames, pattern = parseSearchQuery(args["<pattern>"].([]string))
		graphs             = args["--graph"].(bool)
		stackedGraph       = args["--stacked"].(bool)
		normalGraph        = args["--normal"].(bool)
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
		return karma.Format(
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

	var (
		items     []Item
		webchecks []HTTPTest
	)

	err = withSpinner(
		":: Requesting information about hosts items & web scenarios",
		func() error {
			errs := make(chan error)

			go func() {
				var err error

				items, err = zabbix.GetItems(Params{
					"hostids":  identifiers,
					"webitems": "1",
				})

				errs <- err
			}()

			go func() {
				var err error

				webchecks, err = zabbix.GetHTTPTests(Params{
					"hostids":     identifiers,
					"expandName":  "1",
					"selectSteps": "extend",
				})

				errs <- err
			}()

			for _, err := range []error{<-errs, <-errs} {
				if err != nil {
					return err
				}
			}

			return nil
		},
	)

	if err != nil {
		return karma.Format(
			err,
			"can't obtain zabbix items",
		)
	}

	var matchedItemIDs = []string{}

	for _, item := range items {
		line := fmt.Sprintf(
			"%s\t%s\t%s\t%s\t%-10s",
			hash[item.HostID].Name, item.Type.String(), item.Format(),
			item.DateTime(), item.LastValue,
		)

		if pattern != "" && !matchPattern(pattern, line) {
			continue
		}

		fmt.Fprint(table, line)

		if graphs {
			fmt.Fprintf(table, "\t%s", zabbix.GetGraphURL(item.ID))
		}

		fmt.Fprint(table, "\n")

		matchedItemIDs = append(matchedItemIDs, item.ID)
	}

	for _, check := range webchecks {
		line := fmt.Sprintf(
			"%s\t%s\t%s",
			hash[check.HostID].Name, `scenario`, check.Format(),
		)

		if pattern != "" && !matchPattern(pattern, line) {
			continue
		}

		fmt.Fprintln(table, line)
	}

	switch {
	case stackedGraph:
		fmt.Println(zabbix.GetStackedGraphURL(matchedItemIDs))

	case normalGraph:
		fmt.Println(zabbix.GetNormalGraphURL(matchedItemIDs))

	default:
		table.Flush()
	}

	return nil
}
