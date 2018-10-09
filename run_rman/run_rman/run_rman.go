package main

// Standard imports

// Local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"

import "github.com/daviesluke/run_rman/config"
import "github.com/daviesluke/run_rman/general"
import "github.com/daviesluke/run_rman/locker"
import "github.com/daviesluke/run_rman/resource"

// Local Variables

const (
	version string = "V2.0"
)

func main() {
	// Initialise some global variables
	setup.Initialize()

	// Initialize the logging 
	logger.Initialize(setup.LogDir, setup.LogFileName, setup.LogConfigFileName)

	logger.Infof("Process %s %s starting (PID %s) ...", setup.BaseName, version, setup.CurrentPID)

	// Trap signals to tidy up if received 
	utils.TrapSignal(general.Cleanup)

	// Validate the command line parameters
	general.ValidateFlags()

	// Check the command script provided
	config.SetRMANScript()

	// Read the config file 
	config.GetConfig(setup.ConfigFileName)

	// Check and set the environment
	general.SetEnvironment(setup.Database)

	// Reset logging to reflect the environment
	general.RenameLog()

	// Lock the process if supplied
	locker.LockProcess(general.LockName,setup.Database)

	// Set any resources supplied
	resource.GetResources(general.Resources)

	// Perform file removal, lock removal, resources cleanup needed
	general.Cleanup()

	logger.Info("Process complete")
}
