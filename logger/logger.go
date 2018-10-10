package logger

// standard imports

import "os"
import "path/filepath"
import "runtime"

// local imports

import "github.com/daviesluke/romana/rlog"

// Local functions

func trace2(message string) {
	rlog.Trace(2, message)
}

func Tracef2(messageFormat string, message ...interface{}) {
	rlog.Tracef(2, messageFormat, message...)
}

func getFunctionName() string {
        var callingFuncName string

        trace2("Getting back trace info ...")
        pc, _, _, ok := runtime.Caller(2)
        if ok {
                callingFuncName = runtime.FuncForPC(pc).Name()
                tracef2("Calling function name set to %s", callingFuncName)
                callingFuncName = filepath.Base(callingFuncName)
                tracef2("Reduced function name to %s",callingFuncName)
        } else {
                trace2("Failed to get back trace info from runtime.Caller")
        }

        return callingFuncName
}

func setConfFile (logConfigFileName string) {
	//
	// Check for any Logging Configuration 
	//
	Debugf("Checking file %s ...", logConfigFileName)

	if _, err := os.Stat(logConfigFileName); err == nil {
		Debugf("File %s exists", logConfigFileName)
		tracef2("Setting the config file to %s ...", logConfigFileName)
		rlog.SetConfFile(logConfigFileName)
		Debugf("Config file set to %s", logConfigFileName)
	} else {
		Debugf("File %s does not exist", logConfigFileName)
	}
}


// Global Functions

func Info(message string) {
	callingFuncName := getFunctionName()

	rlog.Infof("%s - %s", callingFuncName, message)
}

func Warn(message string) {
	callingFuncName := getFunctionName()

	rlog.Warnf("%s - %s", callingFuncName, message)
}

func Error(message string) {
	callingFuncName := getFunctionName()

	// Enable out to stream as well as file

	os.Setenv("RLOG_LOG_STREAM","stderr")
	rlog.UpdateEnv()

	rlog.Errorf("%s - %s", callingFuncName, message)

	os.Exit(1)
}

func Critical(message string) {
	callingFuncName := getFunctionName()

	// Enable out to stream as well as file

	os.Setenv("RLOG_LOG_STREAM","stderr")
	rlog.UpdateEnv()

	rlog.Criticalf("%s - %s", callingFuncName, message)

	os.Exit(1)
}

func Debug(message string) {
	rlog.Debug(message)
}

func Trace(message string) {
	rlog.Trace(1, message)
}

func Infof(messageFormat string, message ...interface{}) {
	callingFuncName := getFunctionName()

	messageFormat = callingFuncName + " - " + messageFormat

	rlog.Infof(messageFormat, message...)
}

func Warnf(messageFormat string, message ...interface{}) {
	callingFuncName := getFunctionName()
	
	messageFormat = callingFuncName + " - " + messageFormat

	rlog.Warnf(messageFormat, message...)
}

func Errorf(messageFormat string, message ...interface{}) {
	callingFuncName := getFunctionName()

	// Enable out to stream as well as file

	os.Setenv("RLOG_LOG_STREAM","stderr")
	rlog.UpdateEnv()

	messageFormat = callingFuncName + " - " + messageFormat

	rlog.Errorf(messageFormat, message...)

	os.Exit(1)
}

func Criticalf(messageFormat string, message ...interface{}) {
	callingFuncName := getFunctionName()

	// Enable out to stream as well as file

	os.Setenv("RLOG_LOG_STREAM","stderr")
	rlog.UpdateEnv()

	messageFormat = callingFuncName + " - " + messageFormat

	rlog.Criticalf(messageFormat, message...)

	os.Exit(2)
}

func Debugf(messageFormat string, message ...interface{}) {
	rlog.Debugf(messageFormat, message...)
}

func Tracef(messageFormat string, message ...interface{}) {
	rlog.Tracef(1, messageFormat, message...)
}

