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
	"github.com/samuel/go-ldap/ldap"
	. "github.com/smartystreets/goconvey/convey"
	jww "github.com/spf13/jwalterweatherman"
	"testing"
)

type wrongSession struct{}

func TestLdapProxy_Connect(t *testing.T) {
	Convey("Given a ldap proxy", t, func() {
		proxy := NewLdapProxy()

		So(proxy, ShouldNotBeNil)

		Convey("When there is a new connection", func() {
			ctx, err := proxy.Connect(nil)

			Convey("Then a new session should be created", func() {
				So(err, ShouldBeNil)

				session, ok := ctx.(*session)
				So(ok, ShouldBeTrue)
				So(session, ShouldNotBeNil)
			})
		})

	})
}

func TestLdapproxy_Bind(t *testing.T) {
	Convey("Given a ldap proxy", t, func() {
		proxy := NewLdapProxy()

		Convey("When there is a bind requist with an invalid session", func() {
			id, err := proxy.Bind(&wrongSession{}, nil)

			Convey("Then an error should be returned", func() {
				So(err, ShouldEqual, errInvalidSessionType)
				So(id, ShouldBeNil)
			})
		})

		Convey("Given there is no backend", func() {
			Convey("When there is a bind request", func() {
				sess := &session{Logger: jww.INFO}
				res, err := proxy.Bind(sess, &ldap.BindRequest{
					DN:       "uid=test,ou=People,dc=example,dc=com",
					Password: []byte("secure"),
				})

				Convey("Then the bind fails with 'Invalid Credentials' and the session is uninitialized", func() {
					So(err, ShouldBeNil)
					So(res.Code, ShouldEqual, ldap.ResultInvalidCredentials)
					So(sess.dn, ShouldBeBlank)
				})
			})
		})

		Convey("Given a generic backend", func() {
			tb := &testBackend{
				result: true,
			}

			proxy.AddBackend(tb)
			So(proxy.backends["test"], ShouldEqual, tb)

			Convey("When there is a bind request", func() {
				dn := "uid=test,ou=People,dc=example,dc=com"
				pw := "secure"

				sess := &session{Logger: jww.INFO}
				res, err := proxy.Bind(sess, &ldap.BindRequest{
					DN:       dn,
					Password: []byte(pw),
				})

				Convey("Then the backend is invoced", func() {
					So(err, ShouldBeNil)
					So(res.Code, ShouldEqual, ldap.ResultSuccess)
					So(res.MatchedDN, ShouldEqual, dn)
					So(tb.lastUsername, ShouldEqual, dn)
					So(tb.lastPassword, ShouldEqual, pw)
				})
			})
		})
	})
}

func TestLdapProxy_Whoami(t *testing.T) {
	Convey("Given a ldap proxy", t, func() {
		proxy := NewLdapProxy()

		Convey("When there is a whoami request", func() {
			id, err := proxy.Whoami(&session{
				dn:     "uid=test,ou=People,dc=example,dc=com",
				Logger: jww.INFO,
			})

			Convey("Then the id should be returned", func() {
				So(err, ShouldBeNil)
				So(id, ShouldEqual, "uid=test,ou=People,dc=example,dc=com")
			})
		})

		Convey("When there is a whoami requist with an invalid session", func() {
			id, err := proxy.Whoami(&wrongSession{})

			Convey("Then an error should be returned", func() {
				So(err, ShouldEqual, errInvalidSessionType)
				So(id, ShouldBeBlank)
			})
		})
	})
}
