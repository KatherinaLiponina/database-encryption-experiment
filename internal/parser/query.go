package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xwb1989/sqlparser"
)

var ErrNoFrom = fmt.Errorf("no from part in query")

func ExtractTable(query string) (string, error) {
	query = strings.ToLower(query)
	posFrom := strings.Index(query, "from")
	if posFrom == -1 {
		return "", ErrNoFrom
	}
	posEnd := strings.Index(query, "where")
	if posEnd == -1 {
		posEnd = len(query)
	}
	return strings.TrimSpace(query[posFrom+len("from") : posEnd]), nil
}

func ExtractAttribute(query string, position int) (string, error) {
	query = strings.ToLower(query)
	pos := strings.Index(query, fmt.Sprintf("$%d", position))
	if pos == -1 {
		return "", fmt.Errorf("position not found: %s <- %d", query, position)
	}
	return "", nil

}

type QueryParser struct {
	expr *regexp.Regexp
}

func NewQueryParser() QueryParser {
	return QueryParser{expr: regexp.MustCompile(`(\$[\d]*)`)}
}

func (p *QueryParser) FindTable(query string) (string, error) {
	query = string(p.expr.ReplaceAll([]byte(query), []byte("'$1'")))

	r := strings.NewReader(query)
	tokens := sqlparser.NewTokenizer(r)

	for {
		t, _ := tokens.Scan()
		if t == 0 {
			return "", fmt.Errorf("no from token found")
		}
		if t == sqlparser.FROM {
			break
		}
	}
	_, val := tokens.Scan()
	if t, _ := tokens.Scan(); t != sqlparser.WHERE {
		return "", fmt.Errorf("select from embedded queries is not allowed")
	}
	return string(val), nil
}

func (p *QueryParser) FindSelected(query string) ([]string, error) {
	query = string(p.expr.ReplaceAll([]byte(query), []byte("'$1'")))

	r := strings.NewReader(query)
	tokens := sqlparser.NewTokenizer(r)

	for {
		t, _ := tokens.Scan()
		if t == 0 {
			return nil, fmt.Errorf("no select token found")
		}
		if t == sqlparser.SELECT {
			break
		}
	}

	var selected []string
	for {
		t, val := tokens.Scan()
		if t == 0 || t == sqlparser.FROM {
			return selected, nil
		}
		if t == sqlparser.ID {
			selected = append(selected, string(val))
		}
	}
}

func (p *QueryParser) GetAttribute(query string, position int) (string, error) {
	query = string(p.expr.ReplaceAll([]byte(query), []byte("'$1'")))

	r := strings.NewReader(query)
	tokens := sqlparser.NewTokenizer(r)

	var possibleID string

	for {
		t, val := tokens.Scan()
		if t == 0 {
			return "", fmt.Errorf("no expected token found")
		}
		if t == sqlparser.ID {
			possibleID = string(val)
		}
		if t == sqlparser.STRING && string(val) == fmt.Sprintf("$%d", position) {
			return possibleID, nil
		}
	}
}

