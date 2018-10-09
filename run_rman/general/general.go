package general

// Standard imports

import "flag"
import "os"
import "path/filepath"
import "strings"
import "strconv"
import "time"

// Local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"
import "github.com/daviesluke/run_rman/config"
import "github.com/daviesluke/run_rman/locker"


// local Variables

// Initialise long flags

var configFile = flag.String("config"     , "", "Config File Name")
var database   = flag.String("db"         , "", "Database name")
var errorEmail = flag.String("erroremail" , "", "E-mail list for errors")
var email      = flag.String("email"      , "", "E-mail list for success / error")
var lock       = flag.String("lock"       , "", "Lock name")
var logDir     = flag.String("log"        , "", "Directory for logs")
var resource   = flag.String("resource"   , "", "Resource name")

// Global Variables

var LockName          string

var SuccessEmails     []string
var ErrorEmails       []string

var Resources         map[string]int

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
			setup.SetConfigFile(*configFile)
			logger.Tracef("Config file now set to %s", setup.ConfigFileName)
		} else if flagParam.Name == "log" || flagParam.Name == "L" {
			logger.Infof("Switching log directory to new directory -> %s", *logDir)
			setup.SetLogDir(*logDir)
			logger.Tracef("Log file now set to %s", setup.LogFileName)
		} else if flagParam.Name == "erroremail" || flagParam.Name == "e" {
			logger.Info("Validating error e-mail addresses ...")
			if utils.CheckRegEx(*errorEmail, emailRegEx) {
				logger.Debugf("Error E-mail - %s validated", *errorEmail)
				SetErrorEmail(*errorEmail)
			} else {
				logger.Errorf("Invalid error e-mail address list - %s", *errorEmail)
			}
		} else if flagParam.Name == "email" || flagParam.Name == "E" {
			logger.Info("Validating e-mail addresses ...")
			if utils.CheckRegEx(*email, emailRegEx) {
				logger.Debugf("E-mail - %s validated", *email)
				SetEmail(*email)
			} else {
				logger.Errorf("Invalid e-mail address list - %s", *email)
			}
		} else if flagParam.Name == "resource" || flagParam.Name == "r" {
			logger.Info("Validating resources ...")
			resourceRegEx := "^([a-zA-Z0-9_]+=[0-9]+)+([:;,.][a-zA-Z0-9_]+=[0-9]+)*$"
			if utils.CheckRegEx(*resource, resourceRegEx) {
				logger.Debugf("Resources - %s - validated", *resource)
				SetResource(*resource)
			} else {
				logger.Errorf("Invalid resources - %s", *resource)
			}
		} else if flagParam.Name == "db" || flagParam.Name == "d" {
			setup.SetDatabase(*database)
		} else if flagParam.Name == "lock" || flagParam.Name == "l" {
			SetLock(*lock)
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

	// Sets the correct ORACLE_SID environment 

	os.Setenv("ORACLE_SID",setup.Database)
	logger.Tracef("Set environment ORACLE_SID to %s", setup.Database)
	os.Unsetenv("TWO_TASK")
	logger.Trace("Unset TWO_TASK")


	// See if we can find Oracle Home in the OraTabPath string

	var oracleHome string

	// Now we have the database name examine the OraTabPath to see if there are any overrides

	config.SetConfig(setup.Database, "OraTabPath")

	logger.Tracef("Looping around the OraTabPath %s", config.ConfigValues["OraTabPath"])

	for _, envFile := range strings.Split(config.ConfigValues["OraTabPath"],setup.OratabDelimiter) {

		logger.Debugf("Checking database in file %s", envFile)

		// Check file exists

		if _, err := os.Stat(envFile); err == nil {
			oracleHome = utils.LookupFile(envFile,setup.Database,1,2,setup.OratabDelimiter,1)

			if oracleHome != "" {
				logger.Debug("Found entry in file. Breaking loop ...")
				break
			}
		} else {
			logger.Trace("File not found. Ignoring ...")
		}
	}

	if oracleHome == "" {
		// Check to see if it is set in the environment

		oracleHome := os.Getenv("ORACLE_HOME")

		if oracleHome == "" {
			logger.Errorf("Unable to locate an Oracle Home.  Use the correct SID and environment file.")
		} else {
			logger.Debug("Using ORACLE_HOME already set in environment")
		}
	} else {
		// Now set the ORACLE_HOME

		logger.Trace("Setting the ORACLE_HOME env variable")
		os.Setenv("ORACLE_HOME",oracleHome)
	}

	// Now check it is a valid ORACLE_HOME containing both sqlplus and rman

	commandSQLPLUS := strings.Join( [] string{ "sqlplus", setup.ExecutableSuffix }, "" )
	commandRMAN    := strings.Join( [] string{ "rman"   , setup.ExecutableSuffix }, "" )

	commandSQLPLUS = filepath.Join(  oracleHome, "bin", commandSQLPLUS )
	commandRMAN    = filepath.Join(  oracleHome, "bin", commandRMAN )

	logger.Tracef("Checking for SQLPLUS executable - %s", commandSQLPLUS)
	if _, err := os.Stat(commandSQLPLUS); err != nil {
		logger.Errorf("ORACLE_HOME %s does not contain command %s", oracleHome, commandSQLPLUS);
	}

	logger.Tracef("Checking for RMAN executable - %s", commandRMAN)
	if _, err := os.Stat(commandRMAN); err != nil {
		logger.Errorf("ORACLE_HOME %s does not contain command %s", oracleHome, commandRMAN);
	}

	logger.Infof("ORACLE_HOME set to %s", oracleHome)

	logger.Info("Process complete")
}

