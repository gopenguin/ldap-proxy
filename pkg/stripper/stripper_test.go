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
	"context"
	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/samuel/go-ldap/ldap"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type testBackend struct {
	lastUsername string
	lastPassword string

	result bool
}

func (backend *testBackend) Authenticate(ctx context.Context, username string, password string) bool {
	backend.lastUsername = username
	backend.lastPassword = password

	return backend.result
}

func (backend *testBackend) Name() (name string) {
	return "test"
}

func (backend *testBackend) GetUsers(ctx context.Context, f ldap.Filter) ([]*pkg.User, error) {
	return []*pkg.User{
		{
			Attributes: map[string][]string{
				"uid": {"user1"},
			},
		},
	}, nil
}

func TestStrippingBackend_Authenticate(t *testing.T) {
	Convey("Given a stripping ldap backend", t, func() {
		backend := &testBackend{
			lastUsername: "none",
			lastPassword: "nothing",
			result:       true,
		}
		config := &Config{
			BaseDn:           toPointer("dc=example,dc=com"),
			PeopleRdn:        toPointer("ou=People"),
			UserRdnAttribute: toPointer("uid"),
		}

		stripper := NewBackend(backend, config)

		Convey("When the backend is invoced with a matching dn", func() {
			result := stripper.Authenticate(context.Background(), "uid=admin,ou=People,dc=example,dc=com", "password")

			Convey("Then the user will be stripped, the other parameters will be passed correctly", func() {
				So(result, ShouldBeTrue)
				So(backend.lastUsername, ShouldEqual, "admin")
				So(backend.lastPassword, ShouldEqual, "password")
			})
		})
		Convey("When the backend is invoced with an invalid prefix", func() {
			result := stripper.Authenticate(context.Background(), "cn=admin,ou=People,dc=example,dc=com", "password")

			Convey("Then the backend will return false without calling the delegate", func() {
				So(result, ShouldBeFalse)
				So(backend.lastUsername, ShouldEqual, "none")
				So(backend.lastPassword, ShouldEqual, "nothing")
			})
		})
		Convey("When the backend is invoced with an invalid suffix", func() {
			result := stripper.Authenticate(context.Background(), "uid=admin,dc=com", "password")

			Convey("Then the backend will return false without calling the delegate", func() {
				So(result, ShouldBeFalse)
				So(backend.lastUsername, ShouldEqual, "none")
				So(backend.lastPassword, ShouldEqual, "nothing")
			})
		})
	})
}

func TestStrippingBackend_GetUsers(t *testing.T) {
	Convey("Given a stripping ldap backend", t, func() {
		backend := &testBackend{}

		config := &Config{
			BaseDn:           toPointer("dc=example,dc=com"),
			PeopleRdn:        toPointer("ou=People"),
			UserRdnAttribute: toPointer("uid"),
		}

		stripper := NewBackend(backend, config)

		Convey("When the users are requested", func() {
			users, err := stripper.GetUsers(context.Background(), nil)

			Convey("Then the dn is wrapped with the pre and suffix", func() {
				So(err, ShouldBeNil)
				So(users, ShouldHaveLength, 1)
				So(users[0].DN, ShouldEqual, "uid=user1,ou=People,dc=example,dc=com")
			})
		})
	})
}

func toPointer(value string) *string {
	return &value
}
