package utils

// Standard imports

import "os"
import "os/signal"
import "regexp"
import "strings"
import "strconv"
import "syscall"

// Local imports

import "github.com/daviesluke/logger"

// Local functions 

type fn func() 

// Global functions 

func CheckRegEx(checkString string, regEx string) bool {
	logger.Tracef("Set regular expression -> %s", regEx)
	
	logger.Tracef("Checking to see if string %s matches regex", checkString)
	stringMatch, err := regexp.MatchString(regEx, checkString)
	if err != nil {
		logger.Errorf("Invalid regular exporession %s", regEx)
	}

	if stringMatch {
		logger.Trace("String matched!")
	} else {
		logger.Trace("String NOT matched!")
	}

	logger.Tracef("Returning %s", strconv.FormatBool(stringMatch))

	return stringMatch
}

func RemovePassword(checkString string) string {
	var userName        string

	logger.Info("Removing any passwords found ...")

	logger.Trace("Checking string for passwords ...")

	if CheckRegEx(checkString, ".+/[^@]+") {
		logger.Warn("Password in connection string - consider using SEPS")

		logger.Debug("Removing password for display ...")
	}

	upTokens := strings.SplitN(checkString, "/", 2)
	logger.Debugf("Split string using / delimiter into %d tokens.  Username is %s", len(upTokens), upTokens[0])

	userName = upTokens[0]
	logger.Tracef("Username set to %s", userName)

	passWord := upTokens[1]
	logger.Trace("Password set")

	if strings.Index(passWord,"@") != -1 {
		logger.Debug("Password portion contains a TNS Alias")

		passTokens := strings.SplitN(passWord, "@", 2)
		logger.Debugf("Split string using @ delimiter into %d tokens.  TNS Alias is %s", len(passTokens), passTokens[1])

		tnsAlias := passTokens[1]
		logger.Tracef("TNS Alias set to %s", tnsAlias)

		userName = strings.Join([]string{userName, tnsAlias},"@")
		logger.Tracef("Username now set to %s", userName)
	}

	logger.Infof("Process complete - returning %s", userName)

	return userName
}

func TrapSignal(runFunction fn) {
	logger.Infof("Trapping signals ...")
	
	channel := make(chan os.Signal)
	logger.Debug("Channel set for signals")

	signal.Notify(channel, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	logger.Debug("Interrupt generated for signals 1,2,3 and 15")
	go func() {
		logger.Debug("About to block on signal input (In threaded process) ...")
		signalRecieved := <-channel
		logger.Infof("Received signal %d", signalRecieved)
		runFunction()
		logger.Debug("Exiting ...")
		os.Exit(1)
	}()

	logger.Infof("Process complete")
}
