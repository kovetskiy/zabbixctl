package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	karma "github.com/reconquest/karma-go"
)

func handleHosts(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {
	var (
		hostnames, _  = parseSearchQuery(args["<pattern>"].([]string))
		removeHost, _ = args["--remove"].(string)

		err               error
		hostsTable, hosts []Host
	)

	destiny := karma.Describe("method", "handleHosts")

	switch {
	case removeHost != "":

		err = handleRemoveHosts(zabbix, config, args)
		if err != nil {
			return destiny.Describe(
				"error", err,
			).Describe(
				"hostname", removeHost,
			).Reason(
				"can't remove zabbix hosts",
			)
		}

	default:

		for _, hostname := range hostnames {
			hosts, err = searchHosts(zabbix, hostname)
			if err != nil {
				return destiny.Describe(
					"error", err,
				).Describe(
					"hostname", hostname,
				).Reason(
					"can't search zabbix hosts",
				)
			}
			hostsTable = append(hostsTable, hosts...)

		}
		if len(hostsTable) > 0 {
			printHostsTable(hostsTable)
		}

	}
	return nil
}

func handleRemoveHosts(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {
	var (
		removeHost, _ = args["--remove"].(string)
		confirmation  = !args["--noconfirm"].(bool)

		err   error
		hosts []Host
	)

	destiny := karma.Describe(
		"method", "removeHost",
	)

	hosts, err = searchHosts(zabbix, removeHost)
	if err != nil {
		return destiny.Describe(
			"error", err,
		).Reason(
			"can't obtain zabbix hosts",
		)
	}

	if len(hosts) == 0 {
		return nil
	}
	if len(hosts) > 1 {
		return destiny.Reason(
			"found more then one uniq host",
		)
	}

	printHostsTable(hosts)

	if confirmation {
		if !confirmHost("removing", removeHost) {
			return nil
		}
	}

	err = withSpinner(
		":: Requesting for removing host",
		func() error {
			payload := []string{hosts[0].ID}
			_, err = zabbix.RemoveHosts(payload)
			return err
		},
	)
	return err
}

func searchHosts(zabbix *Zabbix, hostname string) ([]Host, error) {

	var (
		hosts []Host
		err   error
	)

	if len(hostname) == 0 {
		return hosts, nil
	}

	params := Params{
		"search": Params{
			"name": hostname,
		},
		"output": []string{
			"host",
		},
		"searchWildcardsEnabled": "1",
	}

	err = withSpinner(
		":: Requesting information about hosts",
		func() error {
			hosts, err = zabbix.GetHosts(params)
			return err
		},
	)

	return hosts, err
}

func printHostsTable(
	hosts []Host,
) error {

	var lines = [][]string{}

	for _, host := range hosts {
		line := []string{
			host.ID,
			host.Name,
		}
		lines = append(lines, line)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name"})
	table.AppendBulk(lines)
	table.Render()

	return nil
}

func confirmHost(messages, host string) bool {

	var value string
	fmt.Fprintf(
		os.Stderr,
		"\n:: Proceed with %s host %s? [Y/n]:",
		messages,
		host,
	)

	fmt.Scanln(&value)
	return value == "" || value == "Y" || value == "y"
}
