package config

// standard imports

import "bufio"
import "os"
import "strings"

// local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"

// Global variables

const ( 
	DEFAULT_LogKeepTime       string = "14"
	DEFAULT_NLS_DATE_FORMAT   string = "DD_MON_YYYY HH24:MI:SS"
	DEFAULT_OraEnvPath        string = "/etc/oratab:/var/opt/oracle/oratab"
	DEFAULT_CatalogConnection string = ""
	DEFAULT_TargetConnection  string = "/"
	DEFAULT_CheckLockMins     string = "5"
	DEFAULT_CheckResourceMins string = "5"
	DEFAULT_ParallelSlaves    string = "1"
	DEFAULT_ChannelDevice     string = "DISK"
	DEFAULT_FileFormat        string = ""
	DEFAULT_RMANIgnoreErrors  string = ""
	DEFAULT_EmailServer       string = "localhost:25"
)


var ConfigValues map[string]string


// Local functions



// Global Functions

func GetConfig() {
	var callingFuncName string

	callingFuncName = utils.GetFunctionName()

	logger.Infof("%s - Reading configuration file ...", callingFuncName)

	// 
	// Initialize string map
	//
	ConfigValues = make(map[string]string)
	logger.Trace(1,"Initialized config values map")

	//
	// Checking the config file 
	//

	logger.Debugf("Opening file %s ...", setup.ConfigFileName)
	configFile, err := os.Open(setup.ConfigFileName)
	if err != nil {
		logger.Infof("%s\t- Unable to open file %s.  All defaults will be used.", callingFuncName, setup.ConfigFileName)
	} else {
		logger.Debugf("Config file %s opened", setup.ConfigFileName)

		//
		// Close file at the end of the function
		//
		defer configFile.Close()
		logger.Tracef(1, "Deferred closing of file %s at end of function", setup.ConfigFileName)

		logger.Infof("%s\t- Processing config file %s ...", callingFuncName, setup.ConfigFileName)

		configScanner := bufio.NewScanner(configFile)
		logger.Tracef(1,"Set up scanner for config file. Entering loop ...")

		lineNo := 0

		logger.Tracef(1,"Set up variable LineNo and set to %d", lineNo)

		for configScanner.Scan() {
			lineNo++
			logger.Tracef(1,"Line number incremented to %d", lineNo)

			// Ignore blank lines and comments

			configLine := strings.TrimSpace(configScanner.Text())
			logger.Debugf("Trimmed config file line number %d contents - %s", lineNo, configLine)

			if configLine == "" || configLine[0] == '#' {
				logger.Trace(1,"Comment or blank line - ignoring line")
				continue
			}

			variableTokens := strings.SplitN(configLine, "=", 2)
			logger.Tracef(1,"Line split into %d tokens using = as delimiter", len(variableTokens))

			if len(variableTokens) == 0 {
				logger.Trace(1,"No tokens - ignoring line")
				continue
			}

			if len(variableTokens) != 2 {
				logger.Errorf("%s\t- Malformed variables in config file -> %s", callingFuncName, configLine)
			}

			ConfigValues[strings.TrimSpace(variableTokens[0])]=strings.TrimSpace(variableTokens[1])
			logger.Tracef(1,"Set map key %s", strings.TrimSpace(variableTokens[0]))


			//
			// If variable ends with Connection then careful with printing passwords
			//

			if utils.CheckRegEx(strings.TrimSpace(variableTokens[0]),".+Connection$") {
				logger.Infof("%s\t- Set %s to %s", callingFuncName, strings.TrimSpace(variableTokens[0]), utils.RemovePassword(ConfigValues[strings.TrimSpace(variableTokens[0])]))
			} else {
				logger.Infof("%s\t- Set %s to %s", callingFuncName, strings.TrimSpace(variableTokens[0]), ConfigValues[strings.TrimSpace(variableTokens[0])])
			}


			//
			// If variable is all upper case then export it
			//

			//if utils.CheckRegEx(strings.TrimSpace(variableTokens[0]),"^[A-Z1-9_]+$") {
			//	logger.Tracef(1,"String %s is all uppercase.  Exporting variable ...", strings.TrimSpace(variableTokens[0]))
			//
			//	os.Setenv(strings.TrimSpace(variableTokens[0]),strings.TrimSpace(variableTokens[1]))
			//	logger.Debugf("Exported variable %s", strings.TrimSpace(variableTokens[0]))
			//} else {
			//	logger.Tracef(1,"String %s is NOT all uppercase", strings.TrimSpace(variableTokens[0]))
			//}
		}

		//
		// Print out map variables for debug
		//
		logger.Debug("Contents of config map - ")
		for configKey , configValue := range ConfigValues {
			logger.Debugf("Key : %s Value : %s", configKey, configValue)
		}
	}

	logger.Infof("%s - Process complete", callingFuncName)
}
