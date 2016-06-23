package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"

	"github.com/zazab/hierr"
)

type Params map[string]interface{}

type Request struct {
	RPC    string `json:"jsonrpc"`
	Method string `json:"method"`
	Params Params `json:"params"`
	Auth   string `json:"auth,omitempty"`
	ID     int64  `json:"id"`
}

type Zabbix struct {
	address   string
	token     string
	client    *http.Client
	requestID int64
}

func NewZabbix(address, username, password string) (*Zabbix, error) {
	zabbix := &Zabbix{
		client: &http.Client{},
	}

	if !strings.Contains(address, "://") {
		address = "http://" + address
	}

	zabbix.address = strings.TrimSuffix(address, "/") + "/api_jsonrpc.php"

	return zabbix, zabbix.Login(username, password)
}

func (zabbix *Zabbix) Login(username, password string) error {
	var response ResponseLogin

	debugln("* authorizing")

	err := zabbix.call(
		"user.login",
		Params{"user": username, "password": password},
		&response,
	)
	if err != nil {
		return err
	}

	zabbix.token = response.Token

	return nil
}

func (zabbix *Zabbix) Acknowledge(identifiers []string) error {
	var response ResponseRaw

	debugln("* acknowledging triggers")

	err := zabbix.call(
		"event.acknowledge",
		Params{"eventids": identifiers},
		&response,
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
		"expandExpression":  true,
		"expandData":        true,
		"expandDescription": true,
		"skipDependent":     true,
		"preservekeys":      true,
	}

	for key, value := range extend {
		params[key] = value
	}

	var response ResponseTriggersList
	err := zabbix.call("trigger.get", params, &response)
	if err != nil {
		return nil, err
	}

	var triggers []Trigger
	for _, trigger := range unshuffle(response.Data) {
		triggers = append(triggers, trigger.(Trigger))
	}

	return triggers, nil
}

func (zabbix *Zabbix) call(
	method string, params Params, response Response,
) error {
	debugf("~> %s", method)
	debugParams(params)

	request := Request{
		RPC:    "2.0",
		Method: method,
		Params: params,
		Auth:   zabbix.token,
		ID:     atomic.AddInt64(&zabbix.requestID, 1),
	}

	buffer, err := json.Marshal(request)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't encode request to JSON",
		)
	}

	payload, err := http.NewRequest(
		"POST",
		zabbix.address,
		bytes.NewReader(buffer),
	)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't create http request",
		)
	}

	payload.ContentLength = int64(len(buffer))
	payload.Header.Add("Content-Type", "application/json-rpc")
	payload.Header.Add("User-Agent", "zabbixctl")

	resource, err := zabbix.client.Do(payload)
	if err != nil {
		return hierr.Errorf(
			err,
			"http request to zabbix api failed",
		)
	}

	body, err := ioutil.ReadAll(resource.Body)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't read zabbix api response body",
		)
	}

	tracef("<~ %s", string(body))
	debugf("<~ %s", resource.Status)

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
		return hierr.Errorf(
			err,
			"zabbix returned error while working with api method %s",
			method,
		)
	}

	return nil
}

func debugParams(params Params, prefix ...string) {
	for key, value := range params {
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
