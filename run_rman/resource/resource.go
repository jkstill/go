package resource

// Standard imports

import "bufio"
import "os"
import "strconv"
import "strings"
import "time"

// Local imports

import "github.com/daviesluke/filelock"
import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"
import "github.com/daviesluke/run_rman/config"
//import "github.com/daviesluke/mitchellh/go-ps"


// local Variables

// Local functions

func getResource ( resourceName string, resourceValue int, timeOutMins int) {
	logger.Info("Alloctaing resource ...")

	var imaxResource  int
	var iusedResource int
	var err           error

	// First check there is a resource file present

	if _, err = os.Stat(setup.ResourceFileName); err != nil {
		logger.Errorf("Unable to find resource file %s", setup.ResourceFileName)
	}

	maxResource :=  utils.LookupFile(setup.ResourceFileName, resourceName, 1, 2, ":", 1)

	if maxResource == "" {
		logger.Errorf("Resource %s not found in file %s", resourceName, setup.ResourceFileName)
	}

	if imaxResource, err = strconv.Atoi(maxResource); err != nil {
		logger.Errorf("Resource %s not configured properly in %s with value %s", resourceName, setup.ResourceFileName, maxResource)
	}

	// Check value is not over the maximum allowed

	if resourceValue > imaxResource {
		logger.Errorf("Resource %s has maximum value %d, attempting to get %d", resourceName, imaxResource, resourceValue)
	}

	resourceCounter   := 0
	remainingResource := resourceValue
	
	for allocatedResource := 0; resourceValue - allocatedResource > 0; {
		// Check to see of any resources are currently being used 

		iusedResource  = 0

		// Check if used file exists and clean it 

		if _, err = os.Stat(setup.ResourceUsageFileName); err == nil {
			// Clean up resource file just in case there are old entries

			cleanResources()
		}

		// Lock the usage file to prevent anyone else using the file

		filelock.LockFile(setup.ResourceUsageFileName)
	
		// Check again as after the clean the file may have been removed

		if _, err = os.Stat(setup.ResourceUsageFileName); err == nil {
			usedCounter   := 1

			// Loop through used file looking for usage 

			for usedResource := utils.LookupFile(setup.ResourceUsageFileName, resourceName, 1, 2, ":", usedCounter); usedResource != ""; usedCounter++ {
				if usedAmount, err := strconv.Atoi(usedResource); err != nil {
					filelock.UnlockFile(setup.ResourceUsageFileName)
					logger.Errorf("Resource %s not configured properly in %s with value %s", resourceName, setup.ResourceUsageFileName, usedResource)
				} else {
					iusedResource+=usedAmount
				}
			}
		}

		freeResource := imaxResource - iusedResource

		if freeResource == 0 {
			logger.Infof("No resources of type %s available", resourceName)

			resourceCounter++
		} else {
			if remainingResource <= freeResource {
				logger.Infof("Allocating all needed resources for %s", resourceName)

				addResource(resourceName, remainingResource)

				allocatedResource += remainingResource
				remainingResource = 0
			} else {
				logger.Infof("Allocating partitally needed resources for %s", resourceName)

				addResource(resourceName, freeResource)

				allocatedResource += freeResource
				remainingResource -= freeResource
				resourceCounter++
			}
		} 

		// Unlock the usage file

		filelock.UnlockFile(setup.ResourceUsageFileName)

		if remainingResource > 0 {
			if resourceCounter > timeOutMins {
				logger.Warnf("Timed Out!")
				releaseResources(setup.ResourceObtainedFileName)
				logger.Errorf("Unable to obtain %d units for resource %s", resourceValue, resourceName)
			}

			logger.Info("Resource allocation incomplete.  Sleeping for 60 secs ...")
			time.Sleep(60 * time.Second)
		}
	}

	logger.Info("Process complete")
}

