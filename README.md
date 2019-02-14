# RMAN Framework

## The Golang version

## Installing Golang

Go Installation 

1) Get download from https://golang.org/dl/
2) See documentation for installation -> https://golang.org/doc/install

For linux - quick guide
1) Untar into -> tar -C /usr/local -xzf go1.11.linux-amd64.tar.gz
2) PATH=$PATH:/usr/local/go/bin

## Install dependencies

Three external go packages are to be installed in ~/go

- rlog
  - github.com/pythian/rlog
  - this is a fork of github.com/romana/rlog
- go-ps
  - github.com/mitchellh/go-ps
- oci8
  - github.com/mattn/go-oci8

Right now 2 of them can be installed, and the third can be installed following some setup.

First make sure GOPATH is not set, or is set to the default location of $HOME/go
Note: relative paths are used for the go 'import' statements when importing packages that are local to this project


```bash
unset GOPATH
```

### rlog

```bash
  go get github.com/pythian/rlog
```

### go-ps

```bash
  go get github.com/mitchellh/go-ps
```


## Installing run_rman

Now create a directory for the RMAN Framework project

For example

```bash
  mkdir -p ~/oracle/rman/framework/go
  cd ~/oracle/rman/framework/go
```

Now clone the repo

```bash
  git clone https://github.com/daviesluke/go
```

 Now setup the _mattn/go-oci8/oci8.pc_ file as follows,  changing ORACLE_HOME and version as required.

```text
ORACLE_HOME=/u01/app/oracle/product/18.0.0.0/db1
Name: oci8
Description: Oracle Call Interface
Version: 18.0
Cflags: -I${ORACLE_HOME}/rdbms/public
Libs: -L${ORACLE_HOME}/lib -Wl,-rpath,${ORACLE_HOME}/lib -lclntsh
```

Now you are ready to install oci8.

Set env variable PKG_CONFIG_PATH to directory containing the oci8.pc file

eg.

```bash
export PKG_CONFIG_PATH=/$HOME/pythian/rman/framework/go/mattn/go-oci8
```

Now install go-oci8

```bash
go get github.com/daviesluke/mattn/go-oci8
```

Now you should see the following packages installed:

```bash
>  ls -ld ~/go/pkg/*/*/*/*
-rw-r--r-- 1 jkstill dba  95840 Feb 12 14:50 /home/jkstill/go/pkg/linux_amd64/github.com/pythian/rlog.a
-rw-r--r-- 1 jkstill dba 630512 Feb 12 13:55 /home/jkstill/go/pkg/linux_amd64/github.com/mattn/go-oci8.a
-rw-r--r-- 1 jkstill dba  27260 Feb 12 14:13 /home/jkstill/go/pkg/linux_amd64/github.com/mitchellh/go-ps.a
```

Now build the binary:

```bash

>  ls -l
total 5800
-rw-r--r-- 1 jkstill dba    2163 Feb 12 15:30 run_rman.go

jkstill@poirot  ~/pythian/rman/framework/go/run_rman/run_rman $
>  go build

jkstill@poirot  ~/pythian/rman/framework/go/run_rman/run_rman $
>  ls -l
total 5800
-rwxr-xr-x 1 jkstill dba 5932528 Feb 12 15:49 run_rman
-rw-r--r-- 1 jkstill dba    2163 Feb 12 15:30 run_rman.go

jkstill@poirot  ~/pythian/rman/framework/go/run_rman/run_rman $
>  ./run_rman -h
Usage of ./run_rman:
  -E string
    	E-mail List for success / failure
  -L string
    	Alternative Log directory
  -c string
    	Config File Name
  -config string
    	Config File Name
  -d string
    	Database name
  -db string
    	Database name
  -e string
    	E-mail list for failure
  -email string
    	E-mail list for success / failure
  -erroremail string
    	E-mail list for failure
  -l string
    	Lock name
  -lock string
    	Lock name
  -log string
    	Directory for logs
  -r string
    	Resource name
  -resource string
    	Resource name


```
