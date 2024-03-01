#!/bin/bash -x -e

#*    Minimal setup to run analysis using information provided by a request

#*    Take output saved by the server
QUERYPACKID=93522
QUERYLANGUAGE=cpp

# and
DBOWNER=google
DBREPO=flatbuffers

GMSROOT=/Users/hohn/local/ghes-mirva-server

#* Set up derived paths
DBPATH=$GMSROOT/var/codeql/dbs/$DBOWNER/$DBREPO
DBZIP=$GMSROOT/codeql/dbs/$DBOWNER/$DBREPO/${DBOWNER}_${DBREPO}_db.zip
DBEXTRACT=$GMSROOT/var/codeql/dbs/$DBOWNER/$DBREPO

QUERYPACK=$GMSROOT/var/codeql/querypacks/qp-$QUERYPACKID.tgz
QUERYEXTRACT=$GMSROOT/var/codeql/querypacks/qp-$QUERYPACKID

QUERYOUTD=$GMSROOT/var/codeql/sarif/localrun/$DBOWNER/$DBREPO
QUERYOUTF=$QUERYOUTD/${DBOWNER}_${DBREPO}.sarif

#*    Check variable values
sv() {
    printf "%s\t%s\n" $1 ${!1}
}
{
    echo
    sv DBPATH
    sv DBZIP
    sv DBEXTRACT 
    echo
    sv QUERYPACK
    sv QUERYEXTRACT
    sv QUERYOUTD
    sv QUERYOUTF
} 

#*    Prep work before running the command

#**        Extract database
mkdir -p  $DBEXTRACT && cd $DBEXTRACT
unzip -q $DBZIP

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
       -- $DBPATH/$QUERYLANGUAGE $QUERYEXTRACT

#*    report result
printf "output in %s\n" $QUERYOUTF

#*    Manual checks
#**        Check for output
ls -la $QUERYOUTD

#**        compare local output to reference
jq . < $QUERYOUTF > 10-local.sarif
jq . < $GMSROOT/codeql/sarif/reference/$DBOWNER/$DBREPO/${DBOWNER}_${DBREPO}.sarif > 10-reference.sarif
diff 10-*.sarif
