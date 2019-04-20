package main

type Host struct {
	ID   string `json:"hostid"`
	Name string `json:"host"`
}

type Hosts struct {
	ID []string `json:"hostids"`
}
