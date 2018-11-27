package rman 

// Standard imports

import "bufio"
import "fmt"
import "os"
import "os/exec"
import "path/filepath"
import "strconv"
import "strings"

// Local imports

import "github.com/daviesluke/filelock"
import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"
import "github.com/daviesluke/run_rman/config"
import "github.com/daviesluke/run_rman/general"
import "github.com/daviesluke/run_rman/locker"

// local variables

var ResetConfigFileName     string
var ResetConfigLockFileName string

// local functions

func checkDir(dirName string) {
	logger.Debug("Checking scripts directory exists ...")

	if dirInfo, err := os.Stat(dirName); err == nil && dirInfo.IsDir() {
		logger.Debugf("Directory %s already exists", dirName)
	} else {
		if err != nil {
			if err := os.MkdirAll(dirName,0755); err != nil {
				logger.Errorf("Unable to make directory %s", dirName)
			}
		} else {
			logger.Errorf("File %s is not a directory", dirName)
		}
	}

	logger.Debug("Process complete")
}

func getConfig(outputFile string) {
	logger.Info("Getting RMAN configuration ...")

	checkDir(setup.RMANScriptDir)

	commandFileName := strings.Join( []string{ setup.BaseName, setup.CurrentPID }, ".")
	commandFileName = filepath.Join( setup.RMANScriptDir , commandFileName)
	
	// Write the command file

	if commandFile, err := os.OpenFile(commandFileName, os.O_CREATE | os.O_WRONLY | os.O_EXCL, 0600 ); err == nil {
		if _, err := commandFile.WriteString("show all;\n") ; err != nil {
			logger.Errorf("Unable to write file %s", commandFileName)
		}

		commandFile.Close()
	} else {
		logger.Errorf("Unable to open the file %s", commandFileName)
	}

	runRMAN(commandFileName, outputFile)

	// Don't need the command file now so can remove it 

	if err := os.Remove(commandFileName); err != nil {
		logger.Errorf("Unable to remove file %s", commandFileName)
	}

	logger.Debug("Process complete")
}

func addConnections ( cmdFileName string, targetConn string, catalogConn string ) {
	logger.Infof("Adding connections to file %s ...", cmdFileName)

	newCmdFileName := strings.Join( []string{ cmdFileName , "tmp" }, "." )
	logger.Debugf("New command file %s", newCmdFileName)

	if newCmdFile, err :=  os.OpenFile(newCmdFileName, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0600 ); err == nil {
		// Write the connection strings

		connString := strings.Join( []string{ "connect", "target", targetConn }, " ")

		if _, err := newCmdFile.WriteString(connString + setup.NewLine); err != nil {
			logger.Errorf("Unable to write target connection to file %s", newCmdFileName)
		}

		logger.Debug("Target connection written")

		if catalogConn != "" {
			connString = strings.Join( []string{ "connect", "catalog", catalogConn }, " ")

			if _, err := newCmdFile.WriteString(connString + "\n"); err != nil {
				logger.Errorf("Unable to write catalog connection to file %s", newCmdFileName)
			}
			
			logger.Debug("Catalog connection written")
		}

		newCmdFile.Close()

		utils.CopyFileContents(cmdFileName, newCmdFileName, "")

		if err := os.Rename(newCmdFileName, cmdFileName); err != nil {
			logger.Errorf("Unable to rename file from %s to %s", newCmdFileName, cmdFileName)
		}
		logger.Debugf("Renamed file %s to %s", newCmdFileName, cmdFileName)
	} else {
		logger.Errorf("Unable to open file %s", newCmdFileName)
	}

	logger.Debug("Process complete")
}

