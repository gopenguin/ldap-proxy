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

package log

import (
	"log"
	"os"
)

var (
	DebugEnabled = false
	internal     = NewLogger()
)

func NewLogger() Logger {
	if !DebugEnabled {
		return NewProdLogger(os.Stdout, log.LstdFlags)
	} else {
		return NewDebugLogger(os.Stdout, log.LstdFlags)
	}
}

func Reinit() {
	internal = NewLogger()
}

func Print(v ...interface{}) {
	internal.Print(v...)
}
func Println(v ...interface{}) {
	internal.Println(v...)
}
func Printf(format string, v ...interface{}) {
	internal.Printf(format, v...)
}

func Debug(v ...interface{}) {
	internal.Debug(v...)
}
func Debugln(v ...interface{}) {
	internal.Debugln(v...)
}
func Debugf(format string, v ...interface{}) {
	internal.Debugf(format, v...)
}

type Logger interface {
	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})

	Debug(v ...interface{})
	Debugln(v ...interface{})
	Debugf(format string, v ...interface{})
}
