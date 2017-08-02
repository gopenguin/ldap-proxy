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
	"database/sql/driver"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
	"github.com/kolleroot/ldap-proxy/pkg"
)

func TestBackend_CreateUser(t *testing.T) {
	Convey("Given some credentials and a database", t, backendWithMockedDatabase(func(backend Backend, mock sqlmock.Sqlmock) {
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
				So(pkg.VerifyPassword(password, "test123"), ShouldBeTrue)
			})
		})
	}))
}

func TestBackend_Authenticate(t *testing.T) {
	Convey("Given a mocked database with a user 'userA'", t, backendWithMockedDatabase(func(backend Backend, mock sqlmock.Sqlmock) {

		useraRows := sqlmock.NewRows([]string{"password"}).
			AddRow("$2a$04$7aS0AmbLn./PTc0DpX2XeOpKV2VPM6RRrooSHsG/n.zolLV78BGny")
		emptyRows := sqlmock.NewRows([]string{"password"})

		Convey("User userA should be able to authenticate with 'test123'", func() {
			mock.ExpectQuery("^SELECT (.+) FROM users WHERE name = \\$1$").WithArgs("userA").WillReturnRows(useraRows)
			So(backend.Authenticate("userA", "test123"), ShouldBeTrue)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
		Convey("User userA should not be able to authenticate with 'wrongPassword'", func() {
			mock.ExpectQuery("^SELECT (.+) FROM users WHERE name = \\$1$").WithArgs("userA").WillReturnRows(useraRows)
			So(backend.Authenticate("userA", "wrongPassword"), ShouldBeFalse)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
		Convey("User userB should be able to authenticate with 'test123'", func() {
			mock.ExpectQuery("^SELECT (.+) FROM users WHERE name = \\$1$").WithArgs("userB").WillReturnRows(emptyRows)
			So(backend.Authenticate("userB", "test123"), ShouldBeFalse)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	}))
}

func TestBackend_GetUsers(t *testing.T) {
	Convey("Given a mocked database with a user 'userA'", t, backendWithMockedDatabase(func(backend Backend, mock sqlmock.Sqlmock) {
		useraRows := sqlmock.NewRows([]string{"user", "email", "firstname", "lastname"}).
			AddRow("userA", "user-a@example.com", "a", "user")

		Convey("When the users are requested", func() {
			mock.ExpectQuery("^SELECT (.+) FROM users").WillReturnRows(useraRows)
			users, err := backend.GetUsers()

			Convey("Then userA will be returned", func() {
				So(err, ShouldBeNil)
				So(users, ShouldHaveLength, 1)
				So(users[0].Attributes["uid"][0], ShouldEqual, "userA")
				So(users[0].Attributes["gn"][0], ShouldEqual, "a")
				So(users[0].Attributes["sn"][0], ShouldEqual, "user")
				So(users[0].Attributes["email"][0], ShouldEqual, "user-a@example.com")
			})
		})
	}))
}

func backendWithMockedDatabase(test func(backend Backend, mock sqlmock.Sqlmock)) func() {
	return func() {
		db, mock, err := sqlmock.New()
		So(err, ShouldBeNil)

		backend := Backend{
			db: db,
		}

		test(backend, mock)
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