func runRMAN(cmdFile string, outFile string) {
	logger.Info("Running RMAN ...")

	setup.CopyFileToLog("Command file contents", cmdFile)

	addConnections(cmdFile, config.ConfigValues["TargetConnection"], config.ConfigValues["CatalogConnection"])

	cmdParams := []string{ "cmdfile", cmdFile}

	// Open the output file to capture the stdout and stderr

	out, err := os.OpenFile(outFile, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0600 )
	if err != nil {
		logger.Errorf("Unable to open RMAN output file %s", outFile) 
	}

	// Set the command (does not run it yet)

	cmd := exec.Command(general.RMAN , cmdParams ...)
	logger.Debugf("Set command to run %s cmdfile %s", general.RMAN, cmdFile)
	
	// Setting the stdout and stderr 

	cmd.Stdout = out
	cmd.Stderr = out 

	// Now let's run it 

	rmanErr := cmd.Run()

	// Close output file
	out.Close()

	setup.CopyFileToLog("RMAN output", outFile)

	if rmanErr != nil {
		logger.Error("RMAN command failed to run. See log for details")
	}

	if checkRMAN(outFile) {
		logger.Error("RMAN ran with errors. Check log for deatils")
	} else {
		logger.Info("RMAN run successful")
	}
}

func checkRMAN(logFileName string) bool {
	logger.Info("Checking log for failure messages ...")

	rmanRegEx := ""
	oraRegEx  := ""

	rmanCount := 0 
	oraCount  := 0
	regGroup  := 0

	if config.ConfigValues["RMANIgnoreCodes"] != "" {
		rmanErrors := strings.Split(config.ConfigValues["RMANIgnoreCodes"],";")

		for _, rmanError := range rmanErrors {
			// Split each error
			errbd := strings.SplitN(rmanError,"-",2)

			if utils.CheckRegEx(errbd[0],"^(ORA|RMAN)$") && utils.CheckRegEx(errbd[1],"^([0-9]{5})$") {
				if errbd[0] == "RMAN" {
					if rmanCount == 0 {
						rmanRegEx = strings.Join( []string{ rmanRegEx, "(" }, "")
						regGroup++
					} else {
						rmanRegEx = strings.Join( []string{ rmanRegEx, "|" }, "")
					}
					rmanCount++

					rmanRegEx = strings.Join( []string{ rmanRegEx, rmanError }, "")
				} else if errbd[0] == "ORA" {
					if oraCount == 0 {
                                                oraRegEx = strings.Join( []string{ oraRegEx, "(" }, "")
						regGroup++
                                        } else {
						oraRegEx = strings.Join( []string{ oraRegEx, "|" }, "")
					}

                                        oraCount++

                                        oraRegEx = strings.Join( []string{ oraRegEx, rmanError }, "")
				}
				logger.Infof("Ignoring failure code - %s", rmanError)
			} else {
				logger.Warnf("Invalid failure string passed %s - must be ORA-99999 or RMAN-99999 type codes", rmanError)
			}
		}
	} else {
		logger.Debug("No failures to be ignored")
	}

	ignoreRegEx := ""

	if rmanCount > 0 {
		rmanRegEx = strings.Join( []string{ rmanRegEx, ")" }, "")
		ignoreRegEx = rmanRegEx
	}

	if oraCount > 0 {
		oraRegEx = strings.Join( []string{ oraRegEx, ")" }, "")
		ignoreRegEx  = strings.Join( []string{ ignoreRegEx , oraRegEx }, "|" )
	}

	regEx       := "(RMAN-[0-9]{5}|ORA-[0-9]{5})" 

	logger.Debugf("Regular expression set to %s, ignoring %s, ignore groups %d", regEx, ignoreRegEx, regGroup)

	return utils.FindInFile(logFileName,regEx,ignoreRegEx,regGroup) 
}

func saveConfig (newConfigFileName string) {
	logger.Info("Saving RMAN configuration ...")

	// Get the current RMAN settings 

	getConfig(setup.TmpFileName)

	// Write the reset file 

	utils.CopyFileContents(setup.TmpFileName, newConfigFileName, "^CONFIGURE ")

	// Remove the tmp file

	if err := os.Remove(setup.TmpFileName); err != nil {
		logger.Errorf("Unable to remove file %s", setup.TmpFileName)
	}

	logger.Debug("Process complete")
}

