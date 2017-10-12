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

package pkg

import (
	"context"
	"errors"
	"github.com/samuel/go-ldap/ldap"
)

// A user inside the ldap structure with a dn and additional attributes. The
// attributes must include the rdn attribute.
type User struct {
	DN         string              // The unique id of the user
	Attributes map[string][]string // Additional information about the user
}

var (
	ErrInvalidConfigType = errors.New("ldap-proxy: invalid configuration object type")
)

type BackendFactory interface {
	Name() string
	NewConfig() interface{}
	New(config interface{}) (backend Backend, err error)
}

type Backend interface {
	Name() (name string)
	Authenticate(ctx context.Context, username string, password string) bool
	GetUsers(ctx context.Context, f ldap.Filter) ([]*User, error)
}

type Config struct {
	Name        string `json:"name"`
	DNAttribute string `json:"dnAttribute"`
}
