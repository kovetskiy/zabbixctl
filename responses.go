package main

import "github.com/reconquest/hierr-go"

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
		return hierr.Push(
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

type ResponseTriggers struct {
	ResponseRaw
	Data map[string]Trigger `json:"result"`
}

type ResponseItems struct {
	ResponseRaw
	Data []Item `json:"result"`
}

type ResponseHosts struct {
	ResponseRaw
	Data []Host `json:"result"`
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
