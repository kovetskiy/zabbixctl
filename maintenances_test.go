package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

	maintenances, err := zabbix.GetMaintenances(Params{})

	test.NoError(err)
	test.Len(maintenances, 1)

	test.Equal("3", maintenances[0].ID)
	test.Equal("Sunday maintenance", maintenances[0].Name)
	test.Equal("Zabbix servers", maintenances[0].Groups[0].Name)
}
