package withdirectcalls

import (
	"testing"

	"github.com/mvrahden/go-test/pkg/gotest"
)

type DirectCallTestSuite struct{}

func (s *DirectCallTestSuite) TestParallelDirect(t *gotest.T) {
	t.T().Parallel() // want `use SuiteConfig.Parallel instead — T.Parallel bypasses suite lifecycle coordination`
}

func (s *DirectCallTestSuite) TestParallelIndirect(t *gotest.T) {
	tt := t.T()
	tt.Parallel() // want `use SuiteConfig.Parallel instead — T.Parallel bypasses suite lifecycle coordination`
}

func (s *DirectCallTestSuite) TestRunDirect(t *gotest.T) {
	t.T().Run("sub", func(st *testing.T) {}) // want `use It or When instead — T.Run bypasses gotest wrapping`
}

func (s *DirectCallTestSuite) TestRunIndirect(t *gotest.T) {
	tt := t.T()
	tt.Run("sub", func(st *testing.T) {}) // want `use It or When instead — T.Run bypasses gotest wrapping`
}

// Parallel inside closure is OK — scoped to subtest
func (s *DirectCallTestSuite) TestParallelInClosure(t *gotest.T) {
	_ = func() {
		t.T().Parallel()
	}
}

// Run inside closure is OK — likely inside It/When callback
func (s *DirectCallTestSuite) TestRunInClosure(t *gotest.T) {
	_ = func() {
		t.T().Run("sub", func(st *testing.T) {})
	}
}

func (s *DirectCallTestSuite) TestClean(t *gotest.T) {
	gotest.True(t, true)
}

func (s *DirectCallTestSuite) BeforeEach() {}
func (s *DirectCallTestSuite) AfterEach()  {}
