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
	"sync/atomic"
)

type proxyContextKey int

const (
	contextKeyId = proxyContextKey(iota)
	contextKeyDn
)

var (
	sessionCounter int64
)

func setId(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyId, atomic.AddInt64(&sessionCounter, 1))
}

func getId(ctx context.Context) int64 {
	value := ctx.Value(contextKeyId)
	if value == nil {
		return -1
	} else {
		return value.(int64)
	}
}

func setDn(ctx context.Context, dn string) context.Context {
	return context.WithValue(ctx, contextKeyDn, dn)
}

func getDn(ctx context.Context) string {
	value := ctx.Value(contextKeyDn)
	if value == nil {
		return ""
	} else {
		return value.(string)
	}
}