func formatCommand ( oldCmdFile, newCmdFile string ) {
	logger.Info("Adding in substitution strings to RMAN command file ...")

	// Open up both files 

	oldCmd, err := os.Open(oldCmdFile)
	if err != nil {
		logger.Errorf("Unable to open command file %s for reading", oldCmdFile)
	}

	defer oldCmd.Close()

	newCmd, err := os.OpenFile(newCmdFile, os.O_CREATE | os.O_WRONLY | os.O_TRUNC , 0600 )
	if err != nil {
		logger.Errorf("Unable to open command file %s for writing", newCmdFile)
	}

	defer newCmd.Close()

	// Set up the scanner 

	oldScan := bufio.NewScanner(oldCmd)

	for oldScan.Scan() {
		// First replace the <format>

		cmdLine := utils.ReplaceString(oldScan.Text(),"<format>",config.ConfigValues["FileFormat"])

		// Next do the <parallel>

		ParallelSlaves, err := strconv.Atoi(config.ConfigValues["ParallelSlaves"])
		if err != nil {
			logger.Errorf("ParallelSlaves configuration is not an integer")
		}

		parallelCmd := ""

		for i := 0; i < ParallelSlaves; i++ {
			singleCmd := fmt.Sprintf("ALLOCATE CHANNEL C%d DEVICE TYPE %s;", i, config.ConfigValues["ChannelDevice"])

			if i != 0 {
				parallelCmd = strings.Join( []string{ parallelCmd, "\n" }, "" )
			}

			parallelCmd = strings.Join( []string{ parallelCmd, singleCmd }, "" )
		}

		cmdLine = utils.ReplaceString(cmdLine, "<parallel>", parallelCmd)

		if _, err := newCmd.WriteString(cmdLine+"\n"); err != nil {
			logger.Errorf("Unable to write to new command file %s", newCmdFile)
		}
	}

	newCmd.Sync()
	
	logger.Debug("Process complete")
}

func setConfig( oldConfig , newConfig string ) {
	logger.Infof("Old configuration -> %s", oldConfig)
	logger.Infof("New configuration -> %s", newConfig)

	// Open new files to compare and write out any differences

	new, err := os.Open(newConfig)
	if err != nil {
		logger.Errorf("Unable to open new config file %s", newConfig)
	}

	defer new.Close()

	newCmdFile := strings.Join( []string{ newConfig, "run" }, "." )

	out, err := os.OpenFile(newCmdFile, os.O_CREATE | os.O_WRONLY | os.O_TRUNC , 0600 )
	if err != nil {
		logger.Errorf("Unable to open new command file %s", newCmdFile)
	}

	newScan := bufio.NewScanner(new)

	totalWritten := 0

	for newScan.Scan() {
		found := false

		old, err := os.Open(oldConfig)
		if err != nil {
			logger.Errorf("Unable to open old config file %s", oldConfig)
		}

		oldScan := bufio.NewScanner(old)

		for oldScan.Scan() {
			if newScan.Text() == oldScan.Text() {
				logger.Debugf("Found command %s in old file. Ignoring ...", newScan.Text())
				found = true
				break
			}
		}

		old.Close()

		if ! found {
			logger.Debugf("Writing config entry %s", newScan.Text())

			if bytesWritten, err := out.WriteString(newScan.Text()+"\n"); err != nil {
				logger.Errorf("Unable to write file %s", setup.TmpFileName)
			} else {
				totalWritten += bytesWritten
			}
		}
	}

	out.Sync()

	out.Close()

	// Now let's run what's left
	
	if totalWritten > 0 {
		runRMAN(newCmdFile,setup.TmpFileName)

		// No need for the output 

		if err := os.Remove(setup.TmpFileName); err != nil {
			logger.Errorf("Unable to remove output file %s",setup.TmpFileName)
		}
	} else {
		logger.Info("No changes to be made")
	}

	// May remove the command file

	if err := os.Remove(newCmdFile); err != nil {
		logger.Errorf("Unable to remove command file %s",newCmdFile)
	}

	logger.Debug("Process complete")
}