func addResource(resourceName string, resourceValue int) {
	logger.Info("Writing resource file ...")

	// File is already locked when reading and adding entries so do not need to lock again

	// So we can open the file for writing

	// Write string to both files

	writeString := strings.Join( []string{ resourceName, ":", strconv.Itoa(resourceValue), "\n" }, "")

	if resourceUsageFile , err := os.OpenFile(setup.ResourceUsageFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600); err == nil {
		if _, err := resourceUsageFile.WriteString(writeString); err != nil {
			filelock.UnlockFile(setup.ResourceUsageFileName)
			logger.Errorf("Unable to write to resource file %s", setup.ResourceUsageFileName)
		} else {
			logger.Infof("Added entry %s to resource file", writeString)
		}

		resourceUsageFile.Close()
		logger.Debug("Closed resource usage file")
	}

	if resourceObtainedFile , err := os.OpenFile(setup.ResourceObtainedFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600); err == nil {
		if _, err := resourceObtainedFile.WriteString(writeString); err != nil {
			filelock.UnlockFile(setup.ResourceUsageFileName)
			logger.Errorf("Unable to write to resource file %s", setup.ResourceObtainedFileName)
		} else {
			logger.Infof("Added entry %s to resource file", writeString)
		}

		resourceObtainedFile.Close()
		logger.Debug("Closed resource usage file")
	}

	logger.Info("Process complete")
}

func releaseResources(resFileName string) {
	logger.Info("Releasing resources ...")

	// Need to get file lock to affect these files

	// Lock the usage file to prevent anyone else using the file

	filelock.LockFile(setup.ResourceUsageFileName)

	// Open up the file used for cleaning

	if resFile, err := os.Open(resFileName); err == nil {
		// Set up a scanner 

		resScanner := bufio.NewScanner(resFile)

		for resScanner.Scan() {
			resString := resScanner.Text();

			removeUsedResource(resString)
		}

		resFile.Close()
	} else {
		logger.Warnf("Unable to open file %s.  Nothing to do", resFileName)
	}

	// Unlock the usage file

	filelock.UnlockFile(setup.ResourceUsageFileName)

	// Now can remove the file 

	if err := os.Remove(resFileName); err != nil {
		logger.Errorf("Unable to remove file %s", resFileName)
	}
	
	logger.Info("Process complete")
}

func removeUsedResource(resString string) {
	logger.Info("Removing used resource value ...")

	// Create a temporary file 

	tempFileName := strings.Join( []string { setup.ResourceUsageFileName, ".", setup.CurrentPID }, "")

	tempFile , err := os.OpenFile(tempFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)

	if err != nil {
		filelock.UnlockFile(setup.ResourceUsageFileName)
		logger.Errorf("Unable to create temporary file %s", tempFileName)
	}
	
	missedWrite := false

	// Open up the used file

	if usedFile, err := os.Open(setup.ResourceUsageFileName); err == nil {
		// Set up scanner 

		usedScanner := bufio.NewScanner(usedFile)

		for usedScanner.Scan() {
			usedString := usedScanner.Text()

			if missedWrite || usedString != resString {
				if _, err := tempFile.WriteString(usedString+"\n"); err != nil {
					filelock.UnlockFile(setup.ResourceUsageFileName)
					logger.Errorf("Unable to write file %s", tempFileName)
				}
			} else {
				missedWrite = true
			}
		}
	
		usedFile.Close()
	}

	tempFile.Close()

	if err := os.Rename(tempFileName, setup.ResourceUsageFileName); err != nil {
		filelock.UnlockFile(setup.ResourceUsageFileName)
		logger.Errorf("Unable to rename file %s to %s", tempFileName, setup.ResourceUsageFileName)
	}

	if usageInfo, err := os.Stat(setup.ResourceUsageFileName); err == nil {
		if usageInfo.Size() == 0 {
			logger.Info("Usage file is empty. Deleting file ...")
			if err := os.Remove(setup.ResourceUsageFileName); err != nil {
				filelock.UnlockFile(setup.ResourceUsageFileName)
				logger.Errorf("Unable to remove used file %s", setup.ResourceUsageFileName)
			}
		}
	} else {
		filelock.UnlockFile(setup.ResourceUsageFileName)
		logger.Errorf("Unable to find file %s", setup.ResourceUsageFileName)
	}
	
	logger.Info("Process complete")
}

func cleanResources() {
	logger.Info("Cleaning resources file ...")

	logger.Info("Process complete")
}

// Global functions

func GetResources ( resources map[string]int ) {
	logger.Info("Getting resources ...")

	// Loop through resources 

	resourceCount := 0

	config.SetConfig(setup.Database, "CheckResourceMins")

	checkResourceMins, _ := strconv.Atoi(config.ConfigValues["CheckResourceMins"])

	for resourceName, resourceValue := range resources {
		logger.Infof("Checking resource %s, attempting to allocate %d units ...", resourceName, resourceValue)

		getResource(resourceName, resourceValue, checkResourceMins)
	
		resourceCount++
	}

	if resourceCount == 0 {
		logger.Info("No resources to provision")
	}

	logger.Info("Process complete")
}
