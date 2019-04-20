package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	karma "github.com/reconquest/karma-go"
	"github.com/simplereach/timeutils"
)

func handleMaintenances(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {

	var (
		err                  error
		addMaintenance, _    = args["--add"].(string)
		removeMaintenance, _ = args["--remove"].(string)

		maintenances []Maintenance
	)

	destiny := karma.Describe("method", "handleMaintenances")

	switch {

	case addMaintenance != "":

		maintenances, err = searchMaintenances(
			zabbix,
			Params{
				"search": Params{
					"name": addMaintenance,
				},
				"selectGroups": "extend",
				"selectHosts":  "extend",
			})
		if err != nil {
			return destiny.Describe(
				"error", err,
			).Reason(
				"can't obtain zabbix maintenances",
			)
		}

		switch len(maintenances) {

		case 0:
			err = handleAddMaintenance(zabbix, config, args)
		case 1:
			err = handleUpdateMaintenance(zabbix, config, args, maintenances)
		}

	case removeMaintenance != "":

		maintenances, err = searchMaintenances(
			zabbix,
			Params{
				"search": Params{
					"name": removeMaintenance,
				},
				"selectGroups": "extend",
				"selectHosts":  "extend",
			})
		if err != nil {
			return destiny.Describe(
				"error", err,
			).Reason(
				"can't obtain zabbix maintenances",
			)
		}
		err = handleRemoveMaintenance(zabbix, config, args, maintenances)

	default:
		err = handleListMaintenances(zabbix, config, args)
	}

	if err != nil {
		return destiny.Describe(
			"error", err,
		).Reason(
			"can't operate with maintenance",
		)
	}

	return nil
}

func handleAddMaintenance(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {

	var (
		hostnames, _      = parseSearchQuery(args["<pattern>"].([]string))
		addMaintenance, _ = args["--add"].(string)
		confirmation      = !args["--noconfirm"].(bool)
		fromStdin         = args["--read-stdin"].(bool)

		hostids   = []string{}
		uniqHosts = make(map[string]bool)
		err       error
		hosts     []Host
		params    Params
	)

	destiny := karma.Describe(
		"method", "AddMaintenance",
	).Describe(
		"name", addMaintenance,
	)

	if fromStdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			hostnames = append(hostnames, scanner.Text())
		}
	}

	for _, hostname := range hostnames {
		foundHosts, err := searchHosts(zabbix, hostname)
		if err != nil {
			return destiny.Describe(
				"error", err,
			).Describe(
				"hostname", hostname,
			).Reason(
				"can't obtain zabbix hosts",
			)
		}

		for _, host := range foundHosts {
			if _, value := uniqHosts[host.ID]; !value {
				uniqHosts[host.ID] = true
				hostids = append(hostids, host.ID)
				hosts = append(hosts, host)
			}
		}
	}

	printHostsTable(hosts)

	if confirmation {
		if !confirmMaintenance("create", addMaintenance) {
			return nil
		}
	}

	timeperiod, activeTill, err := createTimeperiod(args)
	if err != nil {
		return destiny.Describe(
			"error", err,
		).Reason(
			"can't create timeperiod for maintenance",
		)
	}

	params = Params{
		"name":         addMaintenance,
		"active_since": timeperiod.StartDate,
		"active_till":  activeTill,
		"hostids":      hostids,
		"timeperiods":  []Timeperiod{timeperiod},
	}
	err = withSpinner(
		":: Requesting for create to specified maintenance",
		func() error {
			_, err = zabbix.CreateMaintenance(params)
			return err
		},
	)
	if err != nil {
		return destiny.Describe(
			"error", err,
		).Reason(
			"can't create zabbix maintenance",
		)
	}
	if err != nil {
		return destiny.Describe(
			"error", err,
		).Reason(
			"can't create zabbix maintenance",
		)
	}
	return nil
}

