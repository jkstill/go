package rman 

// Standard imports

import "os"
import "os/exec"
import "path/filepath"
import "strings"

// Local imports

import "github.com/daviesluke/logger"
import "github.com/daviesluke/setup"
import "github.com/daviesluke/utils"
import "github.com/daviesluke/run_rman/config"
import "github.com/daviesluke/run_rman/general"

// local variables

var ResetConfigFileName string

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

	commandFileName := strings.Join( []string{ setup.BaseName, setup.CurrentPID, "rman" }, ".")
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

	logger.Info("Process complete")
}

func addConnections ( cmdFileName string, targetConn string, catalogConn string ) {
	logger.Info("Adding connections to file ...")

	newCmdFileName := strings.Join( []string{ cmdFileName , "tmp" }, "." )
	logger.Debugf("New command file %s", newCmdFileName)

	if newCmdFile, err :=  os.OpenFile(newCmdFileName, os.O_CREATE | os.O_WRONLY , 0600 ); err == nil {
		// Write the connection strings

		connString := strings.Join( []string{ "connect", "target", targetConn }, " ")

		if _, err := newCmdFile.WriteString(connString + "\n"); err != nil {
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

	logger.Info("Process complete")
}

func runRMAN(cmdFile string, outFile string) {
	logger.Info("Running RMAN ...")

	setup.CopyFileToLog("Command file contents", cmdFile)

	addConnections(cmdFile, config.ConfigValues["TargetConnection"], config.ConfigValues["CatalogConnection"])

	cmdParams := []string{ "cmdfile", cmdFile}

	// Open the output file to capture the stdout and stderr

	out, err := os.OpenFile(outFile, os.O_CREATE | os.O_WRONLY , 0600 )
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

	if err := cmd.Run(); err != nil {
		logger.Errorf("RMAN command failed to run. See %s for details", outFile)
	}

	// Close output file
	out.Close()

	setup.CopyFileToLog("RMAN output", outFile)

	if checkRMANErrors(outFile) {
		logger.Errorf("RMAN ran with errors. Check log %s for deatils", outFile)
	} else {
		logger.Info("RMAN ran without error")
	}

	logger.Info("Process complete")
}

func checkRMANErrors(logFileName string) bool {
	logger.Info("Checking log for errors ...")

	rmanRegEx := ""
	oraRegEx  := ""

	rmanCount := 0 
	oraCount  := 0
	regGroup  := 0

	if config.ConfigValues["RMANIgnoreErrors"] != "" {
		rmanErrors := strings.Split(config.ConfigValues["RMANIgnoreErrors"],";")

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
				logger.Infof("Ignoring error - %s", rmanError)
			} else {
				logger.Warnf("Invalid error string passed %s - must be ORA-99999 or RMAN-99999 errors", rmanError)
			}
		}
	} else {
		logger.Debug("No errors to be ignored")
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


// Global functions

func SaveConfig () {
	logger.Info("Saving RMAN configuration ...")

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

		// Get the current RMAN settings 

		getConfig(setup.TmpFileName)

		// Write the reset file 

		utils.CopyFileContents(setup.TmpFileName, ResetConfigFileName, "^CONFIGURE ")

		// Remove the tmp file

		if err := os.Remove(setup.TmpFileName); err != nil {
			logger.Errorf("Unable to remove file %s", setup.TmpFileName)
		}

		// Add a lock on the config file so that only last process resets it

	} else {
		logger.Warn("Not using a custom RMAN config - relying upon control file entries")
	}

	logger.Info("Process complete")
}
