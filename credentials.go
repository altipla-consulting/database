package database

import (
	"fmt"
)

type Credentials struct {
	User, Password     string
	Address, Database  string
	Charset, Collation string
	Protocol           string
}

func (c Credentials) String() string {
	if c.Protocol == "" {
		c.Protocol = "tcp"
	}

	var charset string
	if c.Charset != "" {
		charset = fmt.Sprintf("&charset=%s", c.Charset)
	}
	var collation string
	if c.Collation != "" {
		collation = fmt.Sprintf("&collation=%s", c.Collation)
	}

	return fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true%s%s", c.User, c.Password, c.Protocol, c.Address, c.Database, charset, collation)
}
