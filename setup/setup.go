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

var DirDelimiter string

// Local functions

func setPID () {
	//
	// Get the current PID 
	//
	logger.Trace(1,"Getting current PID ...")

	CurrentPID = strconv.Itoa(os.Getpid())

	logger.Tracef(1,"Current PID set to %s", CurrentPID)
}

func setDirDelimiter () {
	//
	// Get current OS
	//
	logger.Tracef(1,"Current OS identified as %s", runtime.GOOS)

	if runtime.GOOS == "windows" {
		DirDelimiter="\\"
	} else {
		DirDelimiter="/"
	}

	logger.Tracef(1,"Directory delimiter set to %s", DirDelimiter)
}

func setBase () {
	var err error

	//
	// Get current process name and directory 
	//
	logger.Trace(1,"Getting executable name ...")

	BaseName, err = os.Executable()
	if err != nil {
		logger.Error("Unable to get executable name. Exiting with errors ...")
	}

	logger.Tracef(1,"Current Executable is set to %s",BaseName)

	BaseDir = path.Dir(BaseName)
	logger.Tracef(1,"Base directory set to %s",BaseDir)
	BaseDir = path.Dir(BaseDir)
	logger.Tracef(1,"Base directory set to %s",BaseDir)

	BaseName = path.Base(BaseName)
	logger.Tracef(1,"Base name set to %s",BaseName)
}

func setTmpDir () {
	//
	// Get the default temporary directory
	//
	logger.Trace(1,"Getting temp directory name ...")
	TmpDir = os.TempDir()
	logger.Tracef(1,"Temp directory set to %s", TmpDir)
}

func setLog () {
	//
	// Setting up logging names
	//
	LogDir = strings.Join([]string{BaseDir, "logs"}, DirDelimiter)

	logger.Tracef(1,"Default log directory set to %s", LogDir)
	
	LogFileName = strings.Join([]string{LogDir, DirDelimiter, BaseName, "_", CurrentPID, ".", LogSuffix},"")

	logger.Tracef(1,"Default Log file set to %s",LogFileName)
}

func setConfig () {
	//
	// Setting up config names
	//
	ConfigDir = strings.Join([]string{BaseDir, "config"}, DirDelimiter)

	logger.Tracef(1,"Config directory set to %s", ConfigDir)
	
	LogConfigFileName = strings.Join([]string{ConfigDir, DirDelimiter, BaseName, ".", LogSuffix, ConfigSuffix},"")

	logger.Tracef(1,"Log config file set to %s",LogConfigFileName)

	ConfigFileName = strings.Join([]string{ConfigDir, DirDelimiter, BaseName, ".", ConfigSuffix},"")

	logger.Tracef(1,"Default config file set to %s",ConfigFileName)
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
	logger.Tracef(1,"Set old config file name to %s", OldConfigFileName)

	ConfigFileName = configFile
	logger.Tracef(1,"Set new config file name to %s", ConfigFileName)

	logger.Info("Process complete")
}

func SwitchLogDir (logDir string) {
	logger.Info("Switching log directory to new value ...")

	OldLogDir = LogDir
	logger.Tracef(1,"Set old log directory name to %s", OldLogDir)

	LogDir = logDir
	logger.Tracef(1,"Set new log directory name to %s", LogDir)

	logFile := path.Base(LogFileName)
	logger.Tracef(1,"Set log file component to %s", logFile)
	
	SwitchLogFile(strings.Join([]string{LogDir, DirDelimiter, logFile},""))

	logger.Info("Process complete")
}

func SwitchLogFile (logFile string) {
	logger.Info("Switching log file to new value ...")

	OldLogFileName = LogFileName
        logger.Tracef(1,"Set old log file name to %s", OldLogFileName)

	LogFileName = logFile
	logger.Tracef(1,"Set new log file name to %s", LogFileName)

	logger.Info("Process complete")
}
