package setup

// Standard imports

import "os"
import "path"
import "runtime"
import "strings"
import "strconv"

// Local imports

import "github.com/daviesluke/logger"

// Global Variables

const (
	LogSuffix    string = "log"
	ConfigSuffix string = "cfg"
)

var BaseDir           string
var TmpDir            string
var LogDir            string
var OldLogDir         string
var ConfigDir         string

var BaseName          string
var LogFileName       string
var OldLogFileName    string
var LogConfigFileName string
var ConfigFileName    string
var OldConfigFileName string

var CurrentPID        string

var DirDelimiter      string

var Database          string

var SuccessEmails     []string
var ErrorEmails       []string

var Resources         map[string]int

// Local functions

func setPID () {
	//
	// Get the current PID 
	//
	logger.Trace("Getting current PID ...")

	CurrentPID = strconv.Itoa(os.Getpid())

	logger.Tracef("Current PID set to %s", CurrentPID)
}

func setDirDelimiter () {
	//
	// Get current OS
	//
	logger.Tracef("Current OS identified as %s", runtime.GOOS)

	if runtime.GOOS == "windows" {
		DirDelimiter="\\"
	} else {
		DirDelimiter="/"
	}

	logger.Tracef("Directory delimiter set to %s", DirDelimiter)
}

func setBase () {
	var err error

	//
	// Get current process name and directory 
	//
	logger.Trace("Getting executable name ...")

	BaseName, err = os.Executable()
	if err != nil {
		logger.Error("Unable to get executable name. Exiting with errors ...")
	}

	logger.Tracef("Current Executable is set to %s",BaseName)

	BaseDir = path.Dir(BaseName)
	logger.Tracef("Base directory set to %s",BaseDir)
	BaseDir = path.Dir(BaseDir)
	logger.Tracef("Base directory set to %s",BaseDir)

	BaseName = path.Base(BaseName)
	logger.Tracef("Base name set to %s",BaseName)
}

func setTmpDir () {
	//
	// Get the default temporary directory
	//
	logger.Trace("Getting temp directory name ...")
	TmpDir = os.TempDir()
	logger.Tracef("Temp directory set to %s", TmpDir)
}

func setLog () {
	//
	// Setting up logging names
	//
	LogDir = strings.Join([]string{BaseDir, "logs"}, DirDelimiter)

	logger.Tracef("Default log directory set to %s", LogDir)
	
	LogFileName = strings.Join([]string{LogDir, DirDelimiter, BaseName, "_", CurrentPID, ".", LogSuffix},"")

	logger.Tracef("Default Log file set to %s",LogFileName)
}

func setConfig () {
	//
	// Setting up config names
	//
	ConfigDir = strings.Join([]string{BaseDir, "config"}, DirDelimiter)

	logger.Tracef("Config directory set to %s", ConfigDir)
	
	LogConfigFileName = strings.Join([]string{ConfigDir, DirDelimiter, BaseName, ".", LogSuffix, ConfigSuffix},"")

	logger.Tracef("Log config file set to %s",LogConfigFileName)

	ConfigFileName = strings.Join([]string{ConfigDir, DirDelimiter, BaseName, ".", ConfigSuffix},"")

	logger.Tracef("Default config file set to %s",ConfigFileName)
}



// Global Functions

func Initialize() {
	setPID()

	setDirDelimiter()

	setBase()

	setTmpDir()

	setLog()

	setConfig()
}

func SwitchConfigFile (configFile string) {
	logger.Info("Switching config file to new value ...")

	OldConfigFileName = ConfigFileName
	logger.Tracef("Set old config file name to %s", OldConfigFileName)

	ConfigFileName = configFile
	logger.Tracef("Set new config file name to %s", ConfigFileName)

	logger.Info("Process complete")
}

func SwitchLogDir (logDir string) {
	logger.Info("Switching log directory to new value ...")

	OldLogDir = LogDir
	logger.Tracef("Set old log directory name to %s", OldLogDir)

	LogDir = logDir
	logger.Tracef("Set new log directory name to %s", LogDir)

	logFile := path.Base(LogFileName)
	logger.Tracef("Set log file component to %s", logFile)
	
	SwitchLogFile(strings.Join([]string{LogDir, logFile}, DirDelimiter))

	logger.Info("Process complete")
}

func SwitchLogFile (logFile string) {
	logger.Info("Switching log file to new value ...")

	OldLogFileName = LogFileName
        logger.Tracef("Set old log file name to %s", OldLogFileName)

	LogFileName = logFile
	logger.Tracef("Set new log file name to %s", LogFileName)

	logger.Info("Process complete")
}

func SetDatabase (database string) {
	logger.Info("Setting database name ...")

	Database = database

	logger.Infof("Database name set to %s", Database)
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
