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

func LockFile (fileName string) {
	logger.Infof("Putting lock on file %s ...", fileName)

	lockName := strings.Join( []string{ fileName, "locker" }, ".")
	logger.Debugf("Lock file name set to %s", lockName)

	loopCount := 0
	logger.Trace("Initialize loop counter")

        for {
		// O_EXCL is the key - cannot create file if already there 

                fdlock , err := os.OpenFile(lockName , os.O_CREATE | os.O_WRONLY | os.O_EXCL, 0600 )

                if err == nil {
                        logger.Debug("Lock successfully taken")
                        break
                } else {
                        logger.Debugf("Unable to obtain file lock\n")
			loopCount++
			logger.Tracef("Increment loop count, current value %d", loopCount)

			if loopCount < 10 {
				logger.Debug("Sleeping for 100 ms ...")
				time.Sleep(100 * time.Millisecond)
			} else {
				logger.Errorf("Unable to lock file for %d loops.  Exiting ...", loopCount)
			}
                }

		logger.Debug("Closing locker file")
		fdlock.Close()
        }

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
