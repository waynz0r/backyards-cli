// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/internal/platform/buildinfo"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

const (
	defaultVersionString = "unavailable"
	versionEndpoint      = "/version"
)

type versionCommand struct{}

type versionOptions struct {
	shortVersion      bool
	onlyClientVersion bool
}

func newVersionOptions() *versionOptions {
	return &versionOptions{
		shortVersion:      false,
		onlyClientVersion: false,
	}
}

func newVersionCommand(cli cli.CLI) *cobra.Command {
	c := &versionCommand{}
	options := newVersionOptions()

	cmd := &cobra.Command{
		Use:           "version",
		Short:         "Print the client and api version information",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			c.run(cli, options)
		},
	}

	cmd.PersistentFlags().BoolVar(&options.shortVersion, "short", options.shortVersion, "Print the version number(s) only, with no additional output")
	cmd.PersistentFlags().BoolVar(&options.onlyClientVersion, "client", options.onlyClientVersion, "Print the client version only")

	return cmd
}

func (c *versionCommand) run(cli cli.CLI, options *versionOptions) {
	clientVersion := GetRootCommand().Version
	if options.shortVersion {
		fmt.Println(clientVersion)
	} else {
		fmt.Printf("Client version: %s\n", clientVersion)
	}

	if options.onlyClientVersion {
		return
	}

	apiVersion := getAPIVersion(cli, versionEndpoint)
	if options.shortVersion {
		fmt.Println(apiVersion)
	} else {
		fmt.Printf("API version: %s\n", apiVersion)
	}
}

func getAPIVersion(cli cli.CLI, versionEndpoint string) string {
	pf, err := cli.GetPortforwardForIGW(0)
	if err != nil {
		return defaultVersionString
	}

	err = pf.Run()
	if err != nil {
		return defaultVersionString
	}

	resp, err := http.Get(pf.GetURL(versionEndpoint))
	if err != nil {
		return defaultVersionString
	}
	defer resp.Body.Close()

	var bi buildinfo.BuildInfo
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&bi)

	return bi.Version
}
