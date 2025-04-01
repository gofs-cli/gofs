package model

type Diag struct {
	Severity int
	Message  string
}

const (
	SeverityError       = 1
	SeverityWarning     = 2
	SeverityInformation = 3
	SeverityHint        = 4
)

type Pos struct {
	Line int
	Col  int
}
