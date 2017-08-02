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
	"github.com/kolleroot/ldap-proxy/pkg"
	"github.com/kolleroot/ldap-proxy/pkg/config"
	"github.com/kolleroot/ldap-proxy/pkg/memory"
	"github.com/kolleroot/ldap-proxy/pkg/postgres"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"os"
	"github.com/kolleroot/ldap-proxy/pkg/stripper"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start a proxy which delegates to internal backends",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		filename, _ := cmd.Flags().GetString("filename")

		runProxyFromConfigFile(port, filename)
	},
}

func init() {
	RootCmd.AddCommand(proxyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// proxyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	proxyCmd.Flags().IntP("port", "p", 10636, "The port to listen on for secure ldap communication")
	proxyCmd.Flags().StringP("filename", "f", "", "The configuration file for the backends")
}

func runProxy(dbUrl string, port int, baseDn string, peopleRdn string, userRdn string) {
	jww.INFO.Printf("connecting to %s", dbUrl)

	stripperConfig := &stripper.Config{
		BaseDn:           &baseDn,
		PeopleRdn:        &peopleRdn,
		UserRdnAttribute: &userRdn,
	}
	postgresConfig := &postgres.Config{
		Url: dbUrl,
	}

	postgresBackend, err := postgres.NewBackend(postgresConfig)

	if err != nil {
		jww.FATAL.Fatal(err)
	}

	stripperBackend := stripper.NewBackend(postgresBackend, stripperConfig)

	proxy := pkg.NewLdapProxy()
	proxy.AddBackend(stripperBackend)

	proxy.ListenAndServe(":10636")

	/*
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

		<-ch
		close(ch)
		ldapProxy.Stop()
	*/
}

func runProxyFromConfigFile(port int, filename string) {
	if filename == "" {
		jww.FATAL.Fatal("no filename set")
	}

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
