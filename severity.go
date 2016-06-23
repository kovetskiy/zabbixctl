package main

type Severity int

var (
	SeverityInformation Severity = 1
	SeverityWarning     Severity = 2
	SeverityAverage     Severity = 3
	SeverityHigh        Severity = 4
	SeverityDisaster    Severity = 5
)

func (priority Severity) String() string {
	switch priority {
	case SeverityInformation:
		return "INFO"
	case SeverityWarning:
		return "WARN"
	case SeverityAverage:
		return "AVG"
	case SeverityHigh:
		return "HIGH"
	case SeverityDisaster:
		return "DISASTER"
	default:
		return "UNKNOWN"
	}
}
