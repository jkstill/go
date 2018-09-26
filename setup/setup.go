package setup

// standard imports

import "os"
import "path"
import "runtime"
import "strings"
import "strconv"

// local imports

import "github.com/daviesluke/logger"

// Global Variables

const (
	LogSuffix    string = "log"
	ConfigSuffix string = "cfg"
)

var BaseDir           string
var TmpDir            string
var LogDir            string
var ConfigDir         string

var BaseName          string
var LogFileName       string
var LogConfigFileName string
var ConfigFileName    string

var CurrentPID        string

var DirDelimiter string

// Functions

func Initialize() {
	var err error

	//
	// Get the current PID 
	//
	logger.Trace(1,"Getting current PID ...")

	CurrentPID = strconv.Itoa(os.Getpid())

	logger.Tracef(1,"Current PID set to %s", CurrentPID)


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


	//
	// Get the default temporary directory
	//
	logger.Trace(1,"Getting temp directory name ...")
	TmpDir = os.TempDir()
	logger.Tracef(1,"Temp directory set to %s", TmpDir)


	//
	// Setting up logging names
	//
	LogDir = strings.Join([]string{BaseDir, "logs"}, DirDelimiter)

	logger.Tracef(1,"Default log directory set to %s", LogDir)
	
	LogFileName = strings.Join([]string{LogDir, DirDelimiter, BaseName, "_", CurrentPID, ".", LogSuffix},"")

	logger.Tracef(1,"Default Log file set to %s",LogFileName)


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
