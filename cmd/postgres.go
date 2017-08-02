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
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/kolleroot/ldap-proxy/pkg/postgres"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

func init() {
	RootCmd.AddCommand(postgresCmd)
	postgresCmd.AddCommand(addCmd)
	addCmd.AddCommand(addUserCmd)
}

var postgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add some resource to the database",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var addUserCmd = &cobra.Command{
	Use:   "user [name]",
	Short: "Add a new user with a password",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			return
		}

		user := args[0]
		dbUrl, _ := cmd.Flags().GetString("dbUrl")

		fmt.Print("Password: ")
		password, err := gopass.GetPasswd()
		if err != nil {
			jww.FATAL.Fatal(err)
		}

		postgresAddUser(dbUrl, user, password)
	},
}

func postgresAddUser(dbUrl string, user string, password []byte) {
	backend, err := postgres.NewBackend(&postgres.Config{
		Url: dbUrl,
	})

	if err != nil {
		jww.FATAL.Fatal(err)
	}

	err = backend.CreateUser(user, string(password))
	if err != nil {
		jww.FATAL.Fatal(err)
	}
}
