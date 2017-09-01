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
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewMemoryBackend(t *testing.T) {
	Convey("Given a memory backend", t, func() {
		backend := NewBackend(&Config{
			Users: []User{},
		})

		Convey("Expect the backend to implement the backend interface", func() {
			converted := pkg.Backend(backend)
			So(converted, ShouldNotBeNil)
		})
	})
}

func TestBackend_Authenticate(t *testing.T) {
	Convey("Given a memory backend", t, func() {
		backend := NewBackend(&Config{
			Users: []User{
				{Name: "user1", Password: "$2a$04$7aS0AmbLn./PTc0DpX2XeOpKV2VPM6RRrooSHsG/n.zolLV78BGny"},
			},
		})

		Convey("When user1 authenticates", func() {
			result := backend.Authenticate("user1", "test123")

			Convey("Then authentication succeeds", func() {
				So(result, ShouldBeTrue)
			})
		})

		Convey("When user2 authenticates", func() {
			result := backend.Authenticate("user2", "test123")

			Convey("Then authentication fails", func() {
				So(result, ShouldBeFalse)
			})
		})
	})
}

func TestBackend_GetUsers(t *testing.T) {
	Convey("Given a memory backend", t, func() {
		backend := NewBackend(&Config{
			Users: []User{
				{Name: "user1", Password: "$2a$04$7aS0AmbLn./PTc0DpX2XeOpKV2VPM6RRrooSHsG/n.zolLV78BGny"},
			},
		})

		Convey("Given the config is set to list users", func() {
			backend.config.ListUsers = true

			Convey("When the users are listed", func() {
				users, err := backend.GetUsers(nil)

				Convey("Then the users will be returned", func() {
					So(err, ShouldBeNil)
					So(users, ShouldHaveLength, 1)
					So(users[0].DN, ShouldEqual, "user1")
					So(users[0].Attributes["cn"][0], ShouldEqual, "user1")
				})
			})
		})

		Convey("Given the config is set to not list users", func() {
			backend.config.ListUsers = false

			Convey("When the users are listed", func() {
				users, err := backend.GetUsers(nil)

				Convey("Then no users are returned", func() {
					So(err, ShouldBeNil)
					So(users, ShouldHaveLength, 0)
				})
			})
		})
	})
}
