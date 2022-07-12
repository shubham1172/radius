// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package bicep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/project-radius/radius/pkg/cli/tools"
)

// Official regex for semver
// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
const SemanticVersionRegex = `(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`

// Run rad-bicep with the given args and return the stdout. The stderr
// is not capture but instead redirected to that of the current process.
func runBicep(args ...string) (map[string]interface{}, error) {

	if installed, _ := IsBicepInstalled(); !installed {
		return nil, fmt.Errorf("rad-bicep not installed, run \"rad bicep download\" to install")
	}

	binPath, err := tools.GetLocalFilepath(radBicepEnvVar, binaryName)
	if err != nil {
		return nil, fmt.Errorf("failed to find rad-bicep: %w", err)
	}

	// runs 'rad-bicep'
	fullCmd := binPath + " " + strings.Join(args, " ")
	c := exec.Command(binPath, args...)
	c.Stderr = os.Stderr
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	err = c.Start()
	if err != nil {
		return nil, fmt.Errorf("failed executing %q: %w", fullCmd, err)
	}

	// asyncronously copy to our buffer, we don't really need to observe
	// errors here since it's copying into memory
	buf := bytes.Buffer{}
	go func() {
		_, _ = io.Copy(&buf, stdout)
	}()

	// Wait() will wait for us to finish draining stderr before returning the exit code
	err = c.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed executing %q: %w", fullCmd, err)
	}

	// read the content
	bytes, err := io.ReadAll(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read rad-bicep output: %w", err)
	}

	template := map[string]interface{}{}
	err = json.Unmarshal(bytes, &template)
	if err != nil {
		return nil, err
	}

	return template, err
}

// Build the provided `.bicep` file and returns the deployment template.
func Build(filePath string) (map[string]interface{}, error) {
	// rad-bicep is being told to output the template to stdout and we will capture it
	// rad-bicep will output compilation errors to stderr which will go to the user's console
	return runBicep("build", "--stdout", filePath)
}

// Return a Bicep version.
//
// In case we can't determine a version, output "unknown (<failure reason>)".
func Version() string {
	output, err := runBicep("--version")
	if err != nil {
		return fmt.Sprintf("unknown (%s)", err)
	}

	bytes, err := json.Marshal(output)
	if err != nil {
		return fmt.Sprintf("could not marshal json (%s)", err)
	}

	version := regexp.MustCompile(SemanticVersionRegex).FindString(string(bytes))
	if version == "" {
		return fmt.Sprintf("unknown (failed to parse bicep version from %q)", output)
	}
	return version
}
