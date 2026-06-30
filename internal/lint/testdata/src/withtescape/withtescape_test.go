package withtescape

import (
	"github.com/mvrahden/go-test/pkg/gotest"
)

type EscapeTestSuite struct{}

func (s *EscapeTestSuite) TestMethodEscape(t *gotest.T) {
	t.T().Errorf("msg")          // want `Errorf is available on gotest.T — unnecessary T escape`
	t.T().FailNow()              // want `FailNow is available on gotest.T — unnecessary T escape`
	t.T().Skip()                 // want `Skipf is available on gotest.T — unnecessary T escape`
	t.T().SkipNow()              // want `Skipf is available on gotest.T — unnecessary T escape`
	t.T().Skipf("reason")        // want `Skipf is available on gotest.T — unnecessary T escape`
	t.T().Setenv("KEY", "VALUE") // want `Setenv is available on gotest.T — unnecessary T escape`
	_ = t.T().TempDir()          // want `TempDir is available on gotest.T — unnecessary T escape`
}

func (s *EscapeTestSuite) TestAliasEscape(t *gotest.T) {
	tt := t.T()
	tt.Errorf("msg")      // want `Errorf is available on gotest.T — unnecessary T escape`
	tt.Skip()             // want `Skipf is available on gotest.T — unnecessary T escape`
	tt.SkipNow()          // want `Skipf is available on gotest.T — unnecessary T escape`
	gotest.True(tt, true) // want `pass gotest.T directly to True — unnecessary T escape`
}

func (s *EscapeTestSuite) TestAssertionEscape(t *gotest.T) {
	gotest.True(t.T(), true)  // want `pass gotest.T directly to True — unnecessary T escape`
	gotest.Equal(t.T(), 1, 2) // want `pass gotest.T directly to Equal — unnecessary T escape`
}

func (s *EscapeTestSuite) TestNoEscape(t *gotest.T) {
	t.Errorf("msg")
	gotest.True(t, true)
}

func (s *EscapeTestSuite) BeforeEach() {}
func (s *EscapeTestSuite) AfterEach()  {}

func standaloneEscape(t *gotest.T) { //nolint:unused
	t.T().Errorf("msg") // want `Errorf is available on gotest.T — unnecessary T escape`
}
