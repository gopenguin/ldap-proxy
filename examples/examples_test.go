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

package examples

import (
	"testing"
	"path/filepath"
	"os"
	"github.com/kolleroot/ldap-proxy/pkg/config"
	"github.com/kolleroot/ldap-proxy/pkg/memory"
	"github.com/kolleroot/ldap-proxy/pkg/postgres"
)

func TestExamplesLoadable(t *testing.T) {
	wd, _ := os.Getwd()
	matches, _ := filepath.Glob(filepath.Join(wd, "*.json"))

	loader := config.NewLoader()
	loader.AddFactory(memory.NewFactory())
	loader.AddFactory(postgres.NewFactory())

	for _, match := range matches {
		t.Log(match)

		file, err := os.Open(match)
		if err != nil {
			t.Log(err)
			continue
		}
		defer file.Close()

		_, err = loader.Load(file)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}
}
