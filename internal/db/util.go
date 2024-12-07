package db

import (
	"net/url"
	"strconv"
	"strings"
)

type ConnectionURI struct {
	dialect          string
	driverName       string
	connectionString string
}

func ParseConnectionURI(connection_uri string) (ConnectionURI, error) {
	uri := ConnectionURI{}

	u, err := url.Parse(connection_uri)
	if err != nil {
		return uri, err
	}

	switch u.Scheme {
	case "sqlite":
		uri.dialect = "sqlite"
		uri.driverName = "libsql"
		u.Scheme = "file"
		uri.connectionString = strings.Replace(u.String(), "://", ":", 1)
	case "postgresql":
		uri.dialect = "postgres"
		uri.driverName = "pgx"
		uri.connectionString = u.String()
	}

	return uri, nil

}
func adaptQuery(query string) string {
	if dialect == "sqlite" {
		return query
	}

	var q strings.Builder
	pos := 1

	for _, char := range query {
		if char == '?' {
			q.WriteRune('$')
			q.WriteString(strconv.Itoa(pos))
			pos++
		} else {
			q.WriteRune(char)
		}
	}

	return q.String()
}
