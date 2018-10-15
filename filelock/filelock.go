package filelock

// Standard imports

import "os"
import "strings"
import "time"

// Local imports

import "github.com/daviesluke/logger"

// local Variables

// Global Variables

// Local functions

func LockFile (fileName string, lockDuration int) {
	logger.Infof("Putting lock on file %s ...", fileName)

	lockName := strings.Join( []string{ fileName, "locker" }, ".")
	logger.Debugf("Lock file name set to %s", lockName)

	loopCount := 0
	logger.Trace("Initialize loop counter")

	var fdlock *os.File
	var err    error

	var sleepDuration time.Duration
	var maxCount      int

	// Set some sensible defaults for waiting
	if lockDuration < 10 {
		sleepDuration = 100 * time.Millisecond
		maxCount      = 10  * lockDuration
	} else {
		sleepDuration = 1 * time.Second
		maxCount      = lockDuration
	}

        for {
		// O_EXCL is the key - cannot create file if already there 

                fdlock , err = os.OpenFile(lockName , os.O_CREATE | os.O_WRONLY | os.O_EXCL, 0600 )

                if err == nil {
                        logger.Debug("Lock successfully taken")
                        break
                } else {
                        logger.Debugf("Unable to obtain file lock\n")
			loopCount++
			logger.Tracef("Increment loop count, current value %d", loopCount)

			if loopCount < maxCount {
				logger.Debug("Sleeping ...")
				time.Sleep(sleepDuration)
			} else {
				logger.Errorf("Unable to lock file for %d loops.  Exiting ...", loopCount)
			}
                }
        }

	// Closing the locker file

	logger.Debug("Closing locker file")
	fdlock.Close()

	logger.Info("Process complete")
}

func UnlockFile(fileName string) {
	logger.Info("Unlocking file ...")

	lockName := strings.Join( []string{ fileName, "locker" }, ".")
	logger.Debugf("Lock file name set to %s", lockName)

        if err := os.Remove(lockName); err == nil {
		logger.Info("File unlocked")
        } else {
		logger.Errorf("Unable to unlock the file.  Exiting ...")
	}
	
	logger.Info("Process complete")
}
