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
	"encoding/json"
	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/kolleroot/ldap-proxy/pkg/stripper"
	jww "github.com/spf13/jwalterweatherman"
	"io"
)

type Loader struct {
	factories map[string]pkg.BackendFactory
}

type typedConfig struct {
	Kind string `json:"kind"`
}

func NewLoader() (loader *Loader) {
	loader = &Loader{
		factories: make(map[string]pkg.BackendFactory),
	}

	return
}

func (loader *Loader) AddFactory(factory pkg.BackendFactory) {
	jww.INFO.Printf("Adding backend factory %s", factory.Name())

	loader.factories[factory.Name()] = factory
}

func (loader *Loader) Load(reader io.Reader) (backends []pkg.Backend, err error) {
	var rawConfigs []*json.RawMessage

	decoder := json.NewDecoder(reader)

	if err = decoder.Decode(&rawConfigs); err != nil {
		return nil, err
	}

	backends = []pkg.Backend{}

	for _, rawConfig := range rawConfigs {
		backend, err := loader.instantiateBackend(*rawConfig)
		if err != nil {
			return nil, err
		}

		backends = append(backends, backend)
	}

	return
}

func (loader *Loader) instantiateBackend(data json.RawMessage) (backend pkg.Backend, err error) {
	kindWrapper := &typedConfig{}
	err = json.Unmarshal(data, kindWrapper)
	if err != nil {
		return nil, err
	}

	factory := loader.factories[kindWrapper.Kind]

	decodedConfig := factory.NewConfig()
	json.Unmarshal(data, decodedConfig)

	backend, err = factory.New(decodedConfig)
	if err != nil {
		return nil, err
	}

	stripperConfig := &stripper.Config{}
	json.Unmarshal(data, stripperConfig)
	if stripperConfig.BaseDn != nil || stripperConfig.PeopleRdn != nil || stripperConfig.UserRdnAttribute != nil {
		if stripperConfig.BaseDn == nil || stripperConfig.PeopleRdn == nil || stripperConfig.UserRdnAttribute == nil {
			stripperConfig = nil
			jww.WARN.Printf("Incomplete stripper config found in backend '%s' IGNORED", backend.Name())
		}
	} else {
		stripperConfig = nil
	}

	jww.INFO.Printf("Instantiated %s backend '%s'", factory.Name(), backend.Name())

	if stripperConfig != nil {
		backend = stripper.NewBackend(backend, stripperConfig)
		jww.INFO.Printf("Wrapping backend '%s' with stripper ('%s', '%s', '%s')", backend.Name(), *stripperConfig.UserRdnAttribute, *stripperConfig.PeopleRdn, *stripperConfig.BaseDn)
	}
	return backend, nil
}
