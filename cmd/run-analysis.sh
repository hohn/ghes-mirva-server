#!/bin/bash -x -e
#*    Minimal setup to run analysis using information provided by a request

#*    Take output saved by the server
QUERYPACKID=$1
shift
QUERYLANGUAGE=$1
shift

# and
DBOWNER=$1
shift
DBREPO=$1

GMSROOT=/Users/hohn/local/ghes-mirva-server

#* Set up derived paths
DBPATH=$GMSROOT/var/codeql/dbs/$DBOWNER/$DBREPO
DBZIP=$GMSROOT/codeql/dbs/$DBOWNER/$DBREPO/${DBOWNER}_${DBREPO}_db.zip
DBEXTRACT=$GMSROOT/var/codeql/dbs/$DBOWNER/$DBREPO

QUERYPACK=$GMSROOT/var/codeql/querypacks/qp-$QUERYPACKID.tgz
QUERYEXTRACT=$GMSROOT/var/codeql/querypacks/qp-$QUERYPACKID

QUERYOUTD=$GMSROOT/var/codeql/sarif/localrun/$DBOWNER/$DBREPO
QUERYOUTF=$QUERYOUTD/${DBOWNER}_${DBREPO}.sarif

#*    Prep work before running the command

#**        Extract database
mkdir -p  $DBEXTRACT && cd $DBEXTRACT
unzip -o -q $DBZIP
DBINFIX=`\ls | head -1`                   # Could be cpp, codeql_db, or whatever

#   Extract query pack
mkdir -p $QUERYEXTRACT && cd $QUERYEXTRACT
tar zxf $QUERYPACK

#**        Prepare target directory
mkdir -p $QUERYOUTD

#*    run database analyze
cd $GMSROOT
codeql database analyze --format=sarif-latest --rerun \
       --output $QUERYOUTF \
       -j8 \
       -- $DBPATH/$DBINFIX $QUERYEXTRACT

#*    report result
printf "run-analysis-output in %s\n" $QUERYOUTF
