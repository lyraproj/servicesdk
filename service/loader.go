package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lyraproj/pcore/px"
)

var defaultWorkflowsPath []string

func init() {
	executable, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("failed to determine the path of the executable: %s", err.Error()))
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		panic(fmt.Errorf("failed to eval symlinks on the executable: %s %s", executable, err.Error()))
	}
	executableParentDir := filepath.Dir(filepath.Dir(executable))
	// Load workflows from:
	// - WORKING_DIR/workflows
	// - EXECUTABLE_DIR/../workflows (to support brew and running build\lyra irrespective of working dir)
	defaultWorkflowsPath = []string{".", executableParentDir}

	lyraExeDir := os.Getenv(`LYRA_EXEDIR`)
	if lyraExeDir != `` {
		lyraExeDir = filepath.Dir(lyraExeDir)
		if lyraExeDir != executableParentDir {
			defaultWorkflowsPath = append(defaultWorkflowsPath, lyraExeDir)
		}
	}
}

// New creates a new federated loader instance
func FederatedLoader(parentLoader px.Loader) px.Loader {
	var loaders []px.ModuleLoader
	for _, workflowsPathElement := range defaultWorkflowsPath {
		loaders = append(loaders, px.NewFileBasedLoader(parentLoader, workflowsPathElement, "", px.PuppetDataTypePath))
	}
	return px.NewDependencyLoader(loaders)
}
