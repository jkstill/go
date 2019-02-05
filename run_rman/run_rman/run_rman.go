/* 
Version History

2018-11-27	Version 2.1 Luke
            Amend output to stop the word "error" being printed for easier error checking by external scripts
			Amend e-mail subject header to include some meaningful information - database, script and completion time
*/

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
import "github.com/daviesluke/run_rman/oracle"
import "github.com/daviesluke/run_rman/oracle/rman"

// Local Variables

const (
	version string = "V2.1.1"
)

func main() {
	// Grab the start time
	logger.SetStartTime()

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

	// Check the connections
	oracle.CheckConnections()

	// Get RMAN config
	rman.CheckConfig()

	// Run RMAN command
	rman.RunScript()

	// Reset RMAN config
	rman.ResetConfig()

	// Perform file removal, lock removal, resources cleanup needed
	general.Cleanup()

	// Write the history file
	logger.WriteHistory("SUCCESS")

	logger.Info("Process complete")

	// Send the log
	logger.SendLog("SUCCESS")
}
