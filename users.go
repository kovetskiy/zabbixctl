package main

type User struct {
	ID    string `json:"userid"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type UserGroup struct {
	ID     string `json:"usrgrpid"`
	Name   string `json:"name"`
	Status string `json:"users_status"`
	Users  []User `json:"users"`
}

func (group *UserGroup) GetStatus() string {
	if group.Status == "0" {
		return "enabled"
	}

	return "disabled"
}
