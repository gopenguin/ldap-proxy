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
	"crypto/tls"
	"errors"
	"github.com/kolleroot/ldap-proxy/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samuel/go-ldap/ldap"
	"net"
)

var (
	errInvalidSessionType = errors.New("proxy: Invalid session type")
)

var (
	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "proxy",
		Name:      "requests_total",
		Help:      "The total number of requests",
	}, []string{"action"})

	backendActionDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "proxy",
		Name:      "backend_duration",
		Help:      "The time spent by the backend searching",
		Buckets:   []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1},
	}, []string{"action", "backend"})
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(backendActionDuration)
}

type LdapProxy struct {
	backends map[string]Backend

	server *ldap.Server

	context context.Context
}

type session struct {
	context context.Context
	cancle  context.CancelFunc
}

func NewLdapProxy() *LdapProxy {
	proxy := &LdapProxy{
		backends: make(map[string]Backend),
		server:   &ldap.Server{},

		context: context.Background(),
	}

	proxy.server.Backend = LogBackend(proxy)

	return proxy
}

func (ldapProxy *LdapProxy) AddBackend(backends ...Backend) {
	log.Printf("Adding %d backends", len(backends))
	for _, bkend := range backends {
		ldapProxy.backends[bkend.Name()] = bkend
	}
}

func (ldapProxy *LdapProxy) ListenAndServe(addr string) {
	log.Printf("Start listening on %s", addr)
	ldapProxy.server.Serve("tcp", addr)
}

func (ldapProxy *LdapProxy) ListenAndServeTLS(addr string, tlsConfig *tls.Config) {
	log.Printf("Start listening securely on %s", addr)
	ldapProxy.server.ServeTLS("tcp", addr, tlsConfig)
}

func (ldapProxy *LdapProxy) Connect(remoteAddr net.Addr) (ldap.Context, error) {
	requestsTotal.With(prometheus.Labels{"action": "connect"}).Inc()

	ctx, cancle := context.WithCancel(ldapProxy.context)

	return &session{
		context: ctx,
		cancle:  cancle,
	}, nil
}

func (ldapProxy *LdapProxy) Disconnect(ctx ldap.Context) {
	sess, ok := ctx.(*session)
	if !ok {
		return
	}

	sess.cancle()

	requestsTotal.With(prometheus.Labels{"action": "disconnect"}).Inc()
}

func (ldapProxy *LdapProxy) Bind(ctx ldap.Context, req *ldap.BindRequest) (*ldap.BindResponse, error) {
	log.Debugf("bind as %s", req.DN)

	sess, ok := ctx.(*session)
	if !ok {
		return nil, errInvalidSessionType
	}

	requestsTotal.With(prometheus.Labels{"action": "bind"}).Inc()

	res := &ldap.BindResponse{
		BaseResponse: ldap.BaseResponse{
			Code: ldap.ResultInvalidCredentials,
		},
	}

	sess.context = setDn(sess.context, "")

	for _, backend := range ldapProxy.backends {
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			backendActionDuration.With(prometheus.Labels{"action": "auth", "backend": backend.Name()}).Observe(v)
		}))
		authenticated := backend.Authenticate(sess.context, req.DN, string(req.Password))
		timer.ObserveDuration()

		if authenticated {
			sess.context = setDn(sess.context, req.DN)

			res.BaseResponse.Code = ldap.ResultSuccess
			res.MatchedDN = req.DN
			break
		}
	}

	return res, nil
}

func (ldapProxy *LdapProxy) Add(ctx ldap.Context, req *ldap.AddRequest) (*ldap.AddResponse, error) {
	requestsTotal.With(prometheus.Labels{"action": "add"}).Inc()

	return &ldap.AddResponse{
		BaseResponse: ldap.BaseResponse{
			Code: ldap.ResultUnwillingToPerform,
		},
	}, nil
}

func (ldapProxy *LdapProxy) Delete(ctx ldap.Context, req *ldap.DeleteRequest) (*ldap.DeleteResponse, error) {
	requestsTotal.With(prometheus.Labels{"action": "delete"}).Inc()

	return &ldap.DeleteResponse{
		BaseResponse: ldap.BaseResponse{
			Code: ldap.ResultUnwillingToPerform,
		},
	}, nil
}

func (ldapProxy *LdapProxy) ExtendedRequest(ctx ldap.Context, req *ldap.ExtendedRequest) (*ldap.ExtendedResponse, error) {
	requestsTotal.With(prometheus.Labels{"action": "extended"}).Inc()

	return &ldap.ExtendedResponse{
		BaseResponse: ldap.BaseResponse{
			Code: ldap.ResultUnwillingToPerform,
		},
	}, nil
}

func (ldapProxy *LdapProxy) Modify(ctx ldap.Context, req *ldap.ModifyRequest) (*ldap.ModifyResponse, error) {
	requestsTotal.With(prometheus.Labels{"action": "modify"}).Inc()

	return &ldap.ModifyResponse{
		BaseResponse: ldap.BaseResponse{
			Code: ldap.ResultUnwillingToPerform,
		},
	}, nil
}

func (ldapProxy *LdapProxy) ModifyDN(ctx ldap.Context, req *ldap.ModifyDNRequest) (*ldap.ModifyDNResponse, error) {
	requestsTotal.With(prometheus.Labels{"action": "modify_dn"}).Inc()

	return &ldap.ModifyDNResponse{
		BaseResponse: ldap.BaseResponse{
			Code: ldap.ResultUnwillingToPerform,
		},
	}, nil
}

func (ldapProxy *LdapProxy) PasswordModify(ctx ldap.Context, req *ldap.PasswordModifyRequest) ([]byte, error) {
	requestsTotal.With(prometheus.Labels{"action": "modify_password"}).Inc()

	return []byte{}, nil
}

func (ldapProxy *LdapProxy) Search(ctx ldap.Context, req *ldap.SearchRequest) (*ldap.SearchResponse, error) {
	sess, ok := ctx.(*session)
	if !ok {
		return nil, errInvalidSessionType
	}

	requestsTotal.With(prometheus.Labels{"action": "search"}).Inc()

	if getDn(sess.context) == "" {
		return &ldap.SearchResponse{
			BaseResponse: ldap.BaseResponse{
				Code: ldap.ResultInsufficientAccessRights,
			},
		}, nil
	}

	res := &ldap.SearchResponse{
		BaseResponse: ldap.BaseResponse{
			Code: ldap.ResultSuccess,
		},
	}

	var searchResults []*ldap.SearchResult

	for _, backend := range ldapProxy.backends {
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			backendActionDuration.With(prometheus.Labels{"action": "search", "backend": backend.Name()}).Observe(v)
		}))
		users, err := backend.GetUsers(sess.context, req.Filter)
		timer.ObserveDuration()
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			searchResult := ldap.SearchResult{
				DN:         user.DN,
				Attributes: map[string][][]byte{},
			}

			for key, values := range user.Attributes {
				convertedValues := [][]byte{}
				for _, value := range values {
					convertedValues = append(convertedValues, []byte(value))
				}
				searchResult.Attributes[key] = convertedValues
			}

			searchResults = append(searchResults, &searchResult)
		}
	}

	res.Results = searchResults

	return res, nil
}

func (ldapProxy *LdapProxy) Whoami(ctx ldap.Context) (string, error) {
	sess, ok := ctx.(*session)
	if !ok {
		return "", errInvalidSessionType
	}

	requestsTotal.With(prometheus.Labels{"action": "whoami"}).Inc()

	return getDn(sess.context), nil
}
