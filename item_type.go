package main

import (
	"encoding/json"
	"strconv"
)

type ItemType int

const (
	ItemTypeAgent ItemType = iota
	ItemTypeSNMPv1
	ItemTypeTrapper
	ItemTypeSimpleCheck
	ItemTypeSNMPv2
	ItemTypeInternal
	ItemTypeSNMPv3
	ItemTypeAgentActive
	ItemTypeAggregate
	ItemTypeWeb
	ItemTypeExternalCheck
	ItemTypeDatabaseMonitor
	ItemTypeIPMI
	ItemTypeSSH
	ItemTypeTELNET
	ItemTypeCalculated
	ItemTypeJMX
	ItemTypeSNMPTrap
)

func (type_ *ItemType) UnmarshalJSON(data []byte) error {
	var stringValue string

	err := json.Unmarshal(data, &stringValue)
	if err != nil {
		return err
	}

	intValue, err := strconv.ParseInt(stringValue, 10, 64)
	if err != nil {
		return err
	}

	*type_ = ItemType(intValue)

	return nil
}

func (type_ ItemType) String() string {
	switch type_ {
	case ItemTypeAgent:
		return "agent"
	case ItemTypeTrapper:
		return "trapper"
	case ItemTypeSimpleCheck:
		return "check"
	case ItemTypeSNMPv2:
		return "snmp2"
	case ItemTypeInternal:
		return "internal"
	case ItemTypeSNMPv3:
		return "snmp3"
	case ItemTypeAgentActive:
		return "active"
	case ItemTypeAggregate:
		return "aggregate"
	case ItemTypeWeb:
		return "web"
	case ItemTypeExternalCheck:
		return "external"
	case ItemTypeDatabaseMonitor:
		return "dbmon"
	case ItemTypeIPMI:
		return "ipmi"
	case ItemTypeSSH:
		return "ssh"
	case ItemTypeTELNET:
		return "telnet"
	case ItemTypeCalculated:
		return "calc"
	case ItemTypeJMX:
		return "jmx"
	case ItemTypeSNMPTrap:
		return "snmp"
	default:
		return "unknown"
	}
}
