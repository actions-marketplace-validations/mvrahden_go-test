package protocol

const (
	EnvSharedStateFile    = "GOTEST_SHARED_STATE_FILE"
	EnvTeardownBudgetFile = "GOTEST_TEARDOWN_BUDGET_FILE"
	EnvUpdateSnapshots    = "GOTEST_UPDATE_SNAPSHOTS"
	EnvCI                 = "GOTEST_CI"
	EnvCacheDir           = "GOTEST_CACHE_DIR"
)

func BudgetFilePath(binaryPath string) string {
	return binaryPath + ".budget"
}
