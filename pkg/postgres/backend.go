// Copyright © 2017 Stefan Kollmann
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"net/url"

	"errors"
	"fmt"
	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/kolleroot/ldap-proxy/pkg/log"
	"github.com/kolleroot/ldap-proxy/pkg/util"
	"github.com/samuel/go-ldap/ldap"
	sq "gopkg.in/Masterminds/squirrel.v1"
	"regexp"
	"strconv"
	"strings"
)

type Backend struct {
	db     *sql.DB
	config *Config

	userDnRegex *regexp.Regexp

	colAttr map[string]string
	attrCol map[string]string
	cols    []string
	attr    []string
}

var _ pkg.Backend = &Backend{}

type Config struct {
	pkg.Config
	Url     string            `json:"url"`
	Columns map[string]string `json:"columns"`
}

func NewBackend(config *Config) (*Backend, error) {
	parsedUrl, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", parsedUrl.String())
	if err != nil {
		return nil, err
	}

	return newBackend(config, db)
}

func newBackend(config *Config, db *sql.DB) (*Backend, error) {
	backend := Backend{
		db:      db,
		config:  config,
		colAttr: make(map[string]string),
		attrCol: make(map[string]string),
		cols:    []string{},
		attr:    make([]string, len(config.Columns)),
	}

	for col, attr := range config.Columns {
		backend.attrCol[attr] = col
		backend.colAttr[col] = attr

		backend.cols = append(backend.cols, col)
	}

	for i, col := range backend.cols {
		backend.attr[i] = backend.colAttr[col]
	}

	return &backend, nil
}

func (backend *Backend) Name() (name string) {
	return backend.config.Name
}

func (backend *Backend) Authenticate(username string, password string) bool {
	rows, err := backend.db.Query("SELECT password FROM users WHERE name = $1", username)
	if err != nil {
		return false
	}
	defer rows.Close()

	if !rows.Next() {
		return false
	}
	var hashedPassword string
	err = rows.Scan(&hashedPassword)
	if err != nil {
		return false
	}

	log.Debugf("[auth] found user %s", username)

	return util.VerifyPassword(hashedPassword, password)
}

func (backend *Backend) GetUsers(f ldap.Filter) ([]*pkg.User, error) {
	query, args, err := backend.createQuery(f)
	if err != nil {
		return nil, err
	}

	log.Debug(query)

	rows, err := backend.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*pkg.User{}
	columns := make([]interface{}, len(backend.cols))
	var columnsP []interface{}
	for i := range columns {
		columnsP = append(columnsP, &columns[i])
	}

	for rows.Next() {
		if err := rows.Scan(columnsP...); err != nil {
			return nil, err
		}

		user := &pkg.User{
			Attributes: map[string][]string{},
		}

		for i, col := range columns {
			if backend.attr[i] == backend.config.DNAttribute {
				user.DN = col.(string)
			}
			switch col.(type) {
			case string:
				user.Attributes[backend.attr[i]] = []string{col.(string)}
			case int64:
				user.Attributes[backend.attr[i]] = []string{strconv.FormatInt(col.(int64), 10)}
			default:
				return nil, errors.New(fmt.Sprintf("postgres backend: unsupported column type %T", col))
			}
		}

		users = append(users, user)
	}

	return users, nil
}

func (backend *Backend) Close() {
	backend.db.Close()
}

func (backend *Backend) createQuery(f ldap.Filter) (sql string, args []interface{}, err error) {
	log.Debug("convert ldap filter to query")
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		RunWith(backend.db)

	query := psql.
		Select(strings.Join(backend.cols, ", ")).
		From("users")

	if f != nil {
		cond, err := backend.createCondition(f)
		if err != nil {
			log.Debug(err)
			return "", nil, err
		}

		query = query.Where(cond)
	}

	return query.ToSql()
}

func (backend *Backend) createCondition(f ldap.Filter) (cond sq.Sqlizer, err error) {
	switch f.(type) {
	case *ldap.AND:
		a := f.(*ldap.AND)
		log.Debug("START and")

		var ret sq.And
		for _, sa := range a.Filters {
			cond, err := backend.createCondition(sa)
			if err != nil {
				return nil, err
			}
			ret = append(ret, cond)
		}

		log.Debug("END and")
		return ret, nil

	case *ldap.OR:
		o := f.(*ldap.OR)

		log.Debug("START or")

		var ret sq.Or
		for _, sa := range o.Filters {
			cond, err := backend.createCondition(sa)
			if err != nil {
				return nil, err
			}
			ret = append(ret, cond)
		}
		log.Debug("END or")
		return ret, nil

	case *ldap.EqualityMatch:
		e := f.(*ldap.EqualityMatch)

		return backend.equalMatch(e.Attribute, string(e.Value))

	case *ldap.ApproxMatch:
		e := f.(*ldap.ApproxMatch)

		return backend.equalMatch(e.Attribute, string(e.Value))

	case *ldap.Present:
		p := f.(*ldap.Present)

		_, ok := backend.attrCol[p.Attribute]
		return toSqlBool(strings.ToLower(p.Attribute) == "objectclass" || ok), nil

	default:
		return nil, errors.New("unsupported condition type")
	}
}

func (backend *Backend) equalMatch(attr, value string) (sq.Sqlizer, error) {
	log.Debugf("EQ: %s = %s", attr, value)
	if value == "*" {
		return toSqlBool(true), nil
	}

	return sq.Eq{
		backend.attrCol[attr]: value,
	}, nil
}

func toSqlBool(value bool) sq.Sqlizer {
	return &sqlBool{
		value: value,
	}
}

type sqlBool struct {
	value bool
}

var _ sq.Sqlizer = &sqlBool{}

func (this sqlBool) ToSql() (string, []interface{}, error) {
	if this.value {
		return "TRUE", []interface{}{}, nil
	} else {
		return "FALSE", []interface{}{}, nil
	}
}
