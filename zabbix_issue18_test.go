package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	issue18_data = `
{
  "jsonrpc": "2.0",
  "result": [
	{
	  "itemid": "28494",
	  "type": "2",
	  "snmp_community": "",
	  "snmp_oid": "",
	  "hostid": "10084",
	  "name": "Number of csvs today",
	  "key_": "csv.today.count",
	  "delay": "0",
	  "history": "90d",
	  "trends": "365d",
	  "status": "0",
	  "value_type": "3",
	  "trapper_hosts": "",
	  "units": "",
	  "snmpv3_securityname": "",
	  "snmpv3_securitylevel": "0",
	  "snmpv3_authpassphrase": "",
	  "snmpv3_privpassphrase": "",
	  "formula": "",
	  "error": "",
	  "lastlogsize": "0",
	  "logtimefmt": "",
	  "templateid": "28381",
	  "valuemapid": "0",
	  "params": "",
	  "ipmi_sensor": "",
	  "authtype": "0",
	  "username": "",
	  "password": "",
	  "publickey": "",
	  "privatekey": "",
	  "mtime": "0",
	  "flags": "0",
	  "interfaceid": "0",
	  "port": "",
	  "description": "",
	  "inventory_link": "0",
	  "lifetime": "30d",
	  "snmpv3_authprotocol": "0",
	  "snmpv3_privprotocol": "0",
	  "state": "0",
	  "snmpv3_contextname": "",
	  "evaltype": "0",
	  "jmx_endpoint": "",
	  "master_itemid": "0",
	  "timeout": "3s",
	  "url": "",
	  "query_fields": [],
	  "posts": "",
	  "status_codes": "200",
	  "follow_redirects": "1",
	  "post_type": "0",
	  "http_proxy": "",
	  "headers": [],
	  "retrieve_mode": "0",
	  "request_method": "1",
	  "output_format": "0",
	  "ssl_cert_file": "",
	  "ssl_key_file": "",
	  "ssl_key_password": "",
	  "verify_peer": "0",
	  "verify_host": "0",
	  "allow_traps": "0",
	  "lastclock": 1548924066,
	  "lastns": 388228365,
	  "lastvalue": "6",
	  "prevvalue": "6"
	},
	{
		  "itemid": "28461",
		  "type": "0",
		  "snmp_community": "",
		  "snmp_oid": "",
		  "hostid": "10084",
		  "name": "Indices count",
		  "key_": "elastizabbix[cluster,indices.count]",
		  "delay": "60",
		  "history": "7d",
		  "trends": "365d",
		  "status": "0",
		  "value_type": "3",
		  "trapper_hosts": "",
		  "units": "",
		  "snmpv3_securityname": "",
		  "snmpv3_securitylevel": "0",
		  "snmpv3_authpassphrase": "",
		  "snmpv3_privpassphrase": "",
		  "formula": "",
		  "error": "Unsupported item key.",
		  "lastlogsize": "0",
		  "logtimefmt": "",
		  "templateid": "28351",
		  "valuemapid": "0",
		  "params": "",
		  "ipmi_sensor": "",
		  "authtype": "0",
		  "username": "",
		  "password": "",
		  "publickey": "",
		  "privatekey": "",
		  "mtime": "0",
		  "flags": "0",
		  "interfaceid": "1",
		  "port": "",
		  "description": "",
		  "inventory_link": "0",
		  "lifetime": "30d",
		  "snmpv3_authprotocol": "0",
		  "snmpv3_privprotocol": "0",
		  "state": "1",
		  "snmpv3_contextname": "",
		  "evaltype": "0",
		  "jmx_endpoint": "",
		  "master_itemid": "0",
		  "timeout": "3s",
		  "url": "",
		  "query_fields": [],
		  "posts": "",
		  "status_codes": "200",
		  "follow_redirects": "1",
		  "post_type": "0",
		  "http_proxy": "",
		  "headers": [],
		  "retrieve_mode": "0",
		  "request_method": "1",
		  "output_format": "0",
		  "ssl_cert_file": "",
		  "ssl_key_file": "",
		  "ssl_key_password": "",
		  "verify_peer": "0",
		  "verify_host": "0",
		  "allow_traps": "0",
		  "lastclock": "0",
		  "lastns": "0",
		  "lastvalue": "0",
		  "prevvalue": "0"
		}
	]
}
`
)

func TestIssue18(t *testing.T) {
	test := assert.New(t)

	testserver := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, issue18_data)
		},
	))
	defer testserver.Close()

	zabbix := &Zabbix{}
	zabbix.client = testserver.Client()
	zabbix.apiURL = testserver.URL

	items, err := zabbix.GetItems(Params{"hostids": []string{"10084"}})
	test.NoError(err)
	test.Len(items, 2)

	test.Equal("1548924066", items[0].getLastClock())
	test.Equal("0", items[1].getLastClock())

	test.Equal("2019-01-31 11:41:06", items[0].DateTime())
	test.Equal("-", items[1].DateTime())
}
