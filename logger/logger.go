package logger

// standard imports

import "io"
import "os"
import "path/filepath"
import "runtime"

// local imports

import "github.com/daviesluke/romana/rlog"

// Local functions

func getFunctionName() string {
        var callingFuncName string

        Trace("Getting back trace info ...")
        pc, _, _, ok := runtime.Caller(2)
        if ok {
                callingFuncName = runtime.FuncForPC(pc).Name()
                Tracef("Calling function name set to %s", callingFuncName)
                callingFuncName = filepath.Base(callingFuncName)
                Tracef("Reduced function name to %s",callingFuncName)
        } else {
                Trace("Failed to get back trace info from runtime.Caller")
        }

        return callingFuncName
}

func copyFileContents( oldLogFileName string , newLogFileName string ) {
	Trace("Copying file contents ...")

	oldLogFile, err := os.Open(oldLogFileName)
	if err != nil {
		Errorf("Unable to open read file %s for copying", oldLogFileName)
	}

	defer oldLogFile.Close()

	newLogFile, err := os.OpenFile(newLogFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		Errorf("Unable to open write file %s for copying", newLogFileName)
	}

	defer newLogFile.Close()

	// Now copy the contents

	if _, err := io.Copy(newLogFile, oldLogFile); err != nil {
		Errorf("Unable to copy file %s to %s", oldLogFileName, newLogFileName)
	}

	// Flush anything out to disk 

	newLogFile.Sync()
	
	// Deferred files to close at end 

	Trace("Process complete")
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
	Trace("Checking logging level ...")

	if rlog.CheckLevel("DEBUG") {
		Trace("Checking environment variable RLOG_LOG_LEVEL ...")

		envVar := os.Getenv("RLOG_LOG_LEVEL")
		if envVar == "" {
			Trace("Tracing on but not at debug log level. Setting RLOG_LOG_LEVEL to DEBUG ...")

			os.Setenv("RLOG_LOG_LEVEL", "DEBUG")

			Trace("RLOG_LOG_LEVEL set to DEBUG")
		} else {
			Tracef("RLOG_LOG_LEVEL set to %s",envVar)
		}

		Trace("Log Level set to DEBUG or lower")
		Trace("Increasing logging information by setting RLOG_CALLER_INFO to yes")

		os.Setenv("RLOG_CALLER_INFO", "yes")

		Trace("RLOG_CALLER_INFO set to yes")
		Trace("Updating the environment ...")

		rlog.UpdateEnv()

		Trace("Environment updated")
	}

	//
	// Make sure directory exists
	//
	Tracef("Checking directory %s exists ...", logDir)

	if logFileInfo, err := os.Stat(logDir); err == nil {
		Tracef("File %s exists. Checking if it is a directory ...", logDir)

		fileMode := logFileInfo.Mode()

		if fileMode.IsDir() {
			Tracef("Directory %s exists",logDir)
		} else {
			Errorf("File %s is not a directory.  Exiting with errors", logDir)
		}
	} else {
		Tracef("Log directory %s does not exist", logDir)
		Tracef("Making log directory %s ...", logDir)
		if err := os.MkdirAll(logDir,0755); err != nil {
			Errorf("Unable to make log directory %s", logDir)
		}
		Tracef("Log directory %s created", logDir)
	}
		
	//
	// Try opening the file
	//
	Tracef("About to open file %s (modes append, write only, create, perm=0600) ...",logFileName)

	LogFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)

	if err != nil {
		Errorf("Unable to open logfile %s", logFileName)
	}
	Tracef("Log file %s created",logFileName)

	//
	// Closing file - rlog will open it 
	//
	Tracef("Closing file %s ...",logFileName)
	LogFile.Close()
	Tracef("Closed file %s", logFileName)

	//
	// Set the output to the log file
	//
	Tracef("Setting the variable RLOG_LOG_FILE to %s ...",logFileName)
	os.Setenv("RLOG_LOG_FILE",logFileName)

	// 
	// Turn off logging to the stderr if trace is switched off
	//
	Trace("Getting environment variable RLOG_TRACE_LEVEL")
	rlogTraceLevel := os.Getenv("RLOG_TRACE_LEVEL")

	if rlogTraceLevel == "" {
		Trace("RLOG_TRACE_LEVEL not set ...")
		Trace("Setting the variable RLOG_LOG_STREAM to none")
		os.Setenv("RLOG_LOG_STREAM","none")
	} else {
		Tracef("RLOG_TRACE_LEVEL set to %s. Leaving RLOG streaming on ...", rlogTraceLevel)
	}
		
	Trace("Updating environment")
	rlog.UpdateEnv()
	Trace("Environment updated")
	Debugf("Log file %s should now be open", logFileName)

	//
	// Check for any Logging Configuration 
	//
	Debugf("Checking file %s ...", logConfigFileName)

	if _, err := os.Stat(logConfigFileName); err == nil {
		Debugf("File %s exists", logConfigFileName)
		Tracef("Setting the config file to %s ...", logConfigFileName)
		rlog.SetConfFile(logConfigFileName)
		Debugf("Config file set to %s", logConfigFileName)
	} else {
		Debugf("File %s does not exist", logConfigFileName)
	}
}

func CopyLog( oldLogFileName string , newLogFileName string ) {
	Infof("Copying log %s to %s ...", oldLogFileName, newLogFileName)

	// Check the newLogFileName does not already exist
	if _, err := os.Stat(newLogFileName); err == nil {
		Errorf("New log file %s exists with error %d.  Exiting ...", newLogFileName, err)
	}

	// Turn off output 
	os.Unsetenv("RLOG_LOG_FILE")
	rlog.UpdateEnv()
	Trace("Turned off logging")

	// File should be closed - ready to copy

	copyFileContents( oldLogFileName, newLogFileName )

	// Remove old log 

	Trace("Removing old log ...")

	err := os.Remove(oldLogFileName)
	if err != nil {
		Errorf("Unable to remove old log %s", oldLogFileName)
	}

	// Turn on output
	os.Setenv("RLOG_LOG_FILE",newLogFileName)
	rlog.UpdateEnv()
	Trace("Turned on logging")

	Info("Process complete")
}

