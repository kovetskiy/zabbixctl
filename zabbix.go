package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/reconquest/karma-go"
)

const (
	// 900 is default zabbix session ttl, -60 for safety
	ZabbixSessionTTL = 900 - 60
)

var (
	withAuthFlag    = true
	withoutAuthFlag = false
)

type Params map[string]interface{}

type Request struct {
	RPC    string      `json:"jsonrpc"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	Auth   string      `json:"auth,omitempty"`
	ID     int64       `json:"id"`
}

type Zabbix struct {
	basicURL   string
	apiURL     string
	session    string
	client     *http.Client
	requestID  int64
	apiVersion string
}

func NewZabbix(
	address, username, password, sessionFile string,
) (*Zabbix, error) {
	var err error

	zabbix := &Zabbix{
		client: &http.Client{},
	}

	if !strings.Contains(address, "://") {
		address = "http://" + address
	}

	zabbix.basicURL = strings.TrimSuffix(address, "/")
	zabbix.apiURL = zabbix.basicURL + "/api_jsonrpc.php"

	if sessionFile != "" {
		debugln("* reading session file")

		err = zabbix.restoreSession(sessionFile)
		if err != nil {
			return nil, karma.Format(
				err,
				"can't restore zabbix session using file '%s'",
				sessionFile,
			)
		}
	} else {
		debugln("* session feature is not used")
	}

	if zabbix.session == "" {
		err = zabbix.Login(username, password)
		if err != nil {
			return nil, karma.Format(
				err,
				"can't authorize user '%s' in zabbix server",
				username,
			)
		}
	} else {
		debugln("* using session instead of authorization")
	}

	if sessionFile != "" {
		debugln("* rewriting session file")

		// always rewrite session file, it will change modify date
		err = zabbix.saveSession(sessionFile)
		if err != nil {
			return nil, karma.Format(
				err,
				"can't save zabbix session to file '%s'",
				sessionFile,
			)
		}
	}

	if len(zabbix.apiVersion) < 1 {
		err = zabbix.GetAPIVersion()
		if err != nil {
			return nil, karma.Format(
				err,
				"can't get zabbix api version",
			)
		}
	}

	return zabbix, nil
}

func (zabbix *Zabbix) restoreSession(path string) error {
	file, err := os.OpenFile(
		path, os.O_CREATE|os.O_RDWR, 0600,
	)
	if err != nil {
		return karma.Format(
			err, "can't open session file",
		)
	}

	stat, err := file.Stat()
	if err != nil {
		return karma.Format(
			err, "can't stat session file",
		)
	}

	if time.Since(stat.ModTime()).Seconds() < ZabbixSessionTTL {
		session, err := ioutil.ReadAll(file)
		if err != nil {
			return karma.Format(
				err, "can't read session file",
			)
		}

		zabbix.session = string(session)
	} else {
		debugln("* session is outdated")
	}

	return nil
}

func (zabbix *Zabbix) saveSession(path string) error {
	err := ioutil.WriteFile(path, []byte(zabbix.session), 0600)
	if err != nil {
		return karma.Format(
			err,
			"can't write session file",
		)
	}

	return nil
}

func (zabbix *Zabbix) GetAPIVersion() error {
	var response ResponseAPIVersion

	debugln("* apiinfo.version")

	err := zabbix.call(
		"apiinfo.version",
		Params{},
		&response,
		withoutAuthFlag,
	)
	if err != nil {
		return err
	}

	zabbix.apiVersion = response.Version

	return nil
}

func (zabbix *Zabbix) Login(username, password string) error {
	var response ResponseLogin

	debugln("* authorizing")

	err := zabbix.call(
		"user.login",
		Params{"user": username, "password": password},
		&response,
		withAuthFlag,
	)
	if err != nil {
		return err
	}

	zabbix.session = response.Token

	return nil
}

func (zabbix *Zabbix) Acknowledge(identifiers []string) error {
	var response ResponseRaw

	debugln("* acknowledging triggers")

	params := Params{
		"eventids": identifiers,
		"message":  "ack",
	}

	if len(strings.Split(zabbix.apiVersion, ".")) < 1 {
		return karma.Format("can't parse zabbix version %s", zabbix.apiVersion)
	}

	majorZabbixVersion := strings.Split(zabbix.apiVersion, ".")[0]

	switch majorZabbixVersion {

	case "4":
		//https://www.zabbix.com/documentation/4.0/manual/api/reference/event/acknowledge
		params["action"] = 6

	case "3":
		//https://www.zabbix.com/documentation/3.4/manual/api/reference/event/acknowledge
		params["action"] = 1

		//default:
		//https://www.zabbix.com/documentation/1.8/api/event/acknowledge
		//https://www.zabbix.com/documentation/2.0/manual/appendix/api/event/acknowledge

	}

	err := zabbix.call(
		"event.acknowledge",
		params,
		&response,
		withAuthFlag,
	)
	if err != nil {
		return err
	}

	return nil
}

func (zabbix *Zabbix) GetTriggers(extend Params) ([]Trigger, error) {
	debugln("* retrieving triggers list")

	params := Params{
		"monitored":         true,
		"selectHosts":       []string{"name"},
		"selectGroups":      []string{"groupid", "name"},
		"selectLastEvent":   "extend",
		"selectFunctions":   "extend",
		"expandExpression":  true,
		"expandData":        true,
		"expandDescription": true,
		"skipDependent":     true,
		"preservekeys":      true,
	}

	for key, value := range extend {
		params[key] = value
	}

	var response ResponseTriggers
	err := zabbix.call("trigger.get", params, &response, withAuthFlag)
	if err != nil {
		return nil, err
	}

	var triggers []Trigger
	for _, trigger := range unshuffle(response.Data) {
		triggers = append(triggers, trigger.(Trigger))
	}

	return triggers, nil
}

func (zabbix *Zabbix) GetMaintenances(params Params) ([]Maintenance, error) {
	debugln("* retrieving maintenances list")

	var response ResponseMaintenances
	err := zabbix.call("maintenance.get", params, &response, withAuthFlag)
	if err != nil {
		return nil, err
	}

	var maintenances []Maintenance
	for _, maintenance := range response.Data {
		maintenances = append(maintenances, maintenance)
	}

	return maintenances, nil
}

func (zabbix *Zabbix) CreateMaintenance(params Params) (Maintenances, error) {
	debugln("* create maintenances list")

	var response ResponseMaintenancesArray

	err := zabbix.call("maintenance.create", params, &response, withAuthFlag)

	return response.Data, err
}

func (zabbix *Zabbix) UpdateMaintenance(params Params) (Maintenances, error) {
	debugln("* update maintenances list")

	var response ResponseMaintenancesArray

	err := zabbix.call("maintenance.update", params, &response, withAuthFlag)

	return response.Data, err
}

func (zabbix *Zabbix) RemoveMaintenance(params interface{}) (Maintenances, error) {
	debugln("* remove maintenances")

	var response ResponseMaintenancesArray

	err := zabbix.call("maintenance.delete", params, &response, withAuthFlag)

	return response.Data, err
}

func (zabbix *Zabbix) GetItems(params Params) ([]Item, error) {
	debugln("* retrieving items list")

	var response ResponseItems
	err := zabbix.call("item.get", params, &response, withAuthFlag)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (zabbix *Zabbix) GetHTTPTests(params Params) ([]HTTPTest, error) {
	debugln("* retrieving web scenarios list")

	var response ResponseHTTPTests
	err := zabbix.call("httptest.get", params, &response, withAuthFlag)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (zabbix *Zabbix) GetUsersGroups(params Params) ([]UserGroup, error) {
	debugln("* retrieving users groups list")

	var response ResponseUserGroup
	err := zabbix.call("usergroup.get", params, &response, withAuthFlag)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (zabbix *Zabbix) AddUserToGroups(
	groups []UserGroup,
	user User,
) error {
	for _, group := range groups {
		identifiers := []string{user.ID}

		for _, groupUser := range group.Users {
			identifiers = append(identifiers, groupUser.ID)
		}

		debugf("* adding user %s to group %s", user.Alias, group.Name)

		err := zabbix.call(
			"usergroup.update",
			Params{"usrgrpid": group.ID, "userids": identifiers},
			&ResponseRaw{},
			withAuthFlag,
		)
		if err != nil {
			return karma.Format(
				err,
				"can't update usergroup %s", group.Name,
			)
		}
	}

	return nil
}

func (zabbix *Zabbix) RemoveUserFromGroups(
	groups []UserGroup,
	user User,
) error {
	for _, group := range groups {
		identifiers := []string{}

		for _, groupUser := range group.Users {
			if groupUser.ID == user.ID {
				continue
			}

			identifiers = append(identifiers, groupUser.ID)
		}

		debugf("* removing user %s from group %s", user.Alias, group.Name)

		err := zabbix.call(
			"usergroup.update",
			Params{"usrgrpid": group.ID, "userids": identifiers},
			&ResponseRaw{},
			withAuthFlag,
		)
		if err != nil {
			return karma.Format(
				err,
				"can't update usergroup %s", group.Name,
			)
		}
	}

	return nil
}

func (zabbix *Zabbix) GetUsers(params Params) ([]User, error) {
	debugln("* retrieving users list")

	var response ResponseUsers
	err := zabbix.call("user.get", params, &response, withAuthFlag)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (zabbix *Zabbix) GetHosts(params Params) ([]Host, error) {
	debugf("* retrieving hosts list")

	var response ResponseHosts
	err := zabbix.call("host.get", params, &response, withAuthFlag)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (zabbix *Zabbix) RemoveHosts(params interface{}) (Hosts, error) {
	debugf("* remove hosts list")

	var response ResponseHostsArray
	err := zabbix.call("host.delete", params, &response, withAuthFlag)

	return response.Data, err
}

func (zabbix *Zabbix) GetGroups(params Params) ([]Group, error) {
	debugf("* retrieving groups list")

	var response ResponseGroups
	err := zabbix.call("hostgroup.get", params, &response, withAuthFlag)

	return response.Data, err
}

func (zabbix *Zabbix) GetGraphURL(identifier string) string {
	return zabbix.getGraphURL([]string{identifier}, "showgraph", "0")
}

func (zabbix *Zabbix) GetNormalGraphURL(identifiers []string) string {
	return zabbix.getGraphURL(identifiers, "batchgraph", "0")
}

func (zabbix *Zabbix) GetStackedGraphURL(identifiers []string) string {
	return zabbix.getGraphURL(identifiers, "batchgraph", "1")
}

func (zabbix *Zabbix) getGraphURL(
	identifiers []string,
	action string,
	graphType string,
) string {
	encodedIdentifiers := []string{}

	for _, identifier := range identifiers {
		encodedIdentifiers = append(
			encodedIdentifiers,
			"itemids%5B%5D="+identifier,
		)
	}

	return zabbix.basicURL + fmt.Sprintf(
		"/history.php?action=%s&graphtype=%s&%s",
		action,
		graphType,
		strings.Join(encodedIdentifiers, "&"),
	)
}

func (zabbix *Zabbix) GetHistory(extend Params) ([]History, error) {
	debugf("* retrieving items history")

	params := Params{
		"output":    "extend",
		"sortfield": "clock",
		"sortorder": "DESC",
	}

	for key, value := range extend {
		params[key] = value
	}

	var response ResponseHistory
	err := zabbix.call("history.get", params, &response, withAuthFlag)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (zabbix *Zabbix) call(
	method string, params interface{}, response Response, authFlag bool,
) error {
	debugf("~> %s", method)
	debugParams(params)

	request := Request{
		RPC:    "2.0",
		Method: method,
		Params: params,
		ID:     atomic.AddInt64(&zabbix.requestID, 1),
	}

	if authFlag {
		request.Auth = zabbix.session
	}

	buffer, err := json.Marshal(request)
	if err != nil {
		return karma.Format(
			err,
			"can't encode request to JSON",
		)
	}

	payload, err := http.NewRequest(
		"POST",
		zabbix.apiURL,
		bytes.NewReader(buffer),
	)
	if err != nil {
		return karma.Format(
			err,
			"can't create http request",
		)
	}

	payload.ContentLength = int64(len(buffer))
	payload.Header.Add("Content-Type", "application/json-rpc")
	payload.Header.Add("User-Agent", "zabbixctl")

	resource, err := zabbix.client.Do(payload)
	if err != nil {
		return karma.Format(
			err,
			"http request to zabbix api failed",
		)
	}

	body, err := ioutil.ReadAll(resource.Body)
	if err != nil {
		return karma.Format(
			err,
			"can't read zabbix api response body",
		)
	}

	debugf("<~ %s", resource.Status)

	if traceMode {
		var tracing bytes.Buffer
		err = json.Indent(&tracing, body, "", "  ")
		if err != nil {
			return karma.Format(err, "can't indent api response body")
		}
		tracef("<~ %s", tracing.String())
	}

	err = json.Unmarshal(body, response)
	if err != nil {
		// There is can be bullshit case when zabbix sends empty `result`
		// array and json.Unmarshal triggers the error with message about
		// failed type conversion to map[].
		//
		// So, we must check that err is not this case.
		var raw ResponseRaw
		rawErr := json.Unmarshal(body, &raw)
		if rawErr != nil {
			// return original error
			return err
		}

		if result, ok := raw.Result.([]interface{}); ok && len(result) == 0 {
			return nil
		}

		return err
	}

	err = response.Error()
	if err != nil {
		return karma.Format(
			err,
			"zabbix returned error while working with api method %s",
			method,
		)
	}

	return nil
}

func debugParams(params interface{}, prefix ...string) {

	switch params.(type) {
	case Params:
		p, _ := params.(Params)
		for key, value := range p {
			if valueParams, ok := value.(Params); ok {
				debugParams(valueParams, append(prefix, key)...)
				continue
			}

			if key == "password" {
				value = "**********"
			}

			debugf(
				"** %s%s: %v",
				strings.Join(append(prefix, ""), "."),
				key, value,
			)
		}
	case interface{}:
		if p, ok := params.([]string); ok {
			for _, value := range p {
				debugf("** %v", value)
			}
		}
	}
}

func unshuffle(target interface{}) []interface{} {
	tears := reflect.ValueOf(target)

	var values []interface{}
	for _, key := range tears.MapKeys() {
		values = append(
			values,
			tears.MapIndex(key).Interface(),
		)
	}

	return values
}
