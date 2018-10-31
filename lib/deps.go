package lib

import (
	"os/exec"
	"strings"
)

func ListDepsOfCurrentPackage() ([]string, error) {
	raw := `go list -f '{{join .Imports "\n"}}' | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}'`
	cmd := exec.Command("bash", "-c", raw)
	outputs, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	res := strings.Trim(string(outputs), "\n")
	return strings.Split(res, "\n"), nil
}
