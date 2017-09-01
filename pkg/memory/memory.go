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

package memory

import (
	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/samuel/go-ldap/ldap"
)

type backendFactory struct{}

func NewFactory() (factory pkg.BackendFactory) {
	return &backendFactory{}
}

func (backendFactory) Name() (name string) {
	return "in-memory"
}

func (backendFactory) NewConfig() interface{} {
	return &Config{}
}

func (backendFactory) New(untypedConfig interface{}) (bknd pkg.Backend, err error) {
	config, ok := untypedConfig.(*Config)
	if !ok {
		return nil, pkg.ErrInvalidConfigType
	}

	bknd = NewBackend(config)
	return
}

type backend struct {
	config *Config
	users  map[string]User
}

type Config struct {
	pkg.Config
	ListUsers bool   `json:"listUsers"`
	Users     []User `json:"users"`
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func NewBackend(config *Config) (bknd *backend) {
	bknd = &backend{
		config: config,
		users:  make(map[string]User),
	}

	// prepare for lookup
	for _, user := range config.Users {
		bknd.users[user.Name] = user
	}

	return
}

func (backend *backend) Name() (name string) {
	return backend.config.Name
}

func (backend *backend) Authenticate(username string, password string) (successful bool) {
	user, ok := backend.users[username]
	if !ok {
		return false
	}

	return pkg.VerifyPassword(user.Password, password)
}

func (backend *backend) GetUsers(f ldap.Filter) (users []*pkg.User, err error) {
	users = []*pkg.User{}

	if backend.config.ListUsers {
		for _, user := range backend.config.Users {
			users = append(users,
				&pkg.User{
					DN: user.Name,
					Attributes: map[string][]string{
						"cn": {user.Name},
					},
				})
		}
	}

	return
}
