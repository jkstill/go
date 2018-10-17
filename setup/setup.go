package setup

// Standard imports

import "os"
import "path/filepath"
import "runtime"
import "strings"
import "strconv"

// Local imports

import "github.com/daviesluke/logger"

// Global Variables

const (
	LockSuffix        string = "lock"
	LogSuffix         string = "log"
	ConfigSuffix      string = "cfg"
	ResourceSuffix    string = "resources"
	UsedResSuffix     string = "used"
	ObtainedResSuffix string = "obtained"
)

// Directories

var BaseDir                  string
var TmpDir                   string
var LogDir                   string
var OldLogDir                string
var ConfigDir                string
var RMANScriptDir            string

// Files

var BaseName                 string
var LogFileName              string
var OldLogFileName           string
var LogConfigFileName        string
var ConfigFileName           string
var OldConfigFileName        string
var LockFileName             string
var ResourceBaseName         string
var ResourceFileName         string
var ResourceUsageFileName    string
var ResourceObtainedFileName string
var TmpFileName              string
var HistFileName             string

// Misc variables 

var CurrentPID               string

var DirDelimiter             string
var OratabDelimiter          string
var ExecutableSuffix         string

var LogMoved          bool   = true

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

func setDelimiter () {
	//
	// Get current OS
	//
	logger.Tracef("Current OS identified as %s", runtime.GOOS)

	if runtime.GOOS == "windows" {
		DirDelimiter="\\"
		OratabDelimiter=";"
		ExecutableSuffix=".exe"
	} else {
		DirDelimiter="/"
		OratabDelimiter=":"
		ExecutableSuffix=""
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

	BaseDir = filepath.Dir(BaseName)
	logger.Tracef("Interim Base directory set to %s",BaseDir)
	BaseDir = filepath.Dir(BaseDir)
	logger.Tracef("Base directory set to %s",BaseDir)

	BaseName = filepath.Base(BaseName)
	logger.Tracef("Base name set to %s",BaseName)

	// Now remove any extentions if they are there ( windows .exe for example)

	baseName := strings.SplitN(BaseName,".",2)
	logger.Trace("Split string using .")

	BaseName = baseName[0]
	logger.Tracef("Base name set to %s",BaseName)
}

func setTmpFile () {
	//
	// Set the default temp file
	//

	setTmpDir()

	logger.Trace("Getting temp file name ...")
	TmpFileName = strings.Join( []string{ BaseName, CurrentPID }, ".")
	TmpFileName = filepath.Join( TmpDir , TmpFileName )
	logger.Tracef("Temp file set to %s", TmpFileName)
	
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
	LogDir = filepath.Join(BaseDir, "logs")

	logger.Tracef("Default log directory set to %s", LogDir)
	
	LogFileName = strings.Join([]string{BaseName, "_", CurrentPID, ".", LogSuffix},"")
	LogFileName = filepath.Join(LogDir, LogFileName)

	logger.Tracef("Default Log file set to %s",LogFileName)
}

func setConfig () {
	//
	// Setting up config names
	//
	ConfigDir = filepath.Join(BaseDir, "config")

	logger.Tracef("Config directory set to %s", ConfigDir)
	
	LogConfigFileName = strings.Join([]string{BaseName, ".", LogSuffix, ConfigSuffix}, "")
	LogConfigFileName = filepath.Join(ConfigDir, LogConfigFileName)

	logger.Tracef("Log config file set to %s",LogConfigFileName)

	ConfigFileName = strings.Join([]string{BaseName, ConfigSuffix}, ".")
	ConfigFileName = filepath.Join(ConfigDir, ConfigFileName)

	logger.Tracef("Default config file set to %s",ConfigFileName)
}

func setLockFile () {
	//
	// Setting up generic lock name
	//
	LockFileName = strings.Join([]string{BaseName, LockSuffix}, ".")
	LockFileName = filepath.Join(LogDir, LockFileName)

	logger.Tracef("Default lock file set to %s",LockFileName)
}

func setResourceFile () {
	//
	// Setting up resource file names
	//

	ResourceBaseName         = strings.Join([]string{BaseName, ResourceSuffix}, ".")
	ResourceFileName         = filepath.Join(ConfigDir, ResourceBaseName)

	ResourceUsageFileName    = strings.Join([]string{ResourceFileName, UsedResSuffix}, ".")

	ResourceObtainedFileName = strings.Join([]string{ResourceFileName, ObtainedResSuffix, CurrentPID}, ".")
}

func setRMANDir () {
	//
	// Setting up RMAN directory
	//

	RMANScriptDir = filepath.Join(BaseDir, "rman_scripts")

	logger.Tracef("RMAN Script directory set to %s", RMANScriptDir)
}

func setHistFile () {
	//
	// Setting History File Name
	//

	HistFileName = strings.Join([]string{BaseName, "time_hist"}, ".")
	HistFileName = filepath.Join(LogDir, HistFileName)

	logger.Tracef("History file set to %s", HistFileName)
}

// Global Functions

func Initialize() {
	setPID()

	setDelimiter()

	setBase()

	setTmpFile()

	setLog()

	setConfig()

	setLockFile()

	setResourceFile()

	setRMANDir()

	setHistFile()
}

func SetConfigFile (configFile string) {
	logger.Infof("Switching config file to new value %s ...", configFile)

	OldConfigFileName = ConfigFileName
	logger.Tracef("Set old config file name to %s", OldConfigFileName)

	ConfigFileName = configFile
	logger.Tracef("Set new config file name to %s", ConfigFileName)

	logger.Debug("Process complete")
}

func SetLogDir (logDir string) {
	logger.Infof("Switching log directory to new value %s ...", logDir)

	OldLogDir = LogDir
	logger.Tracef("Set old log directory name to %s", OldLogDir)

	LogDir = logDir
	logger.Tracef("Set new log directory name to %s", LogDir)

	logFile := filepath.Base(LogFileName)
	logger.Tracef("Set log file component to %s", logFile)
	
	SetLogFileName(filepath.Join(LogDir, logFile ))

	logger.Debug("Process complete")
}

func SetLogFileName (logFile string) {
	logger.Infof("Switching log file to new value %s ...", logFile)

	if LogMoved {
		OldLogFileName = LogFileName
		logger.Tracef("Set old log file name to %s", OldLogFileName)
	} else {
		logger.Tracef("Old log file not yet moved")
	}

	LogFileName = logFile
	logger.Tracef("Set new log file name to %s", LogFileName)

	LogMoved = false;

	logger.Debug("Process complete")
}

func SetDatabase (database string) {
	logger.Infof("Setting database name %s ...", database)

	Database = database

	logger.Debugf("Database name set to %s", Database)
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

	logger.Debug("Process complete")
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

	logger.Debug("Process complete")
}

func SetErrorEmail (email string) {
	logger.Info("Setting e-mail for errors ...")

	ErrorEmails = strings.Split(email,";")
	logger.Trace("ErrorEmails string array set")

	for _ , emailAddress := range ErrorEmails {
		logger.Infof("Error E-mail set to %s", emailAddress)
	}

	logger.Debug("Process complete")
}

func SetLogMoved ( logMoved bool ) {
	logger.Debug("Registering log move ...")

	LogMoved = logMoved
	logger.Debugf("LogMoved set to %t", LogMoved)

	logger.Debug("Process complete")
}

func RenameLog( oldLogfileName, newLogFileName string) {
	logger.RenameLog( oldLogfileName, newLogFileName, LogConfigFileName)
}

func CopyFileToLog( title string, fileName string ) {
	logger.CopyFileToLog( title, fileName, LogConfigFileName)
}
