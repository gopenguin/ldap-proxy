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
	"context"
	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/kolleroot/ldap-proxy/pkg/util"
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

func (backend *backend) Authenticate(ctx context.Context, username string, password string) (successful bool) {
	user, ok := backend.users[username]
	if !ok {
		return false
	}

	return util.VerifyPasswordCtx(ctx, user.Password, password)
}

func (backend *backend) GetUsers(ctx context.Context, f ldap.Filter) (users []*pkg.User, err error) {
	users = []*pkg.User{}

	if backend.config.ListUsers {
		for _, user := range backend.config.Users {
			if f == nil || backend.filterUser(user, f) {
				users = append(users,
					&pkg.User{
						DN: user.Name,
						Attributes: map[string][]string{
							"cn": {user.Name},
						},
					})
			}
		}
	}

	return
}

func (backend *backend) filterUser(user User, f ldap.Filter) bool {
	switch f.(type) {
	case *ldap.AND:
		a := f.(*ldap.AND)
		res := true

		for _, filter := range a.Filters {
			res = res && backend.filterUser(user, filter)
		}
		return res

	case *ldap.OR:
		a := f.(*ldap.OR)
		res := false

		for _, filter := range a.Filters {
			res = res || backend.filterUser(user, filter)
		}
		return res

	case *ldap.EqualityMatch:
		e := f.(*ldap.EqualityMatch)

		switch e.Attribute {
		case "cn":
			return string(e.Value) == user.Name
		}

	case *ldap.ApproxMatch:
		a := f.(*ldap.ApproxMatch)

		switch a.Attribute {
		case "cn":
			return string(a.Value) == user.Name
		}

	case *ldap.Present:
		p := f.(*ldap.Present)

		switch p.Attribute {
		case "cn":
			return true
		}
	}

	return false
}
