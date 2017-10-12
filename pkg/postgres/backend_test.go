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
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/kolleroot/ldap-proxy/pkg/util"
	"github.com/samuel/go-ldap/ldap"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
)

func TestNewBackend(t *testing.T) {
	Convey("Given a new backend", t, backendWithMockedDatabase(func(backend *Backend, mock sqlmock.Sqlmock) {
		So(backend.config, ShouldNotBeNil)

		So(backend.colAttr, ShouldHaveLength, 4)
		So(backend.attrCol, ShouldHaveLength, 4)
		So(backend.cols, ShouldHaveLength, 4)
		So(backend.attr, ShouldHaveLength, 4)

		So(backend.colAttr, ShouldContainKey, "user")
		So(backend.colAttr, ShouldContainKey, "firstname")
		So(backend.colAttr, ShouldContainKey, "lastname")
		So(backend.colAttr, ShouldContainKey, "email")

		So(backend.attrCol, ShouldContainKey, "uid")
		So(backend.attrCol, ShouldContainKey, "gn")
		So(backend.attrCol, ShouldContainKey, "sn")
		So(backend.attrCol, ShouldContainKey, "email")

		So(backend.colAttr["user"], ShouldEqual, "uid")
		So(backend.colAttr["firstname"], ShouldEqual, "gn")
		So(backend.colAttr["lastname"], ShouldEqual, "sn")
		So(backend.colAttr["email"], ShouldEqual, "email")

		So(backend.attrCol["uid"], ShouldEqual, "user")
		So(backend.attrCol["gn"], ShouldEqual, "firstname")
		So(backend.attrCol["sn"], ShouldEqual, "lastname")
		So(backend.attrCol["email"], ShouldEqual, "email")
	}))
}

func TestBackend_CreateUser(t *testing.T) {
	Convey("Given some credentials and a database", t, backendWithMockedDatabase(func(backend *Backend, mock sqlmock.Sqlmock) {
		Convey("When a new user is inserted", func() {
			passwordCatcher := newArgumentCatcher()
			mock.ExpectExec("INSERT INTO users").
				WithArgs("userA", passwordCatcher.Catcher()).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := backend.CreateUser("userA", "test123")
			So(err, ShouldBeNil)

			Convey("Then a new rows should be inserted into the db", func() {
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				password, ok := passwordCatcher.Value.(string)
				So(ok, ShouldBeTrue)
				Println(password)
				So(util.VerifyPassword(password, "test123"), ShouldBeTrue)
			})
		})
	}))
}

func TestBackend_Authenticate(t *testing.T) {
	Convey("Given a mocked database with a user 'userA'", t, backendWithMockedDatabase(func(backend *Backend, mock sqlmock.Sqlmock) {

		useraRows := sqlmock.NewRows([]string{"password"}).
			AddRow("$2a$04$7aS0AmbLn./PTc0DpX2XeOpKV2VPM6RRrooSHsG/n.zolLV78BGny")
		emptyRows := sqlmock.NewRows([]string{"password"})

		Convey("User userA should be able to authenticate with 'test123'", func() {
			mock.ExpectQuery("^SELECT (.+) FROM users WHERE name = \\$1$").WithArgs("userA").WillReturnRows(useraRows)
			So(backend.Authenticate(context.Background(), "userA", "test123"), ShouldBeTrue)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
		Convey("User userA should not be able to authenticate with 'wrongPassword'", func() {
			mock.ExpectQuery("^SELECT (.+) FROM users WHERE name = \\$1$").WithArgs("userA").WillReturnRows(useraRows)
			So(backend.Authenticate(context.Background(), "userA", "wrongPassword"), ShouldBeFalse)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
		Convey("User userB should be able to authenticate with 'test123'", func() {
			mock.ExpectQuery("^SELECT (.+) FROM users WHERE name = \\$1$").WithArgs("userB").WillReturnRows(emptyRows)
			So(backend.Authenticate(context.Background(), "userB", "test123"), ShouldBeFalse)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	}))
}

func TestBackend_GetUsers(t *testing.T) {
	Convey("Given a mocked database with a user 'userA'", t, backendWithMockedDatabase(func(backend *Backend, mock sqlmock.Sqlmock) {
		useraData := map[string]string{
			"user":      "userA",
			"email":     "user-a@example.com",
			"firstname": "a",
			"lastname":  "user",
		}

		row := make([]driver.Value, len(backend.cols))
		for i, col := range backend.cols {
			row[i] = useraData[col]
		}

		useraRows := sqlmock.NewRows(backend.cols).AddRow(row...)

		Convey("When the users are requested", func() {
			mock.ExpectQuery("^SELECT (.+) FROM users").WillReturnRows(useraRows)
			users, err := backend.GetUsers(context.Background(), nil)

			Convey("Then userA will be returned", func() {
				assertUserA(users, err)
			})
		})

		Convey("Given a filter", func() {
			filter := &ldap.AND{
				Filters: []ldap.Filter{
					&ldap.EqualityMatch{Attribute: "gn", Value: []byte("a")},
					&ldap.EqualityMatch{Attribute: "sn", Value: []byte("user")},
				},
			}

			Convey("When the users are requested", func() {
				mock.ExpectQuery("^SELECT (.+) FROM users WHERE \\(firstname = \\$1 AND lastname = \\$2\\)").WithArgs("a", "user").WillReturnRows(useraRows)
				users, err := backend.GetUsers(context.Background(), filter)

				Convey("Then userA will be returned", func() {
					assertUserA(users, err)
				})
			})
		})
	}))
}

func backendWithMockedDatabase(test func(backend *Backend, mock sqlmock.Sqlmock)) func() {
	return func() {
		db, mock, err := sqlmock.New()
		So(err, ShouldBeNil)

		config := &Config{
			Config: pkg.Config{
				DNAttribute: "uid",
			},
			Columns: map[string]string{
				"user":      "uid",
				"firstname": "gn",
				"lastname":  "sn",
				"email":     "email",
			},
		}

		backend, err := newBackend(config, db)
		So(err, ShouldBeNil)

		test(backend, mock)
	}
}

func assertUserA(users []*pkg.User, err error) {
	So(err, ShouldBeNil)
	So(users, ShouldHaveLength, 1)
	So(users[0].Attributes["uid"][0], ShouldEqual, "userA")
	So(users[0].Attributes["gn"][0], ShouldEqual, "a")
	So(users[0].Attributes["sn"][0], ShouldEqual, "user")
	So(users[0].Attributes["email"][0], ShouldEqual, "user-a@example.com")

	for k, v := range users[0].Attributes {
		fmt.Println(k, v)
	}
}

func newArgumentCatcher() argumentCatcher {
	return argumentCatcher{}
}

type argumentCatcher struct {
	Value driver.Value
}

func (catcher *argumentCatcher) Match(value driver.Value) bool {
	catcher.Value = value

	return true
}

func (catcher *argumentCatcher) Catcher() sqlmock.Argument {
	return catcher
}
