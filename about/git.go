package about

import "regexp"

const (
	PSuite  = "gotest_psuite_test.go"
	PXSuite = "gotest_pxsuite_test.go"
)

var PSuiteRegex = regexp.MustCompile(`gotest_p(x)?suite_test\.go$`)

const (
	Application = "gotest"
	Repo        = "github.com/mvrahden/go-test"
)

var Version = "dev"
