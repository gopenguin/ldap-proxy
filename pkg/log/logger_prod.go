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

func NewProdLogger(out io.Writer, flags int) Logger {
	return &productionLogger{
		logger: log.New(out, "", flags),
	}
}

type productionLogger struct {
	logger *log.Logger
}

var _ Logger = &productionLogger{}

func (pl *productionLogger) Print(v ...interface{}) {
	pl.logger.Print(v...)
}

func (pl *productionLogger) Println(v ...interface{}) {
	pl.logger.Println(v...)
}

func (pl *productionLogger) Printf(format string, v ...interface{}) {
	pl.logger.Printf(format, v...)
}

func (*productionLogger) Debug(v ...interface{}) {
}

func (*productionLogger) Debugln(v ...interface{}) {
}

func (*productionLogger) Debugf(format string, v ...interface{}) {
}