func handleUpdateMaintenance(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
	maintenances []Maintenance,
) error {

	var (
		hostnames, _      = parseSearchQuery(args["<pattern>"].([]string))
		addMaintenance, _ = args["--add"].(string)
		confirmation      = !args["--noconfirm"].(bool)
		fromStdin, _      = args["--read-stdin"].(bool)

		hostids   = []string{}
		uniqHosts = make(map[string]bool)
		err       error
		hosts     []Host
		params    Params
	)

	destiny := karma.Describe(
		"method", "UpdateMaintenance",
	).Describe(
		"name", addMaintenance,
	)

	if fromStdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			hostnames = append(hostnames, scanner.Text())
		}
	}

	for _, hostname := range hostnames {
		foundHosts, err := searchHosts(zabbix, hostname)
		if err != nil {
			return destiny.Describe(
				"error", err,
			).Describe(
				"hostname", hostname,
			).Reason(
				"can't obtain zabbix hosts",
			)
		}

		for _, host := range foundHosts {
			if _, value := uniqHosts[host.ID]; !value {
				uniqHosts[host.ID] = true
				hostids = append(hostids, host.ID)
				hosts = append(hosts, host)
			}
		}
	}

	maintenance := maintenances[0]
	for _, host := range maintenance.Hosts {
		if _, value := uniqHosts[host.ID]; !value {
			uniqHosts[host.ID] = true
			hostids = append(hostids, host.ID)
			hosts = append(hosts, host)
		}
	}

	printHostsTable(hosts)

	if confirmation {
		if fromStdin {
			fmt.Println("Use flag -z with -f only.")
			return nil
		}

		if !confirmMaintenance("updating", addMaintenance) {
			return nil
		}
	}

	params = Params{
		"maintenanceid": maintenance.ID,
		"hostids":       hostids,
	}
	err = withSpinner(
		":: Requesting for updating hosts to specified maintenance",
		func() error {
			_, err = zabbix.UpdateMaintenance(params)
			return err
		},
	)
	if err != nil {
		return destiny.Describe(
			"error", err,
		).Reason(
			"can't create zabbix maintenance",
		)
	}
	return nil
}

func handleRemoveMaintenance(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
	maintenances []Maintenance,
) error {

	var (
		removeMaintenance, _ = args["--remove"].(string)
		confirmation         = !args["--noconfirm"].(bool)

		pattern string
		err     error
		extend  = true
	)

	destiny := karma.Describe(
		"method", "RemoveMaintenances",
	).Describe(
		"name", removeMaintenance,
	)

	if len(maintenances) != 1 {
		return destiny.Reason(
			"can't remove more then one maintenance",
		)
	}

	printMaintenancesTable(maintenances, pattern, extend)

	if confirmation {
		if !confirmMaintenance("removing", removeMaintenance) {
			return nil
		}
	}

	err = withSpinner(
		":: Requesting for removing hosts to specified maintenance",
		func() error {
			maintenance := maintenances[0].ID
			payload := []string{maintenance}
			_, err = zabbix.RemoveMaintenance(payload)
			return err
		},
	)
	if err != nil {
		return destiny.Describe(
			"error", err,
		).Reason(
			"can't remove zabbix maintenances",
		)
	}
	return nil
}

