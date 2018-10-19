package utils

// Standard imports

import "bufio"
import "fmt"
import "io"
import "os"
import "os/signal"
import "path/filepath"
import "regexp"
import "strings"
import "syscall"
import "time"

// Local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/mitchellh/go-ps"

// Local functions 

type fn func() 

// Global functions 

func CheckRegEx(checkString string, regEx string) bool {
	logger.Tracef("Set regular expression -> %s", regEx)
	
	logger.Tracef("Checking to see if string %s matches regex", checkString)
	found, err := regexp.MatchString(regEx, checkString)
	if err != nil {
		logger.Errorf("Invalid regular expression %s", regEx)
	}

	logger.Tracef("Returning %t", found)

	return found
}

func CheckRegExGroup(checkString string, regEx string, ignoreRegEx string, regExGroupCount int) bool {
	logger.Tracef("Regular expression %s, ignore list %s", regEx, ignoreRegEx)

	fullRegEx := strings.Join( []string{ ignoreRegEx, regEx }, "|")

	found := false

	// Compile the regular expression
	if re, err := regexp.Compile(fullRegEx); err == nil {
		
		// Increment the group count as we add the number of ignore groups to the final regexp
		groupCount := regExGroupCount+1
		logger.Tracef("Group count %d", groupCount)

		// $1 for example is produced only when string matches the first group
		groupString := fmt.Sprintf("$%d", groupCount)
		logger.Tracef("Group string %s", groupString)

		// This leaves just the string that matches the last group 
		// i.e. has survived removal during the matching against the ignore list
		ignoreString := re.ReplaceAllString(checkString, groupString)
		logger.Debugf("String after ignore list - %s", ignoreString)

		// Now check this against the orginal string to see if the line still matches
		found = CheckRegEx(ignoreString, regEx)
	}

	logger.Tracef("Returning %t", found)

	return found
}