func removeLockEntry(lockFile string, lockPID string, resetFile string) {
	logger.Debug("Removing lock entry and associated file ...")

	locker.RemoveLockEntry(lockFile, lockPID)

	// Remember to remove the associated file 

	if err := os.Remove(resetFile); err != nil {
		logger.Errorf("Unable to remove old config file %s", resetFile)
	}

	logger.Debug("Process complete")
}

// Global functions

func CheckConfig () {
	logger.Info("Checking RMAN configuration ...")

	if config.ConfigValues["RMANConfig"] != "" {
		// Check file exists

		if _, err := os.Stat(config.ConfigValues["RMANConfig"]); err != nil {
			logger.Errorf("Unable to open RMAN Config file %s", config.ConfigValues["RMANConfig"])
		}

		// Get base directory for the config file so we can create a lock on the process

		baseRMANDir            := filepath.Dir(config.ConfigValues["RMANConfig"])
		baseRMANConfigFileName := filepath.Base(config.ConfigValues["RMANConfig"])

		ResetConfigFileName = strings.Join( []string{ baseRMANConfigFileName, setup.CurrentPID, "reset"}, ".")
		ResetConfigFileName = filepath.Join(baseRMANDir, ResetConfigFileName)

		ResetConfigLockFileName = strings.Join( []string{ baseRMANConfigFileName, "lock" }, ".")
		ResetConfigLockFileName = filepath.Join(baseRMANDir, ResetConfigLockFileName)
		
		// Put a lock on the lock file during the save so that we have list in time order of when the file as used 
		// Have to wait for longer than typical to allow for show all to run - allowing 20 secs

		filelock.LockFile(config.ConfigValues["RMANConfig"],20)

		// See if file exists
		_, err := os.Stat(ResetConfigLockFileName)

		locker.AddLockEntry(ResetConfigLockFileName,setup.CurrentPID,"0")

		saveConfig(ResetConfigFileName)

		filelock.UnlockFile(config.ConfigValues["RMANConfig"])

		// If previous check satted that file did not exist then set the config to the new configj
		if err != nil {
			setConfig(ResetConfigFileName, config.ConfigValues["RMANConfig"])
		}
	} else {
		logger.Warn("Not using a custom RMAN config - relying upon control file entries")
	}

	logger.Debug("Process complete")
}

func RunScript () {
	logger.Info("Running main RMAN script ...")

	// First have to substitute some variables

	newCommandFile := strings.Join( []string{ config.RMANScript, setup.CurrentPID }, ".")

	formatCommand(config.RMANScript, newCommandFile)

	// Set the NLS_DATE_FORMAT for better output 

	if config.ConfigValues["NLS_DATE_FORMAT"] == "" {
		logger.Warnf("NLS_DATE_FORMAT is not set.  Time will not be recorded. This is not recommended")
	}

	os.Setenv("NLS_DATE_FORMAT", config.ConfigValues["NLS_DATE_FORMAT"])

	runRMAN(newCommandFile,setup.TmpFileName)

	// Do not need the log or command file

	if err := os.Remove(newCommandFile); err != nil {
		logger.Errorf("Unable to remove command file %s", newCommandFile)
	}

	if err := os.Remove(setup.TmpFileName); err != nil {
		logger.Errorf("Unable to remove output file %s", setup.TmpFileName)
	}

	logger.Info("Process complete")
}

