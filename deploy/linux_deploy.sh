#!/bin/bash

BaseDir=~/go/src/github.com/daviesluke
DeployDir=${BaseDir}/deploy
ScriptDir=${DeployDir}/scripts

ExeFileName=${ScriptDir}/run_rman
OSType=linux

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

zip -r config scripts rman_scripts $ZipFileName
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

git commit -m "Linux Deploy"
if [[ $? -ne 0 ]]
then
	echo "Error. Unable to commit files for git. Exiting ..."
	exit 1
fi

git push <<!
c6ec4be99ebf88c26cec7c7b984a872b5dc15fd6
!
if [[ $? -ne 0 ]]
then
	echo "Error. Unable to push for git. Exiting ..."
	exit 1
fi

