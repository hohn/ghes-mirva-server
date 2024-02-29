#!/bin/bash

#* Minimal setup to run analysis using information provided by a request

#*    Prep work before running the command
# 
# cd ~/local/ghes-mirva-server
# codeql database analyze --format=sarif-latest --rerun \
#        --output $QUERYOUT \
#        -j8 \
#        -- $DBPATH $QUERYPACK


#**        Extract database
cd `dirname $DBPATH`
unzip -q `basename $DBPATH`
if [ $? -ne 0 ]; then
   echo "DB extraction failed"
fi

#   Extract query pack
cd `dirname $QUERYPACK`
qp=`basename $QUERYPACK`
mkdir ${qp/.tgz/}
cd ${qp/.tgz/}
tar zxf ../$qp
if [ $? -ne 0 ]; then
   echo "query pack extraction failed"
fi
fqp=`pwd`

#**        Prepare target directory
mkdir -p `dirname $QUERYOUT`

#*    Given output saved by the server and the above preparation, 
QUERYPACK=/Users/hohn/local/ghes-mirva-server/querypack-93522.tgz

# and
DBPATH=/Users/hohn/local/ghes-mirva-server/codeql/dbs/google/flatbuffers/google_flatbuffers_db.zip
QUERYOUT=/Users/hohn/local/ghes-mirva-server/codeql/sarif/localrun/google/flatbuffers/google_flatbuffers.sarif

#*    run database analyze
cd ~/local/ghes-mirva-server
codeql database analyze --format=sarif-latest --rerun \
       --output $QUERYOUT \
       -j8 \
       -- `dirname $DBPATH`/cpp $fqp

# Check for output
ls -la $QUERYOUT

#*    manually) Compare local output to reference
jq . < $QUERYOUT > local-1.sarif
jq . < ${QUERYOUT/localrun/reference} > reference-1.sarif
diff local-1.sarif reference-1.sarif
