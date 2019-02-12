package config

// standard imports

import "bufio"
import "flag"
import "os"
import "path/filepath"
import "strings"

// local imports

import "../../logger"
import "../../utils"

// Global variables

// Initially set the defaults

var ConfigValues = map[string]string {
	"LogKeepTime"       : "14",
	"NLS_DATE_FORMAT"   : "DD_MON_YYYY HH24:MI:SS",
	"OraTabPath"        : "/etc/oratab:/var/opt/oracle/oratab",
	"RMANConfig"        : "",
	"CatalogConnection" : "",
	"TargetConnection"  : "/",
	"CheckLockMins"     : "5",
	"CheckResourceMins" : "5",
	"ParallelSlaves"    : "1",
	"ChannelDevice"     : "DISK",
	"FileFormat"        : "",
	"RMANIgnoreCodes"   : "",
	"EmailServer"       : "localhost:25",
}

var ConfigFileValues      map[string]string

var RMANScript        string
var RMANScriptBase    string

// Global Functions

func GetConfig ( configFileName string ) {
	logger.Infof("Reading configuration file %s ...", configFileName)

	// 
	// Initialize string map
	//
	ConfigFileValues = make(map[string]string)
	logger.Trace("Initialized config values map")

	//
	// Checking the config file 
	//

	logger.Debugf("Opening file %s ...", configFileName)
	configFile, err := os.Open(configFileName)
	if err != nil {
		logger.Infof("Unable to open file %s.  All defaults will be used.", configFileName)
	} else {
		logger.Debugf("Config file %s opened", configFileName)

		//
		// Close file at the end of the function
		//
		defer configFile.Close()
		logger.Tracef("Deferred closing of file %s at end of function", configFileName)

		configScanner := bufio.NewScanner(configFile)
		logger.Tracef("Set up scanner for config file. Entering loop ...")

		lineNo := 0

		logger.Tracef("Set up variable LineNo and set to %d", lineNo)

		for configScanner.Scan() {
			lineNo++
			logger.Tracef("Line number incremented to %d", lineNo)

			// Ignore blank lines and comments

			configLine := strings.TrimSpace(configScanner.Text())
			logger.Debugf("Trimmed config file line number %d contents - %s", lineNo, configLine)

			if configLine == "" || configLine[0] == '#' {
				logger.Trace("Comment or blank line - ignoring line")
				continue
			}

			variableTokens := strings.SplitN(configLine, "=", 2)
			logger.Tracef("Line split into %d tokens using = as delimiter", len(variableTokens))

			if len(variableTokens) == 0 {
				logger.Trace("No tokens - ignoring line")
				continue
			}

			if len(variableTokens) != 2 {
				logger.Errorf("Malformed variables in config file -> %s", configLine)
			}

			ConfigFileValues[strings.TrimSpace(variableTokens[0])]=strings.TrimSpace(variableTokens[1])
			logger.Tracef("Set map key %s", strings.TrimSpace(variableTokens[0]))


			//
			// If variable ends with Connection then careful with printing passwords
			//

			if utils.CheckRegEx(strings.TrimSpace(variableTokens[0]),".+Connection$") {
				logger.Infof("Set %s to %s", strings.TrimSpace(variableTokens[0]), utils.RemovePassword(ConfigFileValues[strings.TrimSpace(variableTokens[0])],true))
			} else {
				logger.Infof("Set %s to %s", strings.TrimSpace(variableTokens[0]), ConfigFileValues[strings.TrimSpace(variableTokens[0])])
			}


			//
			// If variable is all upper case then export it
			//

			//if utils.CheckRegEx(strings.TrimSpace(variableTokens[0]),"^[A-Z1-9_]+$") {
			//	logger.Tracef("String %s is all uppercase.  Exporting variable ...", strings.TrimSpace(variableTokens[0]))
			//
			//	os.Setenv(strings.TrimSpace(variableTokens[0]),strings.TrimSpace(variableTokens[1]))
			//	logger.Debugf("Exported variable %s", strings.TrimSpace(variableTokens[0]))
			//} else {
			//	logger.Tracef("String %s is NOT all uppercase", strings.TrimSpace(variableTokens[0]))
			//}
		}

		//
		// Print out map variables for debug
		//
		logger.Debug("Contents of config map - ")
		for configKey , configValue := range ConfigFileValues {
			logger.Debugf("Key : %s Value : %s", configKey, configValue)
		}
	}

	logger.Info("Process complete")
}

func SetRMANScript () {
	logger.Debug("Setting the RMAN script ...")

	programArgs := flag.Args()

	if len(programArgs) < 1 {
		logger.Errorf("Must provide one parameter. An RMAN script to run")
	}

	RMANScript = programArgs[0]
	logger.Tracef("RMAN script set to %s", RMANScript)

	var err error

	RMANScript, err = filepath.Abs(RMANScript)
	if err != nil {
		logger.Errorf("Unable to get absolute pathname for %s", programArgs[0])
	}
	logger.Tracef("Absolute name for RMAN script set to %s", RMANScript)

	// Check that file exists

	if _, err := os.Stat(RMANScript); err == nil {
		logger.Infof("RMAN script to run -> %s", RMANScript)
	} else {
		logger.Errorf("Unable to find RMAN script %s", RMANScript)
	}

	// Derive the base name for the script

	logger.Trace("Splitting RMAN file name using .")
	rmanBaseName := strings.SplitN(filepath.Base(RMANScript),".",2)
	
	RMANScriptBase = rmanBaseName[0]
	logger.Debugf("RMAN Script Base variable set to %s",RMANScriptBase)

	logger.Debug("Process complete")
}

func SetConfig ( database string , configName string ) {
	logger.Debugf("Checking and setting config entry %s for database %s ...", configName, database)

	//
	// If variable ends with Connection then careful with printing passwords
	//

	isConnection := utils.CheckRegEx(configName,".+Connection$")

	newConfigName := strings.Join( []string{ database, configName}, "_")

	logger.Tracef("New config name set to %s", newConfigName)
	
	if _, keyExists := ConfigFileValues[newConfigName]; keyExists {
		ConfigValues[configName] = ConfigFileValues[newConfigName]
		if isConnection {
			logger.Infof("Found config name %s in config file. Reset config name %s to %s", newConfigName, configName, utils.RemovePassword(ConfigValues[configName],false))
		} else {
			logger.Infof("Found config name %s in config file. Reset config name %s to %s", newConfigName, configName, ConfigValues[configName])
		}
	} else if _, keyExists := ConfigFileValues[configName]; keyExists {
		ConfigValues[configName] = ConfigFileValues[configName]
		if isConnection {
			logger.Infof("Found config name %s in config file. Reset config name %s to %s", configName, configName, utils.RemovePassword(ConfigValues[configName],false))
		} else {
			logger.Infof("Found config name %s in config file. Reset config name %s to %s", configName, configName, ConfigValues[configName])
		}
	} else {
		logger.Infof("No changes made from default name for %s - Value %s", configName, ConfigValues[configName])
	}

	logger.Debug("Process complete")
}

func SetAllConfig ( database string ) {
	logger.Info("Checking all config options ...")

	for configName, _ := range ConfigValues {
		SetConfig( database, configName)
	}

	logger.SetEmailServer(ConfigValues["EmailServer"])

	logger.Info("Process complete")
}
