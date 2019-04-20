package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	// https://www.zabbix.com/documentation/3.4/manual/api/reference/maintenance/get
	maintenance_get = `
{
    "jsonrpc": "2.0",
    "result": [
        {
            "maintenanceid": "3",
            "name": "Sunday maintenance",
            "maintenance_type": "0",
            "description": "",
            "active_since": "1358844540",
            "active_till": "1390466940",
            "groups": [
                {
                    "groupid": "4",
                    "name": "Zabbix servers",
                    "internal": "0"
                }
            ],
            "timeperiods": [
                {
                    "timeperiodid": "4",
                    "timeperiod_type": "3",
                    "every": "1",
                    "month": "0",
                    "dayofweek": "1",
                    "day": "0",
                    "start_time": "64800",
                    "period": "3600",
                    "start_date": "2147483647"
                }
            ]
        }
    ],
    "id": 1
}`

	// https://www.zabbix.com/documentation/3.4/manual/api/reference/maintenance/delete
	maintenances_remove = `
{
    "jsonrpc": "2.0",
    "result": {
        "maintenanceids": [
            "3",
            "1"
        ]
    },
    "id": 1
}
`

	// https://www.zabbix.com/documentation/3.4/manual/api/reference/maintenance/create
	maintenance_create = `
{
    "jsonrpc": "2.0",
    "result": {
        "maintenanceids": [
            "3"
        ]
    },
    "id": 1
}
`
)

func TestMaintenanceGet(t *testing.T) {
	test := assert.New(t)

	testserver := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, maintenance_get)
		},
	))
	defer testserver.Close()

	zabbix := &Zabbix{}
	zabbix.client = testserver.Client()
	zabbix.apiURL = testserver.URL

	maintenances, err := zabbix.GetMaintenances(Params{
		"search": Params{
			"name": "Sunday maintenance",
		},
	})

	test.NoError(err)
	test.Len(maintenances, 1)

	test.Equal("3", maintenances[0].ID)
	test.Equal("Sunday maintenance", maintenances[0].Name)
	test.Equal("Zabbix servers", maintenances[0].Groups[0].Name)
}

func TestMaintenanceRemove(t *testing.T) {
	test := assert.New(t)

	testserver := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, maintenances_remove)
		},
	))
	defer testserver.Close()

	zabbix := &Zabbix{}
	zabbix.client = testserver.Client()
	zabbix.apiURL = testserver.URL

	payload := []string{"3", "1"}

	var maintenances Maintenances
	maintenances, err := zabbix.RemoveMaintenance(payload)

	test.NoError(err)
	test.Len(maintenances.ID, 2)

	test.Equal("3", maintenances.ID[0])
	test.Equal("1", maintenances.ID[1])
}

func TestMaintenanceCreate(t *testing.T) {
	test := assert.New(t)

	testserver := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, maintenance_create)
		},
	))
	defer testserver.Close()

	zabbix := &Zabbix{}
	zabbix.client = testserver.Client()
	zabbix.apiURL = testserver.URL

	var timeperiod Timeperiod

	timeperiod.TypeID = "0"
	timeperiod.Every = "1"
	timeperiod.Month = "0"
	timeperiod.DayOfWeek = "0"
	timeperiod.Day = "1"
	timeperiod.StartTime = "0"
	timeperiod.StartDate = strconv.FormatInt(int64(1551092132), 10)
	timeperiod.Period = strconv.FormatInt(int64(3600), 10)

	params := Params{
		"name":         "test maintenance",
		"active_since": "1551092132",
		"active_till":  "1551178532",
		"hostids":      []string{"2"},
		"timeperiods":  []Timeperiod{timeperiod},
	}

	var maintenances Maintenances
	maintenances, err := zabbix.RemoveMaintenance(params)

	test.NoError(err)
	test.Len(maintenances.ID, 1)

	test.Equal("3", maintenances.ID[0])
}
