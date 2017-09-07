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

package stripper

import (
	"fmt"
	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/samuel/go-ldap/ldap"
	jww "github.com/spf13/jwalterweatherman"
	"strings"
)

type Config struct {
	pkg.Config

	BaseDn           *string `json:"baseDn"`
	PeopleRdn        *string `json:"peopleRdn"`
	UserRdnAttribute *string `json:"userRdnAttribute"`
}

type strippingBackend struct {
	delegateBackend pkg.Backend
	config          *Config
}

func NewBackend(delegateBackend pkg.Backend, config *Config) (backend pkg.Backend) {
	return &strippingBackend{
		delegateBackend: delegateBackend,
		config:          config,
	}
}

func (backend *strippingBackend) Name() (name string) {
	return backend.config.Name
}

func (backend *strippingBackend) Authenticate(username string, password string) bool {
	suffix := backend.config.suffix()

	if !strings.HasSuffix(username, suffix) {
		return false // wrong suffix, doesn't match the base dn and people rdn
	}

	prefix := backend.config.prefix()

	if !strings.HasPrefix(username, prefix) {
		return false // wrong prefix, doesn't match the user rdn attribute
	}

	strippedUsername := strings.TrimPrefix(strings.TrimSuffix(username, suffix), prefix)

	jww.INFO.Printf("stripped user %s", strippedUsername)

	return backend.delegateBackend.Authenticate(strippedUsername, password)
}

func (backend *strippingBackend) GetUsers(f ldap.Filter) ([]*pkg.User, error) {
	users, err := backend.delegateBackend.GetUsers(f)

	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.DN = backend.config.formatUserDn(user.Attributes[*backend.config.UserRdnAttribute][0])
	}

	return users, nil
}

func (config *Config) suffix() string {
	return fmt.Sprintf(",%s,%s", *config.PeopleRdn, *config.BaseDn)
}

func (config *Config) prefix() string {
	return fmt.Sprintf("%s=", *config.UserRdnAttribute)
}

func (config *Config) formatUserDn(username string) string {
	return fmt.Sprintf("%s=%s,%s,%s", *config.UserRdnAttribute, username, *config.PeopleRdn, *config.BaseDn)
}
