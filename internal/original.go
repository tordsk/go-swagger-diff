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
package internal

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/go-swagger/go-swagger/cmd/swagger/commands"
	"github.com/go-swagger/go-swagger/cmd/swagger/commands/diff"
	"github.com/spf13/cobra"

	"io/ioutil"
	"os"
)

var ErrContentsIdentical = errors.New("specs are identical, skipping")

func generateDiff(ref string, specPath string) (*os.File, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	fnsplit := strings.Split(specPath, "/")
	oldSpecFileName := fnsplit[len(fnsplit)-1]
	oldSpec, err := ioutil.TempFile(tmpDir, fmt.Sprintf("*-%s", oldSpecFileName))
	if err != nil {
		return nil, err
	}
	file, err := ioutil.TempFile(tmpDir, "swagger-diff.*.json")
	if err != nil {
		return nil, err
	}
	cmdString := fmt.Sprintf("git show %s:./%s", ref, specPath)
	args := strings.Split(cmdString, " ")
	out, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		stdErrStr := string(err.(*exec.ExitError).Stderr)
		if strings.Contains(stdErrStr, fmt.Sprintf("invalid object name '%s'", ref)) {
			return nil, fmt.Errorf("no swagger spec found with ref %s", ref)
		}
		return nil, fmt.Errorf("unable to find original swagger spec: %s", stdErrStr)
	}
	_, err = oldSpec.Write(out)
	if err != nil {
		return nil, err
	}
	oldSpec.Close()
	oldSpec, _ = os.Open(oldSpec.Name())
	oldChecksum := sha256.New()
	if _, err := io.Copy(oldChecksum, oldSpec); err != nil {
		log.Fatal(err)
	}
	specFile, err := os.Open(specPath)
	if err != nil {
		return nil, err
	}
	newCheckSum := sha256.New()
	if _, err := io.Copy(newCheckSum, specFile); err != nil {
		log.Fatal(err)
	}
	if bytes.Equal(newCheckSum.Sum(nil), oldChecksum.Sum(nil)) {
		return nil, ErrContentsIdentical
	}

	diffout := commands.DiffCommand{}
	diffout.Args.OldSpec = oldSpec.Name()
	diffout.Args.NewSpec = specPath
	diffout.Destination = file.Name()
	diffout.Format = "json"
	return file, diffout.Execute(nil)
}

func getFromMaster() {

}

func Breaking(ref *string, debug *bool) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !*debug {
			log.SetOutput(ioutil.Discard)
		}
		outputFile, err := generateDiff(*ref, args[0])
		log.SetOutput(os.Stderr)

		if err == ErrContentsIdentical {
			cmd.Println("spec has not changed")
			os.Exit(0)
		}
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		result, err := ioutil.ReadAll(outputFile)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		differences := diff.SpecDifferences{}
		err = json.Unmarshal(result, &differences)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}

		var breakingChanges, nonBreakingChanges int
		for _, d := range differences {
			switch d.Compatibility {
			case diff.Breaking:
				breakingChanges++
				cmd.Println("BREAKING: ", d.String())

			case diff.NonBreaking:
				nonBreakingChanges++
				cmd.Println("NON-BREAKING: ", d.String())
			}
		}
		if *debug {
			cmd.Println(string(result))
		}

		if breakingChanges > 0 {
			cmd.Printf("\nFound %d breaking changes, bump MAJOR version\n", breakingChanges)
			return
		}
		if nonBreakingChanges > 0 {
			cmd.Printf("Found %d backwards compatible changes, bump MINOR version\n", nonBreakingChanges)
			return
		}
		cmd.Println("only trivial changes found, bump PATCH version")
	}
}
