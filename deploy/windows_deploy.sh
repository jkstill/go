#!/bin/bash

BaseDir=~/go/src/github.com/daviesluke
DeployDir=${BaseDir}/deploy
ScriptDir=${DeployDir}/scripts

ExeFileName=${ScriptDir}/run_rman.exe
OSType=windows

ZipFileName=${DeployDir}/run_rman.${OSType}.zip

SourceDir=${BaseDir}/run_rman/run_rman

# Compile new executable

cd $SourceDir

go build -o $ExeFileName
if [[ $? -ne 0 ]]
then
	echo "Error : Build failed. Exiting ..."
	exit 1
fi

# Create zip file

cd $DeployDir

zip a -tzip $ZipFileName config scripts rman_scripts -r
if [[ $? -ne 0 ]]
then
	echo "Error: Zip failed. Exiting ..."
	exit 1
fi

# Now remove executable to prevent git issues with files with no suffix

rm -f $ExeFileName
if [[ $? -ne 0 ]]
then
	echo "Error. Unable to remove file. Exiting ..."
	exit 1
fi

# Now upload to github

cd $BaseDir

git add . 
if [[ $? -ne 0 ]]
then
	echo "Error. Unable to add files for git. Exiting ..."
	exit 1
fi

git config user.name daviesluke
git config user.email davies@pythian.com

git commit -am "Windows Deploy"
if [[ $? -ne 0 ]]
then
	echo "Error. Unable to commit files for git. Exiting ..."
	exit 1
fi

git push
if [[ $? -ne 0 ]]
then
	echo "Error. Unable to push for git. Exiting ..."
	exit 1
fi