func Initialize(logDir string, logFileName string, logConfigFileName string) {

	//
	// Check level and if set to debug or less then enable extra logging info
	//
	trace2("Checking logging level ...")

	if rlog.CheckLevel("DEBUG") {
		trace2("Checking environment variable RLOG_LOG_LEVEL ...")

		envVar := os.Getenv("RLOG_LOG_LEVEL")
		if envVar == "" {
			trace2("Tracing on but not at debug log level. Setting RLOG_LOG_LEVEL to DEBUG ...")

			os.Setenv("RLOG_LOG_LEVEL", "DEBUG")

			trace2("RLOG_LOG_LEVEL set to DEBUG")
		} else {
			tracef2("RLOG_LOG_LEVEL set to %s",envVar)
		}

		trace2("Log Level set to DEBUG or lower")
		trace2("Increasing logging information by setting RLOG_CALLER_INFO to yes")

		os.Setenv("RLOG_CALLER_INFO", "yes")

		trace2("RLOG_CALLER_INFO set to yes")
		trace2("Updating the environment ...")

		rlog.UpdateEnv()

		trace2("Environment updated")
	}

	//
	// Make sure directory exists
	//
	tracef2("Checking directory %s exists ...", logDir)

	if logFileInfo, err := os.Stat(logDir); err == nil {
		tracef2("File %s exists. Checking if it is a directory ...", logDir)

		fileMode := logFileInfo.Mode()

		if fileMode.IsDir() {
			tracef2("Directory %s exists",logDir)
		} else {
			Errorf("File %s is not a directory.  Exiting with errors", logDir)
		}
	} else {
		tracef2("Log directory %s does not exist", logDir)
		tracef2("Making log directory %s ...", logDir)
		if err := os.MkdirAll(logDir,0755); err != nil {
			Errorf("Unable to make log directory %s", logDir)
		}
		tracef2("Log directory %s created", logDir)
	}
		
	//
	// Try opening the file
	//
	tracef2("About to open file %s (modes append, write only, create, perm=0600) ...",logFileName)

	LogFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)

	if err != nil {
		Errorf("Unable to open logfile %s", logFileName)
	}
	tracef2("Log file %s created",logFileName)

	//
	// Closing file - rlog will open it 
	//
	tracef2("Closing file %s ...",logFileName)
	LogFile.Close()
	tracef2("Closed file %s", logFileName)

	//
	// Set the output to the log file
	//
	tracef2("Setting the variable RLOG_LOG_FILE to %s ...",logFileName)
	os.Setenv("RLOG_LOG_FILE",logFileName)

	// 
	// Turn off logging to the stderr if trace is switched off
	//
	trace2("Getting environment variable RLOG_TRACE_LEVEL")
	rlogTraceLevel := os.Getenv("RLOG_TRACE_LEVEL")

	if rlogTraceLevel == "" {
		trace2("RLOG_TRACE_LEVEL not set ...")
		trace2("Setting the variable RLOG_LOG_STREAM to none")
		os.Setenv("RLOG_LOG_STREAM","none")
	} else {
		tracef2("RLOG_TRACE_LEVEL set to %s. Leaving RLOG streaming on ...", rlogTraceLevel)
	}
		
	trace2("Updating environment")
	rlog.UpdateEnv()
	trace2("Environment updated")
	Debugf("Log file %s should now be open", logFileName)

	setConfFile(logConfigFileName)
}

func RenameLog(oldLogFileName string , newLogFileName string, logConfigFileName string ) {
	Infof("Renaming log %s to %s ...", oldLogFileName, newLogFileName)

	// Check the newLogFileName does not already exist
	if _, err := os.Stat(newLogFileName); err == nil {
		Errorf("New log file %s exists with error %d.  Exiting ...", newLogFileName, err)
	}

	// Redirect output to stdout

	rlog.SetOutput(os.Stdout)
	trace2("Redirected output to stderr")

	// File should be closed - ready to rename

	if err := os.Rename(oldLogFileName, newLogFileName ); err != nil {
		Errorf("Unable to rename the file - %s", err)
	}

	// Turn on output
	os.Setenv("RLOG_LOG_FILE",newLogFileName)
	rlog.UpdateEnv()
	trace2("Turned on logging")

	setConfFile(logConfigFileName)

	Info("Process complete")
}
