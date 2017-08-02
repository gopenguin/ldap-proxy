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

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/kolleroot/ldap-proxy/pkg/config"
	"github.com/kolleroot/ldap-proxy/pkg/memory"
	"github.com/kolleroot/ldap-proxy/pkg/postgres"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

// proxyCmd represents the proxy subcommand.
// It launches the ldap proxy with the provided json configuration
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start a proxy which delegates to the configured backends",
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			jww.ERROR.Fatal(err)
		}
		filename, err := cmd.Flags().GetString("filename")
		if err != nil {
			jww.ERROR.Fatal(err)
		}

		wd, err := os.Getwd()
		if err != nil {
			jww.ERROR.Fatal(err)
		}
		configFilePath := filepath.Join(wd, filename)

		runProxyFromConfigFile(port, configFilePath)
	},
}

func init() {
	RootCmd.AddCommand(proxyCmd)

	proxyCmd.Flags().IntP("port", "p", 10636, "The port to listen on for secure ldap communication")
	proxyCmd.Flags().StringP("filename", "f", "config.json", "The configuration file for the backends in json format")
}

func runProxyFromConfigFile(port int, filename string) {
	jww.INFO.Printf("Loading config from %s", filename)
	f, err := os.Open(filename)
	if err != nil {
		jww.FATAL.Fatal(err)
	}

	loader := config.NewLoader()

	loader.AddFactory(memory.NewFactory())
	loader.AddFactory(postgres.NewFactory())

	reader := bufio.NewReader(f)
	backends, err := loader.Load(reader)
	if err != nil {
		jww.FATAL.Fatal(err)
	}

	proxy := pkg.NewLdapProxy()
	proxy.AddBackend(backends...)
	proxy.ListenAndServe(fmt.Sprintf(":%d", port))
}
