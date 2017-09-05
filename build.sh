#!/bin/bash

SOURCES=/home/sources

if [ ! -d "$SOURCES" ]; then
    mkdir $SOURCES
fi

if [ -d "$SOURCES/qprob_goapi" ]; then
    cd $SOURCES/qprob_goapi
    git fetch --all
    git reset --hard origin/master
else
    cd $SOURCES
    git clone https://github.com/xenu256/qprob_goapi
fi

for PROJECT in bsnssnws entreprnrnws parameterless qprob realestenews stckmrkt webdnl
do
        cp -R $SOURCES/qprob_goapi/api_server/. /home/$PROJECT/api_server/
        cp -R  $SOURCES/qprob_goapi/api_server/. ~/go/src/github.com/xenu256/qprob_goapi/api_server/
        cd /home/$PROJECT/api_server
        /usr/local/go/bin/go build
        chown -R www-data:www-data  /home/$PROJECT/api_server
        stop a$PROJECT
        start a$PROJECT
done
