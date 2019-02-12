package locker

// Standard imports

import "bufio"
import "os"
import "strings"
import "strconv"
import "time"

// Local imports

import "../../filelock"
import "../../logger"
import "../../setup"
import "../../utils"
import "../../run_rman/config"


// local Variables

// Local functions

func checkLock( lockFileName string , lockName string , timeOutMins int ) {
	var lockFile *os.File
	var err      error

	logger.Debugf("Lock file -> %s", lockFileName)
	logger.Debugf("Lock Name -> %s", lockName)
	logger.Debugf("Time Out  -> %d mins", timeOutMins)

	// First check to see if file exists

	logger.Tracef("Attempting to open the file %s ...", lockFileName)

	if lockFile , err = os.Open(lockFileName); err == nil {
		logger.Infof("File %s found. Checking contents ...", lockFileName)

		lockFile.Close()
		logger.Debug("Closed the file")

		CleanLockFile(lockFileName, lockName, 0)

		lockCounter := 0
		logger.Debug("Initialized lock counter to zero")

		// File may now not exist after cleaning - try opening it again in loop

		for {
			lockFound := false

			if lockFile , err = os.Open(lockFileName); err == nil {
				logger.Debugf("Opened lock file %s", lockFileName)

				// Now let's read the lock file for the lock entry
				lockScanner := bufio.NewScanner(lockFile)
	
				logger.Debug("Scanning file ...")

				for lockScanner.Scan() {
					variableTokens := strings.SplitN(lockScanner.Text(), " ", 2)

					lockPID, _  := strconv.Atoi(variableTokens[0])
					fileLockName := variableTokens[1]

					logger.Debugf("Found PID %d with lock name %s", lockPID, fileLockName)

					if fileLockName == lockName {
						logger.Warnf("Process %d has already locked this process using lock name %s!", lockPID, fileLockName)
						lockFound = true
						break
					}
				}

				// Close file
				lockFile.Close()
				logger.Debug("Lock File closed")

				if lockFound {
					lockCounter++
					logger.Debugf("Lock counter incremented to %d", lockCounter)
				}
			} else {
				logger.Debugf("Unable to open file %s", lockFileName)
			}

			if lockFound {
				if lockCounter > timeOutMins {
					logger.Errorf("Unable to obtain the lock %s. Exiting ...", lockName)
				}
				logger.Info("Sleeping for 60 seconds ...")
				time.Sleep(60 * time.Second)
			} else {
				break
			}
		} 
	} else {
		logger.Debugf("Unable to open file %s", lockFileName)
	}

	logger.Debug("Process complete")
}


// Global functions

func LockProcess (lockName, database string) {
	logger.Info("Locking process ...")

	// Reset the number of minutes to wait before locking process

	if lockName != "" {
		checkLockMins, _ := strconv.Atoi(config.ConfigValues["CheckLockMins"])

		checkLock(setup.LockFileName, lockName, checkLockMins)

		// If we get to here then add the entry 

		AddLockEntry(setup.LockFileName, setup.CurrentPID,  lockName)
	} else {
		logger.Info("No lock string provided. No locking necessary")
	}

	logger.Info("Process complete")
}

func RemoveLockEntry(lockFileName string, lockPID string) {
	logger.Infof("Lock File : %s", lockFileName)
	logger.Infof("Lock PID  : %s", lockPID)

	// To write the file - take a real lock
	
	filelock.LockFile(lockFileName,1)
	
	newLockFileName := strings.Join( []string{ lockFileName, setup.CurrentPID }, ".")
	logger.Debugf("New lock file name set to %s", newLockFileName)

	fileWriteSize   := 0

	// Open the file for copying
	if lockFile , err := os.Open(lockFileName); err == nil {
		// Open a temp file for writing
		if newLockFile, err := os.OpenFile(newLockFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600); err == nil {
			// start scanning file and copying contents
			lockScanner := bufio.NewScanner(lockFile)

			for lockScanner.Scan() {
				variableTokens := strings.SplitN(lockScanner.Text(), " ", 2)
				
				fileLockPID := variableTokens[0]

				if fileLockPID != lockPID {
					logger.Debugf("Found PID %s for writing ...", fileLockPID)
					if bytesWritten, err := newLockFile.WriteString(lockScanner.Text()+"\n"); err != nil {
						filelock.UnlockFile(lockFileName)
						logger.Errorf("Unable to write to new lock file %s", newLockFileName)
					} else {
						fileWriteSize+=bytesWritten
						logger.Debugf("Written %d bytes to new lock file, %d written so far", bytesWritten, fileWriteSize)
					}
				} else {
					logger.Debugf("Ignoring PID %s ...", fileLockPID)
				}
			}

			newLockFile.Close()
			logger.Debug("New lock file closed")
		} else {
			filelock.UnlockFile(lockFileName)
			logger.Errorf("Unable to create temp lock file %s", newLockFileName)
		}

		lockFile.Close()
		logger.Debug("Old lock file closed")
	} else {
		logger.Warnf("Unable to open the lock file %s. Must have been deleted", lockFileName)
	}

	// Check file exists and if it does replace the lock file
	
	if _, err := os.Stat(newLockFileName); err == nil {
		logger.Debugf("Found file %s. Renaming to %s ...", newLockFileName, lockFileName)
		if err := os.Rename(newLockFileName, lockFileName); err != nil {
			filelock.UnlockFile(lockFileName)
			logger.Errorf("Unable to move %s to %s", newLockFileName, lockFileName)
		}
		logger.Debug("Renamed file")
	} else {
		logger.Errorf("Unable to find new lock file %s", newLockFileName)
	}

	logger.Info("Removed entry")

	// Check size of remaining lock file and if 0 then remove it

	if fileWriteSize == 0 {
		logger.Infof("Lock file now empty - removing ...")
		if err := os.Remove(lockFileName); err != nil {
			filelock.UnlockFile(lockFileName)
			logger.Errorf("Unable to remove lock file %s", lockFileName)
		}
		logger.Debug("Successfully removed empty lock file")
	} else {
		logger.Debug("Lock file still contains entries.")
	}
	
	// Unlock the file

	filelock.UnlockFile(lockFileName)

	logger.Debug("Process complete")
}