func SetLock (lock string) {
	logger.Info("Setting lock name ...")

	LockName = lock

	logger.Infof("Lock name set to %s", LockName)
}

func SetResource (resource string) {
	logger.Info("Setting resources ...")

	//
	// May be multiple resources so need to split them up
	//

	Resources = make(map[string]int)
	logger.Trace("Initialized Resources map")

	logger.Debug("Checking for correct delimiter - can be . , ; or :")

	delimiterList     := []string { "." , "," , ";" , ":" }
	resourceDelimiter := ":"

	for _, delimiter := range delimiterList {
		logger.Tracef("Checking for delimiter %s", delimiter)

		if strings.Contains(resource,delimiter) {
			resourceDelimiter = delimiter
			logger.Debugf("Delimiter set to %s", delimiter)
			break
		}
	}

	for _, resourceList := range strings.Split(resource, resourceDelimiter) {
		logger.Tracef("Resource list element is %s. Splitting into name and value ...", resourceList)
		resourceElement := strings.SplitN(resourceList,"=",2)
		logger.Tracef("Resource Name = %s , Value %s", resourceElement[0], resourceElement[1])
		Resources[resourceElement[0]] , _  = strconv.Atoi(resourceElement[1])
		logger.Trace("Resource set")
	}

	logger.Trace("Resources set")

	for resourceName, resourceValue := range Resources {
		logger.Infof("Resource %s set to %d", resourceName, resourceValue)
	}

	logger.Info("Process complete")
}

func SetEmail (email string) {
	logger.Info("Setting e-mail for all outcomes ...")

	//
	// May be multiple e-mails so need to split them up
	//
	
	SuccessEmails = strings.Split(email,";")
	logger.Trace("SuccessEmails string array set")

	for _ , emailAddress := range SuccessEmails {
		logger.Infof("Success E-mail set to %s", emailAddress)
	}

	if len(ErrorEmails) == 0 {
		logger.Trace("Error Email not yet set")
		SetErrorEmail(email)
	} else {
		logger.Trace("Error E-mail already set")
	}

	logger.Info("Process complete")
}

func SetErrorEmail (email string) {
	logger.Info("Setting e-mail for errors ...")

	ErrorEmails = strings.Split(email,";")
	logger.Trace("ErrorEmails string array set")

	for _ , emailAddress := range ErrorEmails {
		logger.Infof("Error E-mail set to %s", emailAddress)
	}

	logger.Info("Process complete")
}

func RenameLog () {
	logger.Info("Renaming log ...")

	logger.Trace("Getting current date and time ...")

	t := time.Now()

	tFormat := t.Format("20060201150405")
	logger.Debugf("Time set to %s", tFormat)

	newLogFileName := strings.Join( []string{ setup.BaseName, "_", setup.Database, "_", config.RMANScriptBase, "_", tFormat, ".", setup.LogSuffix}, "" )
	newLogFileName = filepath.Join( setup.LogDir, newLogFileName )
	logger.Infof("Log file set to %s", newLogFileName)

	setup.SetLogFileName(newLogFileName)

	logger.RenameLog(setup.OldLogFileName, setup.LogFileName, setup.LogConfigFileName)

	setup.SetLogMoved(true)

	logger.Info("Process complete")
}

func Cleanup() {
	logger.Infof("Running cleanup ...")

	locker.RemoveLockEntry(setup.LockFileName,setup.CurrentPID)

	logger.Infof("Process complete")
}

