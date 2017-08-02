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

	"github.com/kolleroot/ldap-proxy/pkg"
	jww "github.com/spf13/jwalterweatherman"
	"regexp"
)

type backendFactory struct{}

func NewFactory() (factory pkg.BackendFactory) {
	return &backendFactory{}
}

func (backendFactory) Name() (name string) {
	return "postgres"
}

func (backendFactory) NewConfig() interface{} {
	return &Config{}
}

func (backendFactory) New(untypedConfig interface{}) (bknd pkg.Backend, err error) {
	config, ok := untypedConfig.(*Config)
	if !ok {
		return nil, pkg.ErrInvalidConfigType
	}

	bknd, err = NewBackend(config)
	if err != nil {
		return nil, err
	}

	return
}

type Backend struct {
	db     *sql.DB
	config *Config

	userDnRegex *regexp.Regexp
}

type Config struct {
	pkg.Config
	Url string `json:"url"`
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

	backend := Backend{
		db:     db,
		config: config,
	}

	return &backend, nil
}

// Methods in LdapBackend

func (backend *Backend) Init() error {
	jww.INFO.Print("creating table users ... ")
	_, err := backend.db.Exec("CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(256), password VARCHAR(1024), email VARCHAR(256), firstname VARCHAR(256), lastname VARCHAR(256))")
	if err != nil {
		return err
	}

	err = backend.validateDatabaseStructure()
	if err != nil {
		return err
	}

	return nil
}

func (backend *Backend) Cleanup() error {
	jww.INFO.Print("deleting tabel users ...")
	_, err := backend.db.Exec("DROP TABLE users")
	if err != nil {
		jww.FATAL.Fatal(err)
	}

	return nil
}

func (backend *Backend) Close() {
	backend.db.Close()
}

// public methods

func (backend *Backend) CreateUser(name string, password string) error {
	hash := pkg.HashPassword(password, 12)

	jww.INFO.Print("Password hashed ...")

	res, err := backend.db.Exec("INSERT INTO users (name, password) VALUES ($1, $2)", name, string(hash))
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	jww.INFO.Printf("%d User inserted ...", rows)
	return nil
}

func (backend *Backend) validateDatabaseStructure() error {
	_, err := backend.db.Query("SELECT name, password FROM users LIMIT 1")

	if err != nil {
		return err
	}

	return nil
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

	jww.INFO.Printf("found user %s", username)

	return pkg.VerifyPassword(hashedPassword, password)
}

func (backend *Backend) GetUsers() ([]*pkg.User, error) {
	rows, err := backend.db.Query("SELECT name, email, firstname, lastname FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*pkg.User{}
	for rows.Next() {
		var name, email, firstname, lastname string
		if err := rows.Scan(&name, &email, &firstname, &lastname); err != nil {
			return nil, err
		}

		users = append(users, &pkg.User{
			DN: name,
			Attributes: map[string][]string{
				"uid":   {name},
				"gn":    {firstname},
				"sn":    {lastname},
				"email": {email},
			},
		})
	}

	return users, nil
}
