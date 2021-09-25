//	Copyright 2021 Tord Kloster
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tordsk/go-swagger-diff/internal"
)

var ref string

var breakingCmd = &cobra.Command{
	Use:   "breaking <swagger>",
	Short: "will let you know if your api is breaking",
	Long: `This will compare the supplied spec to another branch.
Hints to which version to bump as well`,
	Example: "breaking swagger.yaml --debug",
	Args:    cobra.ExactArgs(1),

	Run: internal.Breaking(&ref, &debug),
}

func init() {
	breakingCmd.Flags().StringVarP(&ref, "ref", "r", "master", "ref to read the original spec from")
	rootCmd.AddCommand(breakingCmd)
}
