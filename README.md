# go

Go Installation 

1) Get download from https://golang.org/dl/
2) See documentation for installation -> https://golang.org/doc/install

For linux - quick guide
1) Untar into -> tar -C /usr/local -xzf go1.11.linux-amd64.tar.gz
2) PATH=$PATH:/usr/local/go/bin


run_rman Installation

Default GOPATH is ~/go

So to install for default is 

1) mkdir ~/go/src/github.com/daviesluke
2) mkdir ~/go/bin
3) cd ~/go/src/github.com/daviesluke
4) git clone https://github.com/daviesluke/go
4a) Follow oracle setup
5) cd ~/go/src/github.com/daviesluke/run_rman/run_rman
6) go install
7) ~/go/bin/run_rman

For installtion to non-default location

1) Set environment GOPATH to fully qualified directory path e.g. /home/luke/testing/go
2) mkdir -p $GOPATH/src/github.com/daviesluke
3) mkdir -p $GOPATH/bin
4) cd $GOPATH/src/github.com/daviesluke
5) git clone https://github.com/daviesluke/go
5a) Follow oracle setup
6) cd $GOPATH/src/github.com/daviesluke/run_rman/run_rman
7) go install
8) PATH=$GOPATH/bin:$PATH
9) run_rman 


Oracle Setup

1) Set up oci8.pc file with following contents - changing ORACLE_HOME and version as required

ORACLE_HOME=/u01/app/oracle/product/18.0.0.0/db1
Name: oci8
Description: Oracle Call Interface
Version: 18.0
Cflags: -I${ORACLE_HOME}/rdbms/public
Libs: -L${ORACLE_HOME}/lib -Wl,-rpath,${ORACLE_HOME}/lib -lclntsh

2) Set env variable PKG_CONFIG_PATH to directory containing the oci8.pc file
3) go get daviesluke/mattn/go-oci8

