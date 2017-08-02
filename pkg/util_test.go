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
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPasswordHelper(t *testing.T) {
	Convey("Verify the password ''", t, func() {
		So(VerifyPassword(HashPassword("", 4), ""), ShouldBeTrue)
		So(VerifyPassword("$2y$04$cKpQ30fsvP7wBJ//mZzEB.tQaLOIvw5y0Jt4xpMaF6cVbqXkXltaq", ""), ShouldBeTrue)
	})

	Convey("Verify the password 'a'", t, func() {
		So(VerifyPassword(HashPassword("a", 4), "a"), ShouldBeTrue)
		So(VerifyPassword("$2a$04$wIvmqg9WXCUKrr/kI6AOgOeKR5gLTWAPfn8fqJVrIvA0r03oNOYb6", "a"), ShouldBeTrue)
	})

	Convey("Verify the password 'abcdefg'", t, func() {
		So(VerifyPassword(HashPassword("abcdefg", 4), "abcdefg"), ShouldBeTrue)
		So(VerifyPassword("$2a$04$h0PYJJ8cVWJuRW7OrLGGLuunLymVAhFZhotHM2Gz3nvOiJTGoZzWa", "abcdefg"), ShouldBeTrue)
	})

	Convey("Test wrong password", t, func() {
		So(VerifyPassword(HashPassword("a", 4), "b"), ShouldBeFalse)
		So(VerifyPassword(HashPassword("b", 4), "a"), ShouldBeFalse)
	})
}
