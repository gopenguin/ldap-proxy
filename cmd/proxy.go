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
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"crypto/tls"
)

type proxyConfig struct {
	Port   int
	Config string

	ServerCert string
	ServerKey  string

	Prometheus     bool
	PrometheusAddr string
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

	proxyCmd.Flags().IntVarP(&c.Port, "port", "p", 10636, "port to listen on for secure ldap communication")
	proxyCmd.Flags().StringVar(&c.Config, "config", "config.json", "configuration file for the backends in json format")

	proxyCmd.Flags().StringVar(&c.ServerCert, "server-cert", "server.pem", "the server certificate")
	proxyCmd.Flags().StringVar(&c.ServerKey, "server-key", "server-key.pem", "the servers private key")

	proxyCmd.Flags().BoolVar(&c.Prometheus, "prometheus", false, "enable prometheus metrics")
	proxyCmd.Flags().StringVar(&c.PrometheusAddr, "prometheus-addr", ":8080", "port to serve the prometheus metrics on")

	return proxyCmd
}

func init() {
	RootCmd.AddCommand(proxyCmd())
}

func runProxyFromConfigFile(c *proxyConfig) {
	initPrometheus(c)

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

	tlsConfig := loadTlsConfig(c)

	proxy := pkg.NewLdapProxy()
	proxy.AddBackend(backends...)
	proxy.ListenAndServeTLS(fmt.Sprintf(":%d", c.Port), tlsConfig)
}

func loadTlsConfig(c *proxyConfig) *tls.Config {
	cer, err := tls.LoadX509KeyPair(c.ServerCert, c.ServerKey)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cer},
	}
}

func initPrometheus(c *proxyConfig) {
	if !c.Prometheus {
		if c.PrometheusAddr != ":8080" {
			log.Print("Prometheus wont be startet. Please also set the flag --prometheus")
		}

		return
	}

	http.Handle("/metrics", promhttp.Handler())

	log.Print("Starting prometheus server on ", c.PrometheusAddr)
	go http.ListenAndServe(c.PrometheusAddr, nil)
}
