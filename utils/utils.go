package utils

// Standard imports

import "os"
import "os/signal"
import "path"
import "regexp"
import "runtime"
import "strings"
import "strconv"
import "syscall"

// Local imports

import "github.com/daviesluke/logger"

// Local functions 

type fn func() 

// Global functions 

func GetFunctionName() string {
	var callingFuncName string

	logger.Trace(1,"Getting back trace info ...")
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		callingFuncName = runtime.FuncForPC(pc).Name()
		logger.Tracef(1,"Calling function name set to %s", callingFuncName)
		callingFuncName = path.Base(callingFuncName)
		logger.Tracef(1,"Reduced function name to %s",callingFuncName)
	} else {
		logger.Trace(1,"Failed to get back trace info from runtime.Caller")
	}

	return callingFuncName
}


func CheckRegEx(checkString string, regEx string) bool {
	logger.Tracef(1,"Set regular expression -> %s", regEx)
	
	logger.Tracef(1,"Checking to see if string %s matches regex", checkString)
	stringMatch, err := regexp.MatchString(regEx, checkString)
	if err != nil {
		logger.Errorf("Invalid regular exporession %s", regEx)
	}

	if stringMatch {
		logger.Trace(1,"String matched!")
	} else {
		logger.Trace(1,"String NOT matched!")
	}

	logger.Tracef(1,"Returning %s", strconv.FormatBool(stringMatch))

	return stringMatch
}

func RemovePassword(checkString string) string {
	var callingFuncName string
	var userName        string

	callingFuncName = GetFunctionName()

	logger.Infof("%s - Removing any passwords found ...", callingFuncName)

	logger.Trace(1,"Checking string for passwords ...")

	if CheckRegEx(checkString, ".+/[^@]+") {
		logger.Warnf("%s\t- Password in connection string - consider using SEPS", callingFuncName)

		logger.Debug("Removing password for display ...")
	}

	upTokens := strings.SplitN(checkString, "/", 2)
	logger.Debugf("Split string using / delimiter into %d tokens.  Username is %s", len(upTokens), upTokens[0])

	userName = upTokens[0]
	logger.Tracef(1,"Username set to %s", userName)

	passWord := upTokens[1]
	logger.Trace(1,"Password set")

	if strings.Index(passWord,"@") != -1 {
		logger.Debug("Password portion contains a TNS Alias")

		passTokens := strings.SplitN(passWord, "@", 2)
		logger.Debugf("Split string using @ delimiter into %d tokens.  TNS Alias is %s", len(passTokens), passTokens[1])

		tnsAlias := passTokens[1]
		logger.Tracef(1,"TNS Alias set to %s", tnsAlias)

		userName = strings.Join([]string{userName, tnsAlias},"@")
		logger.Tracef(1,"Username now set to %s", userName)
	}

	logger.Infof("%s - Process complete - returning %s", callingFuncName, userName)

	return userName
}

func TrapSignal(runFunction fn) {
	var callingFuncName string

	callingFuncName = GetFunctionName()

	logger.Infof("%s - Trapping signals ...", callingFuncName)
	
	channel := make(chan os.Signal)
	logger.Debug("Channel set for signals")

	signal.Notify(channel, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	logger.Debug("Interrupt generated for signals 1,2,3 and 15")
	go func() {
		logger.Debug("About to block on signal input (In threaded process) ...")
		signalRecieved := <-channel
		logger.Infof("%s\t- Received signal %d", callingFuncName, signalRecieved)
		runFunction()
		logger.Debug("Exiting ...")
		os.Exit(1)
	}()
}
