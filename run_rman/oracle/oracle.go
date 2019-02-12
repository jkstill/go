package oracle

// Standard imports

import "database/sql"
import "strings"

// Local imports

import "../../logger"
import "../../setup"
import "../../utils"
import "../../run_rman/config"
import _  "github.com/mattn/go-oci8"

// local variables

// local functions

func checkConnection (connString string) {
	logger.Debug("Checking connection ...")

	logger.Debugf("Connection string -> %s", connString)

	if db, err := sql.Open("oci8", connString); err == nil {

		if err = db.Ping(); err != nil {
			logger.Errorf("Unable to connect to database %s using %s", setup.Database, utils.RemovePassword(connString,false))
		} else {
			logger.Info("Successfully connected to the target database")
		}
		
		db.Close()
	}
	
	logger.Debug("Process complete")
}
	

func checkTargetConnection () {
	logger.Info("Checking target connection ...")

	targetConnection := config.ConfigValues["TargetConnection"]

	if targetConnection == "/" {
		targetConnection = "/@?as=sysdba" // sys/.@?as=sysdba
	} else {
		// If starts with SYS then add as=sysdba
		regEx := "^[Ss][Yy][Ss][@/].+$"

		if utils.CheckRegEx(targetConnection,regEx) {
			// Check if it has a connection part 
			regEx = ".+@.+"

			if utils.CheckRegEx(targetConnection,regEx) {
				targetConnection = strings.Join( []string{ targetConnection , "as=sysdba"} , "?")
			} else {
				targetConnection = strings.Join( []string{ targetConnection , "as=sysdba"} , "@?")
			}
		}
	}

	checkConnection(targetConnection)

	logger.Debug("Process complete")
}

func checkCatalogConnection () {
	logger.Info("Checking catalog connection ...")

	catalogConnection := config.ConfigValues["CatalogConnection"]

	if catalogConnection == "" { 
		logger.Infof("No RMAN catalog has been configured - running with control file only")
	} else {
		checkConnection(catalogConnection)
	}

	logger.Debug("Process complete")
}


// Global functions

func CheckConnections () {
	checkTargetConnection()

	checkCatalogConnection()
}
