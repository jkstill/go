package general

// Standard imports


// Local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/utils"

// Local functions

// Global functions

func Cleanup() {
	var callingFuncName string
	
	callingFuncName = utils.GetFunctionName()

	logger.Infof("%s - Running cleanup ...", callingFuncName)

	logger.Infof("%s - Process complete", callingFuncName)
}
