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
	"github.com/kolleroot/ldap-proxy/pkg"
	jww "github.com/spf13/jwalterweatherman"
)

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