func handleListMaintenances(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {

	var (
		hostnames, pattern = parseSearchQuery(args["<pattern>"].([]string))

		hostids      = []string{}
		groupids     = []string{}
		extend       bool
		err          error
		hosts        []Host
		groups       []Group
		maintenances []Maintenance
	)

	destiny := karma.Describe("method", "ListMaintenances")

	params := Params{}

	for _, hostname := range hostnames {
		hosts, err = searchHosts(zabbix, hostname)
		if err != nil {
			return destiny.Describe(
				"error", err,
			).Describe(
				"hostname", hostname,
			).Reason(
				"can't obtain zabbix hosts",
			)
		}

		for _, host := range hosts {
			hostids = append(hostids, host.ID)
		}
	}

	if len(hostids) > 0 {

		err = withSpinner(
			":: Requesting information about groups",
			func() error {
				groups, err = zabbix.GetGroups(Params{
					"output":       "extend",
					"selectGroups": "extend",
					"hostids":      hostids,
				})
				return err
			},
		)
		if err != nil {
			return destiny.Describe(
				"error", err,
			).Reason(
				"can't obtain zabbix groups",
			)
		}

		for _, group := range groups {
			groupids = append(groupids, group.ID)
		}

		params["hostids"] = hostids
		params["groupids"] = groupids
	}

	if len(hostnames) > 0 || pattern != "" {
		extend = true
		params["selectGroups"] = "extend"
		params["selectHosts"] = "extend"
	}

	maintenances, err = searchMaintenances(zabbix, params)
	if err != nil {
		return destiny.Describe(
			"error", err,
		).Reason(
			"can't obtain zabbix maintenances",
		)
	}

	printMaintenancesTable(maintenances, pattern, extend)

	return nil
}

func printMaintenancesTable(
	maintenances []Maintenance,
	pattern string,
	extend bool,
) error {

	var lines = [][]string{}

	for _, maintenance := range maintenances {
		if pattern != "" && !matchPattern(pattern, maintenance.GetString()) {
			continue
		}

		// calculate max row number
		size := []int{len(maintenance.Timeperiods)}

		if extend {
			size = append(size, len(maintenance.Hosts))
			size = append(size, len(maintenance.Groups))

			sort.Slice(
				size,
				func(i, j int) bool { return size[i] > size[j] },
			)
		}

		for i := 0; i < size[0]; i++ {

			var (
				timeperiodType, timeperiodStarDate, hostName, groupName string
				maintenanceID, maintenanceName, maintenanceTypeCollect  string
				maintenanceSince, maintenanceTill, maintenanceStatus    string
				timeperiodPeriod                                        string
			)

			if i == 0 {
				maintenanceID = maintenance.ID
				maintenanceName = maintenance.Name
				maintenanceSince = maintenance.GetDateTime(maintenance.Since)
				maintenanceTill = maintenance.GetDateTime(maintenance.Till)
				maintenanceStatus = maintenance.GetStatus()
				maintenanceTypeCollect = maintenance.GetTypeCollect()
			}

			if len(maintenance.Timeperiods) > i {
				timeperiodType = maintenance.Timeperiods[i].GetType()
				timeperiodStarDate = maintenance.Timeperiods[i].GetStartDate()
				timeperiodPeriod = strconv.FormatInt(
					int64(maintenance.Timeperiods[i].GetPeriodMinute()), 10,
				)
			}
			if len(maintenance.Groups) > i {
				groupName = maintenance.Groups[i].Name
			}
			if len(maintenance.Hosts) > i {
				hostName = maintenance.Hosts[i].Name
			}

			line := []string{
				maintenanceID,
				maintenanceName,
				maintenanceSince,
				maintenanceTill,
				maintenanceStatus,
				maintenanceTypeCollect,
				timeperiodType,
				timeperiodStarDate,
				timeperiodPeriod,
			}

			if extend {
				line = append(line, groupName, hostName)
			}
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{
		"ID",
		"Name",
		"Since",
		"Till",
		"Status",
		"Data",
		"Type",
		"Start",
		"Minute",
	}
	if extend {
		header = append(header, "Group", "Host")
	}
	table.SetHeader(header)
	table.AppendBulk(lines)
	table.Render()

	return nil
}

func searchMaintenances(zabbix *Zabbix, extend Params) ([]Maintenance, error) {

	var (
		maintenances []Maintenance
		err          error
	)

	params := Params{
		"output":                 "extend",
		"selectTimeperiods":      "extend",
		"searchWildcardsEnabled": "1",
	}

	for key, value := range extend {
		params[key] = value
	}

	err = withSpinner(
		":: Requesting information about maintenances",
		func() error {
			maintenances, err = zabbix.GetMaintenances(params)
			return err
		},
	)
	return maintenances, err
}

func createTimeperiod(
	args map[string]interface{},
) (Timeperiod, string, error) {

	var (
		startDate, _ = args["--start"].(string)
		endDate, _   = args["--end"].(string)
		period, _    = args["--period"].(string)

		timeperiod Timeperiod
		activeTill int64
	)

	periodSeconds, err := parsePeriod(period)
	if err != nil {
		return timeperiod, "", err
	}

	activeSince, err := parseDate(startDate)
	if err != nil {
		return timeperiod, "", err
	}

	switch {
	case endDate == "":
		activeTill = activeSince + periodSeconds

	case endDate != "":
		activeTill, err = parseDate(endDate)
		if err != nil {
			return timeperiod, "", err
		}

		if activeTill < activeSince+periodSeconds {
			activeTill = activeSince + periodSeconds
		}
	}

	timeperiod.TypeID = "0"
	timeperiod.Every = "1"
	timeperiod.Month = "0"
	timeperiod.DayOfWeek = "0"
	timeperiod.Day = "1"
	timeperiod.StartTime = "0"
	timeperiod.StartDate = strconv.FormatInt(int64(activeSince), 10)
	timeperiod.Period = strconv.FormatInt(int64(periodSeconds), 10)

	return timeperiod, strconv.FormatInt(int64(activeTill), 10), nil
}

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

func parsePeriod(targets string) (int64, error) {

	var (
		err                  error
		days, hours, minutes int64
	)

	destiny := karma.Describe(
		"method", "parsePeriod",
	).Describe(
		"period", targets,
	)

	switch {
	case strings.HasSuffix(targets, "d"):
		days, err = strconv.ParseInt(
			strings.TrimSuffix(targets, "d"), 10, 64,
		)
		return days * 86400, nil
	case strings.HasSuffix(targets, "h"):
		hours, err = strconv.ParseInt(
			strings.TrimSuffix(targets, "h"), 10, 64,
		)
		return hours * 3600, nil
	case strings.HasSuffix(targets, "m"):
		minutes, err = strconv.ParseInt(
			strings.TrimSuffix(targets, "m"), 10, 64,
		)
		return minutes * 60, nil
	}
	return 0, destiny.Describe(
		"error", err,
	).Describe(
		"period", targets,
	).Reason("can't parse")
}

func confirmMaintenance(messages, maintenance string) bool {

	var value string
	fmt.Fprintf(
		os.Stderr,
		"\n:: Proceed with %s to maintenance %s? [Y/n]:",
		messages,
		maintenance,
	)

	fmt.Scanln(&value)
	return value == "" || value == "Y" || value == "y"
}
