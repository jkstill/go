package general

// Standard imports

import "flag"
import "os"
import "strings"

// Local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"
import "github.com/daviesluke/run_rman/config"

// local Variables

// Initialise long flags

var configFile = flag.String("config"     , "", "Config File Name")
var database   = flag.String("db"         , "", "Database name")
var errorEmail = flag.String("erroremail" , "", "E-mail list for errors")
var email      = flag.String("email"      , "", "E-mail list for success / error")
var lock       = flag.String("lock"       , "", "Lock name")
var logDir     = flag.String("log"        , "", "Directory for logs")
var resource   = flag.String("resource"   , "", "Resource name")


// Local functions

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

// Global functions

func ValidateFlags () {
	logger.Info("Validating command line arguments ...")

	flag.Parse()

	visitor := func(flagParam *flag.Flag) {
		logger.Infof("Parameter %s set to %s", flagParam.Usage, flagParam.Value)

		// Validate and do appropriate changes for each option

		emailRegEx := "^(([a-zA-Z0-9_\\-\\.]+)@([a-zA-Z0-9_\\-\\.]+)\\.([a-zA-Z]{2,5}){1,25})+([;.](([a-zA-Z0-9_\\-\\.]+)@([a-zA-Z0-9_\\-\\.]+)\\.([a-zA-Z]{2,5}){1,25})+)*$"

		if flagParam.Name == "config" || flagParam.Name == "c" {
			logger.Infof("Config file switched from default to named file -> %s", *configFile)
			setup.SwitchConfigFile(*configFile)
			logger.Tracef("Config file now set to %s", setup.ConfigFileName)
		} else if flagParam.Name == "log" || flagParam.Name == "l" {
			logger.Infof("Switching log directory to new directory -> %s", *logDir)
			setup.SwitchLogDir(*logDir)
			//logger.SwitchLog()
			logger.Tracef("Log file now set to %s", setup.LogFileName)
		} else if flagParam.Name == "erroremail" || flagParam.Name == "e" {
			logger.Info("Validating error e-mail addresses ...")
			if utils.CheckRegEx(*errorEmail, emailRegEx) {
				logger.Debugf("Error E-mail - %s validated", *errorEmail)
				setup.SetErrorEmail(*errorEmail)
			} else {
				logger.Errorf("Invalid error e-mail address list - %s", *errorEmail)
			}
		} else if flagParam.Name == "email" || flagParam.Name == "E" {
			logger.Info("Validating e-mail addresses ...")
			if utils.CheckRegEx(*email, emailRegEx) {
				logger.Debugf("E-mail - %s validated", *email)
				setup.SetEmail(*email)
			} else {
				logger.Errorf("Invalid e-mail address list - %s", *email)
			}
		} else if flagParam.Name == "resource" || flagParam.Name == "r" {
			logger.Info("Validating resources ...")
			resourceRegEx := "^([a-zA-Z0-9_]+=[0-9]+)+([:;,.][a-zA-Z0-9_]+=[0-9]+)*$"
			if utils.CheckRegEx(*resource, resourceRegEx) {
				logger.Debugf("Resources - %s - validated", *resource)
				setup.SetResource(*resource)
			} else {
				logger.Errorf("Invalid resources - %s", *resource)
			}
		} else if flagParam.Name == "db" || flagParam.Name == "d" {
			setup.SetDatabase(*database)
		}
	}

	flag.Visit(visitor)

	logger.Info("Process complete")
}

func SetEnvironment ( database string ) {
	logger.Infof("Setting database environment ...")

	// Checking for database name 

	if database == "" {
		// Check ORACLE_SID in the environment
		
		logger.Trace("Getting ORACLE_SID from environment")
		setup.Database = os.Getenv("ORACLE_SID") 

		if  setup.Database == "" {
			// Check TWO_TASK to see if that is set (may have a env entry for that)

			logger.Trace("Getting TWO_TASK from environment")
			setup.Database = os.Getenv("TWO_TASK")

			if setup.Database == "" {
				//
				// Last chance is to check the TargetConnection in config to see if that has any entries in the OraTabPath
				//

				// First check for anything specified in the file 
			
				config.SetConfig("", "TargetConnection")

				if strings.Contains(config.ConfigValues["TargetConnection"],"@") {
					logger.Trace("Getting database name from target connection in config file")

					targetBreakdown := strings.SplitN(config.ConfigValues["TargetConnection"],"@",2)

					setup.Database = targetBreakdown[1]

					logger.Debug("Database set from TargetConnection")
				} else {
					logger.Errorf("Unable to find a database name.  Set in the command line option -d | -db.")
				}
			} else {
				logger.Debug("Database set from TWO_TASK")
			}
		} else {
			logger.Debug("Database set from ORACLE_SID")
		}
	} else {
		setup.Database = database
		logger.Debug("Database set from parameter")
	}

	logger.Infof("Database set to %s", setup.Database)

	// Now we have the database name examine the OraTabPath to see if there are any overrides

	config.SetConfig(setup.Database, "OraTabPath")

	// Sets the correct ORACLE_SID environment 

	os.Setenv("ORACLE_SID",setup.Database)
	logger.Tracef("Set environment ORACLE_SID to %s", setup.Database)
	os.Unsetenv("TWO_TASK")
	logger.Trace("Unset TWO_TASK")

	// Now check the environment to set the ORACLE_HOME

	oracleHome := os.Getenv("ORACLE_HOME")

	if oracleHome == "" {
		// See if we can find Oracle Home in the envPath string

		logger.Tracef("Looping around the OraTabPath %s", config.ConfigValues["OraTabPath"])

		for _, envFile := range strings.Split(config.ConfigValues["OraTabPath"],":") {

			logger.Debugf("Checking database in file %s", envFile)

			// Check file exists

			if _, err := os.Stat(envFile); err == nil {
				oracleHome = utils.LookupFile(envFile,setup.Database,1,2,":")

				if oracleHome != "" {
					logger.Debug("Found entry in file. Breaking loop ...")
					break
				}
			} else {
				logger.Trace("File not found. Ignoring ...")
			}
		}

		if oracleHome == "" {
			logger.Errorf("Unable to locate an Oracle Home.  Use the correct SID and environment file.")
		} else {
			// Now set the ORACLE_HOME

			logger.Trace("Setting the ORACLE_HOME env variable")
			os.Setenv("ORACLE_HOME",oracleHome)
		}
	} else {
		logger.Debug("ORACLE_HOME is already set")
	}

	// Now check it is a valid ORACLE_HOME containing both sqlplus and rman

	commandSQLPLUS := strings.Join( []string{ oracleHome, "bin", "sqlplus"}, setup.DirDelimiter )
	commandRMAN    := strings.Join( []string{ oracleHome, "bin", "rman"}, setup.DirDelimiter )

	logger.Trace("Checking for SQLPLUS executable")
	if _, err := os.Stat(commandSQLPLUS); err != nil {
		logger.Errorf("ORACLE_HOME %s does not contain command %s", oracleHome, commandSQLPLUS);
	}

	logger.Trace("Checking for RMAN executable")
	if _, err := os.Stat(commandRMAN); err != nil {
		logger.Errorf("ORACLE_HOME %s does not contain command %s", oracleHome, commandRMAN);
	}

	logger.Infof("ORACLE_HOME set to %s", oracleHome)

	logger.Info("Process complete")
}

func Cleanup() {
	logger.Infof("Running cleanup ...")

	logger.Infof("Process complete")
}

