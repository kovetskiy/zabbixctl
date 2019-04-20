package main

import "github.com/reconquest/karma-go"

type Response interface {
	Error() error
}

type ResponseRaw struct {
	Err struct {
		Data    string `json:"data"`
		Message string `json:"message"`
	} `json:"error"`

	Result interface{} `json:"result"`
}

func (response *ResponseRaw) Error() error {
	if response.Err.Data != "" && response.Err.Message != "" {
		return karma.Push(
			response.Err.Message,
			response.Err.Data,
		)
	}

	return nil
}

type ResponseLogin struct {
	ResponseRaw
	Token string `json:"result"`
}

type ResponseApiVersion struct {
	ResponseRaw
	Version string `json:"result"`
}

type ResponseTriggers struct {
	ResponseRaw
	Data map[string]Trigger `json:"result"`
}

type ResponseMaintenances struct {
	ResponseRaw
	Data []Maintenance `json:"result"`
}

// Response Create/Delete maintenace
type ResponseMaintenancesArray struct {
	ResponseRaw
	Data Maintenances `json:"result"`
}

type ResponseItems struct {
	ResponseRaw
	Data []Item `json:"result"`
}

type ResponseHTTPTests struct {
	ResponseRaw
	Data []HTTPTest `json:"result"`
}

type ResponseHosts struct {
	ResponseRaw
	Data []Host `json:"result"`
}

// Response Create/Delete host
type ResponseHostsArray struct {
	ResponseRaw
	Data Hosts `json:"result"`
}

type ResponseGroups struct {
	ResponseRaw
	Data []Group `json:"result"`
}

type ResponseUserGroup struct {
	ResponseRaw
	Data []UserGroup `json:"result"`
}

type ResponseUsers struct {
	ResponseRaw
	Data []User `json:"result"`
}

type ResponseHistory struct {
	ResponseRaw
	Data []History `json:"result"`
}
