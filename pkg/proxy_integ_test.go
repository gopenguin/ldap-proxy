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
	"github.com/gopenguin/ldap-proxy/pkg/util"
	"github.com/samuel/go-ldap/ldap"
	. "github.com/smartystreets/goconvey/convey"
	"path/filepath"
	"testing"
	"time"
)

type integTestBackend struct {
}

var _ Backend = &integTestBackend{}

func (*integTestBackend) Name() string {
	return "integ-test-backend"
}

func (*integTestBackend) Authenticate(ctx context.Context, username string, password string) bool {
	return false
}

func (*integTestBackend) GetUsers(ctx context.Context, f ldap.Filter) ([]*User, error) {
	return []*User{}, nil
}

func TestLdapProxy_Bind2(t *testing.T) {
	Convey("Given a ldap server with a mocked backend and a client connected to it", t, func() {
		dirname, cleanupTmpDir := util.TmpDir(t)
		defer cleanupTmpDir()

		unixSocketPath := filepath.Join(dirname, "ldap-proxy.sock")

		proxy := NewLdapProxy()
		go proxy.ListenAndServe("unix", unixSocketPath)
		err := proxy.server.WaitReady(1 * time.Second)
		So(err, ShouldBeNil)
		defer proxy.server.Close()

		Convey("When the server is bound to", func() {
			client, err := ldap.Dial("unix", unixSocketPath)
			So(err, ShouldBeNil)
			defer client.Close()

			tb := &testBackend{result: true}
			proxy.AddBackend(tb)

			bindErr := client.Bind("testUSER1223", []byte("somecomplicatedPassword"))

			Convey("Then the backend gets the username and password", func() {
				So(bindErr, ShouldBeNil)

				So(tb.lastUsername, ShouldEqual, "testUSER1223")
				So(tb.lastPassword, ShouldEqual, "somecomplicatedPassword")
			})
		})

		Convey("When the server gets a search request", func() {
			client, err := ldap.Dial("unix", unixSocketPath)
			So(err, ShouldBeNil)
			defer client.Close()

			tb := &testBackend{}
			proxy.AddBackend(tb)

			Convey("When the rootDSE is requested", func() {
				res, err := client.Search(&ldap.SearchRequest{})

				Convey("Then the result is returned", func() {
					So(err, ShouldBeNil)
					So(res, ShouldNotBeNil)
				})
			})

			Convey("When some data except the rootDSE requested", func() {
				_, err := client.Search(&ldap.SearchRequest{Scope: ldap.ScopeSingleLevel})

				Convey("Then the request will be rejected", func() {
					So(err, ShouldNotBeNil)
					br, ok := err.(*ldap.BaseResponse)
					So(ok, ShouldBeTrue)
					So(br.Code, ShouldEqual, ldap.ResultInsufficientAccessRights)
				})
			})

			Convey("When a authenticated client", func() {
				tb.result = true

				err := client.Bind("cn=test", []byte("password"))
				So(err, ShouldBeNil)

				Convey("When some data is requested", func() {
					someData := []*User{{DN: "cn=test", Attributes: make(map[string][]string)}}
					someData[0].Attributes["cn"] = []string{"test"}
					tb.user = someData

					res, err := client.Search(&ldap.SearchRequest{Scope: ldap.ScopeSingleLevel})

					Convey("Then the request will be delegated to the backend an the result back to the client", func() {
						So(err, ShouldBeNil)

						So(res, ShouldHaveLength, len(someData))
						for i := range someData {
							So(res[i].DN, ShouldEqual, someData[i].DN)
							for attr, val := range someData[i].Attributes {
								convertedVal := make([][]byte, len(val))
								for i := range val {
									convertedVal[i] = []byte(val[i])
								}

								So(res[i].Attributes[attr], ShouldResemble, convertedVal)
							}
						}
					})
				})
			})
		})
	})
}