func CleanLockFile(lockFileName string, lockName string, startingEntry int) []string {
	logger.Info("Cleaning lock file of dead processes ...")

	var lockPIDS []string

	lockCount := 0
	lineCount := 0

	if lockFile , err := os.Open(lockFileName); err == nil {
		lockScanner := bufio.NewScanner(lockFile)

		for lockScanner.Scan() {
			variableTokens := strings.SplitN(lockScanner.Text(), " ", 2)

			lockPID       := variableTokens[0]

			if lockPID != setup.CurrentPID {
				ilockPID, _   := strconv.Atoi(variableTokens[0])
				fileLockName := variableTokens[1]

				if fileLockName == lockName {
					logger.Debugf("Found PID with lock name %s. Checking PID %d is still alive ...", fileLockName, ilockPID)

					if pidAlive, pidIsName := utils.CheckProcess(ilockPID, setup.BaseName); pidAlive {
						if pidIsName {
							logger.Infof("Process %d is running %s.  Valid entry", ilockPID, setup.BaseName)
						} else {
							if lineCount >= startingEntry {
								logger.Warnf("Process %d is not running %s. Invalid entry. Removing ...", ilockPID, setup.BaseName)

							
								lockPIDS = append(lockPIDS,lockPID)
								lockCount++
							} else {
								logger.Warnf("Process %d is not running %s. But is at line %d in lock file. Keeping ...", ilockPID, setup.BaseName)
							}
						}
					} else {
						logger.Warnf("Old PID %d found in lock file and is no longer running. Removing ...", ilockPID)

						lockPIDS = append(lockPIDS,lockPID)
						lockCount++
					}
				} else {
					logger.Debugf("Found PID %d for lock name %s", ilockPID, fileLockName)
				}
			} else {
				logger.Debug("PID found is current PID.  Ignoring ...")
			}

			lineCount++
		}

		lockFile.Close()
		logger.Debug("Closed lock file")
	} else {
		logger.Infof("Lock file not longer present")
	}

	// Remove any PIDs found

	for lockCounter := 0; lockCounter < lockCount; lockCounter++ {
		logger.Debugf("Removing %s from lock file", lockPIDS[lockCounter])
		RemoveLockEntry(lockFileName, lockPIDS[lockCounter])
	}

	logger.Info("Process complete")

	return lockPIDS
}

func AddLockEntry(lockFileName, pid, lockName string) {
	logger.Debug("Adding lock entry ...")

	// To write the file - take a real lock
	
	filelock.LockFile(lockFileName,1)

	if lockFile, err := os.OpenFile(lockFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600); err == nil {
		writeString := strings.Join( []string{ pid, " ", lockName }, "")
	
		if _, err := lockFile.WriteString(writeString+"\n"); err != nil {
			filelock.UnlockFile(lockFileName)
			logger.Errorf("Unable to write to lockfile %s", lockFileName)
		} else {
			logger.Infof("Added entry %s to lock file", writeString)
		}

		lockFile.Close()
		logger.Debug("Closed lock file")
	} else {
		filelock.UnlockFile(lockFileName)
		logger.Errorf("Unable to open file %s", lockFileName)
	}
		
	// Unlock the file

	filelock.UnlockFile(lockFileName)

	logger.Debug("Process complete")
}
