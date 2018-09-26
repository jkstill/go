package logger

// standard imports

import "os"
import "path"
import "runtime"

// local imports

import "github.com/daviesluke/romana/rlog"

// Local functions

func getFunctionName() string {
        var callingFuncName string

        Trace(1,"Getting back trace info ...")
        pc, _, _, ok := runtime.Caller(2)
        if ok {
                callingFuncName = runtime.FuncForPC(pc).Name()
                Tracef(1,"Calling function name set to %s", callingFuncName)
                callingFuncName = path.Base(callingFuncName)
                Tracef(1,"Reduced function name to %s",callingFuncName)
        } else {
                Trace(1,"Failed to get back trace info from runtime.Caller")
        }

        return callingFuncName
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

func Trace(traceLevel int, message string) {
	rlog.Trace(traceLevel, message)
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

func Tracef(traceLevel int, messageFormat string, message ...interface{}) {
	rlog.Tracef(traceLevel, messageFormat, message...)
}

func Initialize(logDir string, logFileName string, logConfigFileName string) {

	//
	// Check level and if set to debug or less then enable extra logging info
	//
	Trace(1,"Checking logging level ...")

	if rlog.CheckLevel("DEBUG") {
		Trace(1,"Checking environment variable RLOG_LOG_LEVEL ...")

		envVar := os.Getenv("RLOG_LOG_LEVEL")
		if envVar == "" {
			Trace(1,"Tracing on but not at debug log level. Setting RLOG_LOG_LEVEL to DEBUG ...")

			os.Setenv("RLOG_LOG_LEVEL", "DEBUG")

			Trace(1,"RLOG_LOG_LEVEL set to DEBUG")
		} else {
			Tracef(1,"RLOG_LOG_LEVEL set to %s",envVar)
		}

		Trace(1,"Log Level set to DEBUG or lower")
		Trace(1,"Increasing logging information by setting RLOG_CALLER_INFO to yes")

		os.Setenv("RLOG_CALLER_INFO", "yes")

		Trace(1,"RLOG_CALLER_INFO set to yes")
		Trace(1,"Updating the environment ...")

		rlog.UpdateEnv()

		Trace(1,"Environment updated")
	}

	//
	// Make sure directory exists
	//
	Tracef(1,"Checking directory %s exists ...", logDir)

	if logFileInfo, err := os.Stat(logDir); err == nil {
		Tracef(1,"File %s exists. Checking if it is a directory ...", logDir)

		fileMode := logFileInfo.Mode()

		if fileMode.IsDir() {
			Tracef(1,"Directory %s exists",logDir)
		} else {
			Errorf("File %s is not a directory.  Exiting with errors", logDir)
		}
	} else {
		Tracef(1,"Log directory %s does not exist", logDir)
		Tracef(1,"Making log directory %s ...", logDir)
		if err := os.MkdirAll(logDir,0755); err != nil {
			Errorf("Unable to make log directory %s", logDir)
		}
		Tracef(1,"Log directory %s created", logDir)
	}
		
	//
	// Try opening the file
	//
	Tracef(1,"About to open file %s (modes append, write only, create, perm=0600) ...",logFileName)

	LogFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		Errorf("Unable to open logfile %s", logFileName)
	}
	Tracef(1,"Log file %s created",logFileName)

	//
	// Closing file - rlog will open it 
	//
	Tracef(1,"Closing file %s ...",logFileName)
	LogFile.Close()
	Tracef(1,"Closed file %s", logFileName)

	//
	// Set the output to the log file
	//
	Tracef(1,"Setting the variable RLOG_LOG_FILE to %s ...",logFileName)
	os.Setenv("RLOG_LOG_FILE",logFileName)

	// 
	// Turn off logging to the stderr if trace is switched off
	//
	Trace(1,"Getting environment variable RLOG_TRACE_LEVEL")
	rlogTraceLevel := os.Getenv("RLOG_TRACE_LEVEL")

	if rlogTraceLevel == "" {
		Trace(1,"RLOG_TRACE_LEVEL not set ...")
		Trace(1,"Setting the variable RLOG_LOG_STREAM to none")
		os.Setenv("RLOG_LOG_STREAM","none")
	} else {
		Tracef(1,"RLOG_TRACE_LEVEL set to %s. Leaving RLOG streaming on ...", rlogTraceLevel)
	}
		
	Trace(1,"Updating environment")
	rlog.UpdateEnv()
	Trace(1,"Environment updated")
	Debugf("Log file %s should now be open", logFileName)

	//
	// Check for any Logging Configuration 
	//
	Debugf("Checking file %s ...", logConfigFileName)

	if _, err := os.Stat(logConfigFileName); err == nil {
		Debugf("File %s exists", logConfigFileName)
		Tracef(1,"Setting the config file to %s ...", logConfigFileName)
		rlog.SetConfFile(logConfigFileName)
		Debugf("Config file set to %s", logConfigFileName)
	} else {
		Debugf("File %s does not exist", logConfigFileName)
	}
}