func ResetConfig () {
	logger.Info("Reset the configuration ...")

	// We should only reset the config if the process is the last one using that specific config file
	// So we need to check the lock file and make sure that if we are the last one in the list 
	// then we do the reset

	if config.ConfigValues["RMANConfig"] != "" {
		
		// Set some variables

		baseDir       := filepath.Dir(config.ConfigValues["RMANConfig"])
		baseFileName  := filepath.Base(config.ConfigValues["RMANConfig"])

		// First lets clear out any dead processes in the lock file (except the first entry)

		lockPIDS := locker.CleanLockFile(ResetConfigLockFileName, "0", 1)

		// Remove any associated files 

		for _, lockPID := range lockPIDS {
			resetFileName := strings.Join( []string{ baseFileName, lockPID, "reset"}, ".")
			resetFileName = filepath.Join(baseDir, resetFileName)
			
			if _, err := os.Stat(resetFileName); err != nil {
				logger.Warnf("File %s has already been removed", resetFileName)
			} else {
				if err := os.Remove(resetFileName); err != nil {
					logger.Errorf("Unable to remove old config file %s", resetFileName)
				}
			}
		}

		// Now check the file - see how many process have started using this config file
		// We only want to reset using the first config file saved

		// Lock up the config to avoid anyone else using it whilst we are checking

		filelock.LockFile(config.ConfigValues["RMANConfig"],20)

		lineCount := utils.CountLines(ResetConfigLockFileName)

		logger.Debugf("%d lines found in file %s", lineCount, ResetConfigLockFileName)

		if lineCount == 0 { 
			logger.Errorf("File %s is missing or empty. Something has gone wrong", ResetConfigLockFileName)
		}

		logger.Debugf("Found %d processes using the config file %s", lineCount, config.ConfigValues["RMANConfig"])

		// Get the first PID in the file

		lockPID := utils.LookupFile(ResetConfigLockFileName, "0", 2, 1, " ", 1)

		logger.Debugf("First PID in file is %s", lockPID)

		// Check the number of processes using the config file

		if lineCount == 1 { 
			// If the line count is one then it must be our process so we can just set it back

			if lockPID != setup.CurrentPID {
				logger.Errorf("Only PID in the file is not our own.  Something has gone wrong. Exiting ...")
			}

			setConfig(config.ConfigValues["RMANConfig"], ResetConfigFileName)

			// Get rid of our entry and files

			removeLockEntry(ResetConfigLockFileName,setup.CurrentPID,ResetConfigFileName)
		} else {
			// This means there is our process and others
			// First of all check the first process is not us

			if lockPID != setup.CurrentPID {

				ilockPID , err := strconv.Atoi(lockPID)
				if err != nil {
					logger.Errorf("Lock PID is corrupted %s, should be a number", lockPID)
				}

				logger.Debugf("Checking PID %d to see if is running ...", ilockPID)

				pidAlive, pidIsName := utils.CheckProcess(ilockPID, setup.BaseName)

				if pidAlive && pidIsName {
					logger.Warnf("Process %d is still running. Will not reset the config", ilockPID)
				} else {
					logger.Debug("Pid is either dead or is not running this process")

					if lineCount == 2 {
						// If the line count for number of processes in lock file is 2
						// then it must be our process and another.
						// So we can remove the process and reset the config

						logger.Warnf("Process %d is not running %s. Cleaning up config", ilockPID, setup.BaseName)

						resetFileName := strings.Join( []string{ baseFileName, lockPID, "reset"}, ".")
						resetFileName = filepath.Join(baseDir, resetFileName)

						setConfig(config.ConfigValues["RMANConfig"], resetFileName)

						removeLockEntry(ResetConfigLockFileName,lockPID,resetFileName)
					} 

				}

				// Then remove our entry as we do not need it here - some other process will reset the config

				removeLockEntry(ResetConfigLockFileName,setup.CurrentPID,ResetConfigFileName)
			} else {
				// First PID is our PID but there are other processes using this config file 
				// So we do not want to remove our entry - the remaining processes will do it

				logger.Warn("Another process found using this configuration file. Leaving entry in lock file")
			}
		} 

		// Unlock the main config file

		filelock.UnlockFile(config.ConfigValues["RMANConfig"])
	}

	logger.Debug("Process complete")
}
