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
	"github.com/kolleroot/ldap-proxy/pkg/log"
	"github.com/samuel/go-ldap/ldap"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

type logBackendContextKey struct{}

var (
	contextKeyId = logBackendContextKey{}
)

var (
	sessionCounter int64
)

func setSessId(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyId, atomic.AddInt64(&sessionCounter, 1))
}

func getSessId(ctx context.Context) int64 {
	value := ctx.Value(contextKeyId)
	if value == nil {
		return -1
	} else {
		return value.(int64)
	}
}

func LogBackend(backend ldap.Backend) ldap.Backend {
	return &logBackend{
		backend: backend,
	}
}

type logBackend struct {
	backend ldap.Backend
}

var _ ldap.Backend = &logBackend{}

func (l *logBackend) logCtx(name string, ctx ldap.Context, start time.Time) {
	duration := time.Since(start)

	sess := ctx.(*session)
	if sess.dn != "" {
		log.Print(getSessId(sess.context), " ", rightPad(10, name), " ", sess.dn, " ", duration)
	} else {
		log.Print(getSessId(sess.context), " ", rightPad(10, name), " ", duration)
	}
}

func (l *logBackend) logCtxAuth(name string, ctx ldap.Context, authenticated bool, start time.Time) {
	duration := time.Since(start)

	sess := ctx.(*session)
	if sess.dn != "" {
		log.Print(getSessId(sess.context), " ", rightPad(10, name), " ", sess.dn, " ", authenticated, " ", duration)
	} else {
		log.Print(getSessId(sess.context), " ", rightPad(10, name), " ", authenticated, " ", duration)
	}
}

func (l *logBackend) logCtxFilter(name string, ctx ldap.Context, filter ldap.Filter, start time.Time) {
	duration := time.Since(start)

	sess := ctx.(*session)
	if sess.dn != "" {
		log.Print(getSessId(sess.context), " ", rightPad(10, name), " ", sess.dn, " ", filter, " ", duration)
	} else {
		log.Print(getSessId(sess.context), " ", rightPad(10, name), " ", filter, " ", duration)
	}
}

func (l *logBackend) Add(ctx ldap.Context, req *ldap.AddRequest) (*ldap.AddResponse, error) {
	defer l.logCtx("ADD", ctx, time.Now())

	return l.backend.Add(ctx, req)
}

func (l *logBackend) Bind(ctx ldap.Context, req *ldap.BindRequest) (*ldap.BindResponse, error) {
	start := time.Now()

	res, err := l.backend.Bind(ctx, req)

	l.logCtxAuth("BIND", ctx, err == nil && res.BaseResponse.Code == ldap.ResultSuccess, start)
	return res, err
}

func (l *logBackend) Connect(remoteAddr net.Addr) (ldap.Context, error) {
	start := time.Now()

	ctx, err := l.backend.Connect(remoteAddr)

	sess := ctx.(*session)
	sess.context = setSessId(sess.context)

	l.logCtx("CONNECT", ctx, start)
	return ctx, err
}

func (l *logBackend) Delete(ctx ldap.Context, req *ldap.DeleteRequest) (*ldap.DeleteResponse, error) {
	defer l.logCtx("DELETE", ctx, time.Now())

	return l.backend.Delete(ctx, req)
}

func (l *logBackend) Disconnect(ctx ldap.Context) {
	defer l.logCtx("DISCONNECT", ctx, time.Now())

	l.backend.Disconnect(ctx)
}

func (l *logBackend) ExtendedRequest(ctx ldap.Context, req *ldap.ExtendedRequest) (*ldap.ExtendedResponse, error) {
	defer l.logCtx("EXTENDED", ctx, time.Now())

	return l.backend.ExtendedRequest(ctx, req)
}

func (l *logBackend) Modify(ctx ldap.Context, req *ldap.ModifyRequest) (*ldap.ModifyResponse, error) {
	defer l.logCtx("MODIFY", ctx, time.Now())

	return l.backend.Modify(ctx, req)
}

func (l *logBackend) ModifyDN(ctx ldap.Context, req *ldap.ModifyDNRequest) (*ldap.ModifyDNResponse, error) {
	defer l.logCtx("MODIFYDN", ctx, time.Now())

	return l.backend.ModifyDN(ctx, req)
}

func (l *logBackend) PasswordModify(ctx ldap.Context, req *ldap.PasswordModifyRequest) ([]byte, error) {
	defer l.logCtx("PWMODIFY", ctx, time.Now())

	return l.backend.PasswordModify(ctx, req)
}

func (l *logBackend) Search(ctx ldap.Context, req *ldap.SearchRequest) (*ldap.SearchResponse, error) {
	defer l.logCtxFilter("SEARCH", ctx, req.Filter, time.Now())

	return l.backend.Search(ctx, req)
}

func (l *logBackend) Whoami(ctx ldap.Context) (string, error) {
	defer l.logCtx("WHOAMI", ctx, time.Now())

	return l.backend.Whoami(ctx)
}

func rightPad(length int, value string) string {
	return value + strings.Repeat(" ", length-len(value))
}