func RemovePassword(checkString string, printWarn bool) string {
	logger.Debug("Removing any passwords found ...")

	var userName        string

	logger.Trace("Checking string for passwords ...")

	if CheckRegEx(checkString, ".+/[^@]+") {
		if printWarn {
			logger.Warn("Password in connection string - consider using SEPS")
		}

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

	logger.Debugf("Process complete - returning %s", userName)

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

func LookupFile(searchFileName string, searchString string, searchIndex int, returnIndex int, delimiter string, returnCounter int) string {
	logger.Infof("Searching for %s in position %d in file %s demilited by %s ...", searchString, searchIndex, searchFileName, delimiter)

	logger.Tracef("Trying to open file %s ...", searchFileName)

	searchFile, err := os.Open(searchFileName)
	if err != nil {
		logger.Errorf("Unable to find file %s", searchFileName)
	}

	// Defer the close to auto close at the end of the procedure 

	defer searchFile.Close()
	logger.Tracef("Deferred closing of file %s at end of function", searchFileName)

	searchScanner := bufio.NewScanner(searchFile)
	logger.Tracef("Set up scanner for search file. Entering loop ...")

	lineNo := 0
	logger.Tracef("Set up variable LineNo and set to %d", lineNo)

	findCounter  := 1
	returnString := ""

	for searchScanner.Scan() {
		lineNo++
		logger.Tracef("Line number incremented to %d", lineNo)

		// Ignore blank lines and comments

		searchLine := strings.TrimSpace(searchScanner.Text())
		logger.Debugf("Trimmed search file line number %d contents - %s", lineNo, searchLine)

		if searchLine == "" || searchLine[0] == '#' {
			logger.Trace("Comment or blank line - ignoring line")
			continue
		}

		variableTokens := strings.Split(searchLine, delimiter)
		logger.Tracef("Line split into %d tokens using %s as delimiter", len(variableTokens), delimiter)

		if len(variableTokens) == 0 {
			logger.Trace("No tokens - ignoring line")
			continue
		}

		if strings.TrimSpace(variableTokens[searchIndex-1]) == searchString {
			if findCounter == returnCounter {
				returnString = strings.TrimSpace(variableTokens[returnIndex-1])
				logger.Infof("Search criteria found.  Returning entry %d => %s",  returnIndex, returnString)
				break
			} else {
				findCounter++
			}
		} else {
			logger.Tracef("String %s does not match search %s", strings.TrimSpace(variableTokens[searchIndex-1]), searchString)
		}
	}

	return returnString
}

func CopyFileContents( fromFileName string , toFileName string , regEx string ) {
	logger.Debugf("Copying from %s to %s", fromFileName, toFileName)

        fromFile, err := os.Open(fromFileName)
        if err != nil {
               logger.Errorf("Unable to open read file %s for copying", fromFileName)
        }

        defer fromFile.Close()

        toFile, err := os.OpenFile(toFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
        if err != nil {
                logger.Errorf("Unable to open write file %s for copying", toFileName)
        }

        defer toFile.Close()

	if regEx == "" {
		// Copy the entire contents
	
		if _, err := io.Copy(toFile, fromFile); err != nil {
			logger.Errorf("Unable to copy file %s to %s", fromFileName, toFileName)
		}
	} else {
		logger.Debugf("Using regular expression %s to write files", regEx)

		// Copy line by line if matches the regular expression

		fromScanner := bufio.NewScanner(fromFile)

		for fromScanner.Scan() {
			logger.Tracef("Checking line %s", fromScanner.Text())

			if CheckRegEx(fromScanner.Text(), regEx) {
				if _, err := toFile.WriteString(fromScanner.Text()+"\n"); err != nil {
					logger.Errorf("Unable to write to file %s", toFileName)
				}
				logger.Trace("Written to file")
			} else {
				logger.Trace("Ignored line")
			}
		}
	}

        // Flush anything out to disk

        toFile.Sync()

        // Deferred files to close at end

        logger.Debug("Process complete")
}

func CheckProcess (pid int, processName string) (bool, bool) {
	logger.Debugf("Checking PID %d is running process %s", pid, processName)

	pidAlive  := false
	pidIsName := false

	if checkProcess , err := ps.FindProcess(pid); checkProcess != nil && err == nil {
		logger.Debugf("Pid %d is running", pid)

		pidAlive     = true

		pidName      := checkProcess.Executable()

		pidName       = filepath.Base(pidName)
		pidNameParts := strings.SplitN(pidName,".",2)
		pidName       = pidNameParts[0]

		logger.Infof("PID %d found, Process is running %s", pid, pidName)

		if pidName == processName {
			logger.Debugf("Pid %d matches name %s", pid, processName)
			pidIsName = true
		}
	} else {
		logger.Debugf("Pid %d not found running", pid)
	}

	return pidAlive, pidIsName
}

func FindFiles ( dirPath string, fileRegEx string, daysOld int ) []string {
	logger.Debugf("Finding files of name %s in directory %s older than %d days ...", fileRegEx, dirPath, daysOld)

	var fileList []string

	// Get time 

	timeFormat := "2006-Jan-02 15:04:05"

	checkTime := time.Now()

	// Calculate modified date 
	checkTime = checkTime.AddDate(0, 0, daysOld*-1) 

	logger.Debugf("Check time set to %s",checkTime.Format(timeFormat))
	
	// Check that directory exists and is a directory
	if dirInfo , err := os.Stat(dirPath); err == nil && dirInfo.IsDir() {
		if dir , err := os.Open(dirPath); err == nil {
			if dirList, err := dir.Readdirnames(0); err == nil {
				for _, fileName := range dirList {
					// Check file name
					if CheckRegEx( fileName, fileRegEx ) {
						logger.Debugf("File %s Matched Reg Exp!", fileName)

						// Get the file modification date

						fullFile := filepath.Join( dirPath , fileName)

						if fileInfo, err := os.Stat(fullFile); err == nil {
							fileModTime := fileInfo.ModTime()

							logger.Debugf("File has time stamp %s", fileModTime.Format(timeFormat))

							// Check dates 
							if fileModTime.Before(checkTime) {
								logger.Debug("Added to file list")
								fileList = append(fileList, fullFile)
							} else {
								logger.Debug("Not added to file list")
							}
						} else {
							logger.Warnf("Unable to find file %s. Perhaps it has been deleted", fullFile)
						}
					} else {
						logger.Debugf("File %s did not match reg exp!", fileName)
					}
				}
			} else {
				logger.Warnf("Unable to read directory %s", dirPath)
			}

			// Close directory
			dir.Close()
		} else {
			logger.Warnf("Unable to open directory %s", dirPath)
		}
	} else {
		logger.Warnf("%s is not a directory", dirPath)
	}

	return fileList
}

func FindInFile( fileName string, regEx string, ignoreRegEx string, regGroup int ) bool {
	logger.Debug("Checking file for regular expression ...")

	found := false

	if file, err := os.Open(fileName); err == nil {
		fileScanner := bufio.NewScanner(file)

		for fileScanner.Scan() {
			if ignoreRegEx == "" {
				found = CheckRegEx(fileScanner.Text(),regEx)
			} else {
				found = CheckRegExGroup(fileScanner.Text(),regEx,ignoreRegEx,regGroup) 
			}

			if found {
				logger.Debug("Found matching string")
				break
			}
		}

		file.Close()
	}

	logger.Debugf("Returning %t", found)

	return found
}

func ReplaceString( inString string, regEx string, replaceString string ) string {
	logger.Debug("Replacing regex by string ...")

	outString := inString

	if re, err := regexp.Compile(regEx); err == nil {
		outString = re.ReplaceAllString(inString, replaceString)
	} else {
		logger.Debugf("Invalid regular expression %s", regEx)
	}

	logger.Tracef("Returning %s", outString)

	return outString
}

func CountLines( fileName string ) int {
	logger.Debugf("Count lines in file %s ...", fileName)

	lineCount := 0

	if file, err := os.Open(fileName); err == nil {
		fileScanner := bufio.NewScanner(file)

		for fileScanner.Scan() {
			lineCount++
		}

		file.Close()
	}

	logger.Debugf("Returning %d", lineCount)
	
	return lineCount
}
