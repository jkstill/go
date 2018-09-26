package main

// Standard imports

import "flag"

// Local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"

import "github.com/daviesluke/run_rman/config"
import "github.com/daviesluke/run_rman/general"

// Local Variables

const (
	version string = "V1.17"
)

// Flags

var configFile = flag.String("config"     , "", "Config File Name")
var database   = flag.String("db"         , "", "Database name")
var errorEmail = flag.String("erroremail" , "", "E-mail list for errors")
var email      = flag.String("email"      , "", "E-mail list for success / error")
var lock       = flag.String("lock"       , "", "Lock name")
var logDir     = flag.String("log"        , "", "Directory for logs")
var resource   = flag.String("resource"   , "", "Resource name")


func init() {
	//
	// Setting up short flags
	//

	flag.StringVar(configFile, "c", "", "Config File Name")
	flag.StringVar(database  , "d", "", "Database name")
	flag.StringVar(errorEmail, "e", "", "E-mail list for errors")
	flag.StringVar(email     , "E", "", "E-mail List for success / error")
	flag.StringVar(lock      , "l", "", "Lock name")
	flag.StringVar(logDir    , "L", "", "Alternative Log directory")
	flag.StringVar(resource  , "r", "", "Resource name")
}

func listFlags() {
	var callingFuncName string
	
	callingFuncName = utils.GetFunctionName()

	logger.Infof("%s - Listing command line arguments ...", callingFuncName)

	visitor := func(flagParam *flag.Flag) {
		logger.Infof("%s\t- Parameter %s set to %s", callingFuncName, flagParam.Usage, flagParam.Value)
	}

	flag.Visit(visitor)

	logger.Infof("%s - Process complete", callingFuncName)
}

func main() {
	var callingFuncName string

	setup.Initialize()

	logger.Initialize(setup.LogDir, setup.LogFileName, setup.LogConfigFileName)

	callingFuncName = utils.GetFunctionName()

	logger.Infof("%s - Process %s %s starting (PID %s) ...", callingFuncName, setup.BaseName, version, setup.CurrentPID)

	utils.TrapSignal(general.Cleanup)

	flag.Parse()

	listFlags()

	config.GetConfig()

	logger.Infof("%s - Process complete", callingFuncName)
}
