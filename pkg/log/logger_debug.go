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
	"io"
	"log"
)

func NewDebugLogger(out io.Writer, flags int) Logger {
	return &debugLogger{
		logger:      log.New(out, "", flags),
		debugLogger: log.New(out, "DEBUG", flags),
	}
}

type debugLogger struct {
	logger      *log.Logger
	debugLogger *log.Logger
}

var _ Logger = &debugLogger{}

func (dl *debugLogger) Print(v ...interface{}) {
	dl.logger.Print(v...)
}

func (dl *debugLogger) Println(v ...interface{}) {
	dl.logger.Println(v...)
}

func (dl *debugLogger) Printf(format string, v ...interface{}) {
	dl.logger.Printf(format, v...)
}

func (dl *debugLogger) Debug(v ...interface{}) {
	dl.debugLogger.Print(v...)
}

func (dl *debugLogger) Debugln(v ...interface{}) {
	dl.debugLogger.Println(v...)
}

func (dl *debugLogger) Debugf(format string, v ...interface{}) {
	dl.debugLogger.Printf(format, v...)
}
