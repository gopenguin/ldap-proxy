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
	"github.com/samuel/go-ldap/ldap"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateCondition(t *testing.T) {
	testBackend := Backend{
		colAttr: map[string]string{
			"cn": "cn",
			"l":  "l",
		},
		attrCol: map[string]string{
			"cn": "cn",
			"l":  "l",
		},
		cols: []string{"cn", "l"},
		attr: []string{"cn", "l"},
	}

	Convey("Given an equality test", t, func() {
		f := &ldap.EqualityMatch{
			Attribute: "cn",
			Value:     []byte("test"),
		}

		Convey("Expect a query", func() {
			cond, err := testBackend.createCondition(f)
			sql, params, _ := cond.ToSql()

			So(err, ShouldBeNil)
			So(sql, ShouldEqual, `cn = ?`)
			So(params, ShouldHaveLength, 1)
			So(params[0], ShouldEqual, "test")
		})
	})

	Convey("Given a present test", t, func() {
		f := &ldap.Present{
			Attribute: "cn",
		}

		Convey("Expect a query with true", func() {
			cond, err := testBackend.createCondition(f)
			sql, params, _ := cond.ToSql()

			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "TRUE")
			So(params, ShouldHaveLength, 0)
		})

		f = &ldap.Present{
			Attribute: "gn",
		}

		Convey("Expect a query with false", func() {
			cond, err := testBackend.createCondition(f)
			sql, params, _ := cond.ToSql()

			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "FALSE")
			So(params, ShouldHaveLength, 0)
		})
	})

	Convey("Given an AND conjecture", t, func() {
		f := &ldap.AND{
			Filters: []ldap.Filter{
				&ldap.EqualityMatch{
					Attribute: "cn",
					Value:     []byte("test"),
				},
				&ldap.EqualityMatch{
					Attribute: "l",
					Value:     []byte("UK"),
				},
			},
		}

		Convey("Expect a nested AND query", func() {
			cond, err := testBackend.createCondition(f)
			sql, params, _ := cond.ToSql()

			So(err, ShouldBeNil)
			So(sql, ShouldEqual, `(cn = ? AND l = ?)`)
			So(params, ShouldHaveLength, 2)
			So(params[0], ShouldEqual, "test")
			So(params[1], ShouldEqual, "UK")
		})
	})

	Convey("Given an OR conjecture", t, func() {
		f := &ldap.OR{
			Filters: []ldap.Filter{
				&ldap.EqualityMatch{
					Attribute: "cn",
					Value:     []byte("test"),
				},
				&ldap.EqualityMatch{
					Attribute: "l",
					Value:     []byte("UK"),
				},
			},
		}

		Convey("Expect a nested OR query", func() {
			cond, err := testBackend.createCondition(f)
			sql, params, _ := cond.ToSql()

			So(err, ShouldBeNil)
			So(sql, ShouldEqual, `(cn = ? OR l = ?)`)
			So(params, ShouldHaveLength, 2)
			So(params[0], ShouldEqual, "test")
			So(params[1], ShouldEqual, "UK")
		})
	})
}
