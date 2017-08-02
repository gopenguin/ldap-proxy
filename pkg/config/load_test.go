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

package config

import (
	"bytes"
	"errors"
	"github.com/kolleroot/ldap-proxy/pkg"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

type testFactory struct {
	lastConfig *testConfig
}

func (*testFactory) Name() (name string) {
	return "test"
}

func (*testFactory) NewConfig() (config interface{}) {
	return &testConfig{}
}

func (factory *testFactory) New(generalConfig interface{}) (backend pkg.Backend, err error) {
	factory.lastConfig = generalConfig.(*testConfig)
	if factory.lastConfig.TestValue == "fail" {
		return nil, errors.New("config: test error")
	}
	return &testBackend{}, nil
}

type testConfig struct {
	TestValue string `json:"value"`
}

type testBackend struct{}

func (testBackend) Name() (name string) {
	return "test"
}

func (testBackend) Authenticate(username string, password string) bool {
	return false
}

func (testBackend) GetUsers() ([]*pkg.User, error) {
	return []*pkg.User{}, nil
}

func TestNewLoader(t *testing.T) {
	Convey("Given a loader", t, func() {
		loader := NewLoader()

		Convey("A new loader instance should be returned", func() {
			So(loader, ShouldNotBeNil)
			So(loader.factories, ShouldNotBeNil)
		})
	})
}

func TestLoader_Load(t *testing.T) {
	Convey("Given a loader", t, func() {
		tf := &testFactory{}

		loader := NewLoader()
		loader.AddFactory(tf)

		Convey("When loading a config", func() {
			backends, err := loader.Load(toReader(`[{"kind": "test", "value": "testValue"}]`))

			Convey("Then a test backend should be loaded", func() {
				So(err, ShouldBeNil)
				So(backends, ShouldHaveLength, 1)
				So(tf.lastConfig.TestValue, ShouldEqual, "testValue")
			})
		})

		Convey("When the general config is invalid", func() {
			backends, err := loader.Load(toReader(`{}`))

			Convey("Then an error should be returned", func() {
				So(backends, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the backend config is invalid", func() {
			backends, err := loader.Load(toReader(`[{"kind": "test", "value": "fail"}]`))

			Convey("Then an error should be returned", func() {
				So(backends, ShouldHaveLength, 0)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When there is a partial stripper config", func() {
			backends, err := loader.Load(toReader(`[{"kind": "test", "value": "testValue", "baseDn": "dc=example,dc=com"}]`))

			Convey("Then the backend won't be wrapped", func() {
				So(err, ShouldBeNil)
				So(backends, ShouldHaveLength, 1)
				So(backends[0], ShouldHaveSameTypeAs, &testBackend{})
			})
		})

		Convey("When there is a complete stripper config", func() {
			backends, err := loader.Load(toReader(`[{"kind": "test", "value": "testValue", "baseDn": "dc=example,dc=com", "peopleRdn": "ou=People", "userRdnAttribute": "uid"}]`))

			Convey("Then the backend won't be wrapped", func() {
				So(err, ShouldBeNil)
				So(backends, ShouldHaveLength, 1)
				So(backends[0], ShouldNotHaveSameTypeAs, &testBackend{})
			})
		})
	})
}

func toReader(data string) io.Reader {
	return bytes.NewReader([]byte(data))
}
