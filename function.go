package main

type Function struct {
	ID        string `json:"functionid"`
	ItemID    string `json:"itemid"`
	TriggerID string `json:"triggerid"`
	Name      string `json:"function"`
	Parameter string `json:"parameter"`
}
