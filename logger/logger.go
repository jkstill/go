package logger

// standard imports

import "bufio"
import "fmt"
import "io"
import "net/smtp"
import "os"
import "os/user"
import "path/filepath"
import "runtime"
import "strconv"
import "strings"
import "time"

// local imports

import "github.com/daviesluke/romana/rlog"

// Local variables

var startTime   time.Time

var database    string
var historyFile string
var scriptName  string

var emailServer string

var successEmails []string
var errorEmails   []string

var currentLog string

// Local functions

func copyLog(oldLog, newLog string) {
	trace2("Copying files ...")

	old, err := os.Open(oldLog)
	if err != nil {
		Errorf("Unable to open log file %s for reading", oldLog)
	}

	defer old.Close()

	new, err := os.OpenFile(newLog, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		Errorf("Unable to open log file %s for writing", newLog)
	}

	defer new.Close()

	if _, err := io.Copy(new,old); err != nil {
		Errorf("Unable to write file %s", newLog)
	}

	new.Sync()

	trace2("File copied")
}

func info(message string) {
	rlog.Info(message)
}

func infof(messageFormat string, message ...interface{}) {
	rlog.Infof(messageFormat, message)
}

func trace2(message string) {
	rlog.Trace(2, message)
}

func tracef2(messageFormat string, message ...interface{}) {
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

func init () {

	// Get default e-mail Server

	var err error

	emailServer, err = os.Hostname()
	if err != nil {
		tracef2("Unable to get hostname - %s", err)
	}

	emailServer = strings.Join( []string{ emailServer, "25" }, ":" )
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

	SendLog("FAILURE")

	WriteHistory("FAILURE")

	os.Exit(1)
}

func Critical(message string) {
	callingFuncName := getFunctionName()

	// Enable out to stream as well as file

	os.Setenv("RLOG_LOG_STREAM","stderr")
	rlog.UpdateEnv()

	rlog.Criticalf("%s - %s", callingFuncName, message)

	SendLog("FAILURE")

	WriteHistory("FAILURE")

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

	SendLog("FAILURE")

	WriteHistory("FAILURE")

	os.Exit(1)
}

func Criticalf(messageFormat string, message ...interface{}) {
	callingFuncName := getFunctionName()

	// Enable out to stream as well as file

	os.Setenv("RLOG_LOG_STREAM","stderr")
	rlog.UpdateEnv()

	messageFormat = callingFuncName + " - " + messageFormat

	rlog.Criticalf(messageFormat, message...)

	SendLog("FAILURE")

	WriteHistory("FAILURE")

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

	currentLog = logFileName
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

	// Cannot use Rename as may be on different filesystems so copy the log
	copyLog(oldLogFileName, newLogFileName )

	// So now need to remove old log

	if err := os.Remove(oldLogFileName); err != nil {
		Errorf("Unable to remove old log file %s", oldLogFileName)
	}

	// Turn on output
	os.Setenv("RLOG_LOG_FILE",newLogFileName)
	rlog.UpdateEnv()
	trace2("Turned on logging")

	setConfFile(logConfigFileName)

	currentLog = newLogFileName

	Debug("Process complete")
}

func CopyFileToLog(title string, fileName string, logConfigFileName string) {
	Debug("Copying file content to log ...")

	// Using local functiont to not print function calls

	info(title)

	// Turn off some output logging
	
	os.Setenv("RLOG_LOG_NOTIME","yes")
	rlog.UpdateEnv()

	Debug("Not logging time now")

	if file, err := os.Open(fileName); err == nil {
		fileScanner := bufio.NewScanner(file)

		for fileScanner.Scan() {
			info(fileScanner.Text())
		}

		file.Close()
	} else {
		Errorf("Unable to open file %s - %s", fileName, err)
	}

	os.Unsetenv("RLOG_LOG_NOTIME")
	rlog.UpdateEnv()
	setConfFile(logConfigFileName)
	
	Debug("Process complete")
}

func SetStartTime () {
	startTime = time.Now()
}

func SetHistoryVars ( fileName string, db string, funcName string ) {
	Trace("Setting history variables ...")

	historyFile = fileName
	database    = db
	scriptName  = funcName

	Tracef("Settings: History file %s, Database %s, scriptName %s", historyFile, database, scriptName)

	Trace("Process complete")
}
	
func WriteHistory (status string) {
	Trace("Writing history file ...")

	if history, err := os.OpenFile(historyFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND , 0600); err == nil {
		
		timeDiff := time.Since(startTime)

		writeString := strings.Join ( []string{ time.Now().Format("2006/02/01:15:04:05"), database, scriptName, strconv.FormatFloat(timeDiff.Seconds(),'f',0,64), status }, " ")
	
		Tracef("Writing - %s", writeString)

		history.WriteString(writeString+"\n")

		history.Sync()

		history.Close()
	} else {
		Tracef("Unable to open file %s - %s", historyFile, err)
	}

	Trace("Process complete")
}

func SetEmailServer( serverString string ) {
	Trace("Setting email server ...")

	emailServer = serverString

	Trace("Process complete")
}

func SetEmailRecipients( errorList []string, successList []string ) {
	Trace("Setting the e-mail recipients ...")

	errorEmails   = errorList
	successEmails = successList

	Trace("Process complete")
}

func SendLog (status string) {
	// Check we have some recipients otherwise ignore

	var recipientList []string
	recipientCount := 0

	if status == "ERROR" {
		recipientList = errorEmails
		recipientCount = len(recipientList)
	} else if status == "SUCCESS" {
		recipientList = successEmails
		recipientCount = len(recipientList)
	}
		
	if recipientCount != 0 {
		
		// Redirect output to stdout and close log file

		rlog.SetOutput(os.Stdout)
		trace2("Redirected output to stderr")

		// Connect to the SMTP server.
		emailConnect, err := smtp.Dial(emailServer)
		if err != nil {
			Errorf("Unable to connect to the mail server %s", emailServer)
		}

		// Set the sender 

		userInfo, err := user.Current()
		if err != nil {
			Error("Unable to get the current user information")
		}

		hostName, err := os.Hostname()
		if err != nil {
			Error("Unable to get the host name")
		}

		sender := strings.Join( []string{ userInfo.Username , hostName }, "@" )

		if err := emailConnect.Mail(sender); err != nil {
			Errorf("Unable to set the sender address - %s", sender)
		}

		for _, receiver := range recipientList {
			if err := emailConnect.Rcpt(receiver); err != nil {
				Errorf("Unable to set the receiver address %s - %s", receiver)
			}
		}

		// Open up the body writer

		body, err := emailConnect.Data()
		if err != nil {
			Error("Unable to open writer for e-mail")
		}

		// Set up the header

		baseName, err := os.Executable()
		baseName = filepath.Base(baseName)
		baseSplit := strings.SplitN(baseName,".",2)
		baseName = baseSplit[0]

		fmt.Fprintf(body, "Subject: %s ran with status %s\n", baseName, status)

		// Send the log file 

		if currentLogFile, err := os.Open(currentLog); err == nil {
			logScanner := bufio.NewScanner(currentLogFile)

			for logScanner.Scan() {
				if _, err := fmt.Fprintf(body, "%s\n", logScanner.Text()); err != nil {
					Error("Unable to write to e-mail body")
				}
			}

			currentLogFile.Close()
		} else {
			Errorf("Unable to open log file %s", currentLog)
		}

		if err := body.Close(); err != nil {
			Errorf("Unable to close writer - %s", err)
		}

		// now send the final quit to send the mail

		if err := emailConnect.Quit(); err != nil {
			Errorf("Unable to finalize e-mail - %s", err)
		}
	}
}
