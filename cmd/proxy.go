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
	"github.com/kolleroot/ldap-proxy/pkg/log"
)

type proxyConfig struct {
	Port   int
	Config string
}

// proxyCmd represents the proxy subcommand.
// It launches the ldap proxy with the provided json configuration
func proxyCmd() *cobra.Command {
	c := &proxyConfig{}

	proxyCmd := &cobra.Command{
		Use:   "proxy",
		Short: "Start a proxy which delegates to the configured backends",
		Run: func(cmd *cobra.Command, args []string) {
			wd, err := os.Getwd()
			if err != nil {
				jww.ERROR.Fatal(err)
			}
			c.Config = filepath.Join(wd, c.Config)

			runProxyFromConfigFile(c)
		},
	}

	proxyCmd.Flags().IntVarP(&c.Port, "port", "p", 10636, "The Port to listen on for secure ldap communication")
	proxyCmd.Flags().StringVar(&c.Config, "config", "config.json", "The configuration file for the backends in json format")

	return proxyCmd
}

func init() {
	RootCmd.AddCommand(proxyCmd())
}

func runProxyFromConfigFile(c *proxyConfig) {
	log.Printf("Loading Config from %s", c.Config)
	f, err := os.Open(c.Config)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	loader := config.NewLoader()

	loader.AddFactory(memory.NewFactory())
	loader.AddFactory(postgres.NewFactory())

	reader := bufio.NewReader(f)
	backends, err := loader.Load(reader)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	proxy := pkg.NewLdapProxy()
	proxy.AddBackend(backends...)
	proxy.ListenAndServe(fmt.Sprintf(":%d", c.Port))
}
