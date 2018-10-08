package main

// Standard imports

// Local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"

import "github.com/daviesluke/run_rman/config"
import "github.com/daviesluke/run_rman/general"
import "github.com/daviesluke/run_rman/locker"

// Local Variables

const (
	version string = "V1.17"
)

func main() {
	setup.Initialize()

	logger.Initialize(setup.LogDir, setup.LogFileName, setup.LogConfigFileName)

	logger.Infof("Process %s %s starting (PID %s) ...", setup.BaseName, version, setup.CurrentPID)

	utils.TrapSignal(general.Cleanup)

	general.ValidateFlags()

	config.SetRMANScript()

	config.GetConfig(setup.ConfigFileName)

	general.SetEnvironment(setup.Database)

	general.RenameLog()

	locker.LockProcess(general.LockName,setup.Database)

	general.Cleanup()

	logger.Info("Process complete")
}
