package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/reconquest/karma-go"
)

func handleUsersGroups(
	zabbix *Zabbix,
	config *Config,
	args map[string]interface{},
) error {
	var (
		groups, pattern = parseSearchQuery(args["<pattern>"].([]string))

		addUser, _    = args["--add"].(string)
		removeUser, _ = args["--remove"].(string)
		confirmation  = !args["--noconfirm"].(bool)

		table = tabwriter.NewWriter(os.Stdout, 1, 4, 2, ' ', 0)
	)

	var usersgroups []UserGroup
	var err error

	err = withSpinner(
		":: Requesting information about users groups",
		func() error {
			usersgroups, err = getUsersGroups(zabbix, groups)
			return err
		},
	)

	if err != nil {
		return karma.Format(
			err,
			"can't obtain users groups %s", groups,
		)
	}

	found := false
	for _, group := range usersgroups {
		names := []string{}
		for _, user := range group.Users {
			names = append(names, user.Alias)
		}

		line := fmt.Sprintf(
			"%s\t%s\t%s",
			group.GetStatus(), group.Name,
			strings.Join(names, " "),
		)

		if pattern != "" && !matchPattern(pattern, line) {
			continue
		}

		fmt.Fprintln(table, line)
		found = true
	}

	table.Flush()

	if !found || (addUser == "" && removeUser == "") {
		return nil
	}

	switch {
	case addUser != "":
		if confirmation {
			if !confirmAdding(addUser) {
				return nil
			}
		}

		user, err := getUser(zabbix, addUser)
		if err != nil {
			return karma.Format(
				err,
				"can't obtain user '%s'", addUser,
			)
		}

		err = withSpinner(
			":: Requesting for adding user to specified groups",
			func() error {
				return zabbix.AddUserToGroups(usersgroups, user)
			},
		)

		if err != nil {
			return karma.Format(
				err,
				"can't add user '%s' to specified users groups",
				user.Alias,
			)
		}

	case removeUser != "":
		if confirmation {
			if !confirmRemoving(removeUser) {
				return nil
			}
		}

		user, err := getUser(zabbix, removeUser)
		if err != nil {
			return karma.Format(
				err,
				"can't obtain user '%s'", removeUser,
			)
		}

		err = withSpinner(
			":: Requesting for removing user from specified groups",
			func() error {
				return zabbix.RemoveUserFromGroups(usersgroups, user)
			},
		)

		if err != nil {
			return karma.Format(
				err,
				"can't remove user '%s' from specified users groups",
				removeUser,
			)
		}
	}

	return nil
}

func getUser(zabbix *Zabbix, username string) (User, error) {
	users, err := zabbix.GetUsers(Params{
		"search": Params{
			"alias": username,
		},
	})
	if err != nil {
		return User{}, karma.Format(
			err, "can't obtain user with specified name",
		)
	}

	if len(users) == 0 {
		return User{}, errors.New("user with specified name not found")
	}

	return users[0], nil
}

func getUsersGroups(zabbix *Zabbix, groups []string) ([]UserGroup, error) {
	var params Params
	if len(groups) == 0 {
		params = Params{
			"selectUsers": "1",
		}
	} else {
		params = Params{
			"selectUsers": "1",
			"search": Params{
				"name": groups,
			},
			"searchWildcardsEnabled": "1",
		}
	}

	usersgroups, err := zabbix.GetUsersGroups(params)
	if err != nil {
		return nil, karma.Format(
			err,
			"can't obtain zabbix users groups",
		)
	}

	var (
		usersIdentifiers     = []string{}
		usersIdentifiersHash = map[string]User{}
	)

	for _, usersgroup := range usersgroups {
		for _, user := range usersgroup.Users {
			usersIdentifiers = append(usersIdentifiers, user.ID)
		}
	}

	users, err := zabbix.GetUsers(
		Params{
			"userids": usersIdentifiers,
		},
	)
	if err != nil {
		return nil, karma.Format(
			err,
			"can't obtain users from specified groups %q",
			usersIdentifiers,
		)
	}

	for _, user := range users {
		usersIdentifiersHash[user.ID] = user
	}

	for _, usersgroup := range usersgroups {
		for i, user := range usersgroup.Users {
			usersgroup.Users[i] = usersIdentifiersHash[user.ID]
		}
	}

	return usersgroups, nil
}

func confirmAdding(user string) bool {
	var value string
	fmt.Fprintf(
		os.Stderr,
		"\n:: Proceed with adding user %s to specified groups? [Y/n]: ",
		user,
	)
	fmt.Scanln(&value)
	return value == "" || value == "Y" || value == "y"
}

func confirmRemoving(user string) bool {
	var value string
	fmt.Fprintf(
		os.Stderr,
		"\n:: Proceed with removing user %s from specified groups? [Y/n]: ",
		user,
	)
	fmt.Scanln(&value)
	return value == "" || value == "Y" || value == "y"
}
