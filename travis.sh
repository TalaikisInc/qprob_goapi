#!/bin/bash

mkdir $TRAVIS_BUILD_DIR/sources
cd $TRAVIS_BUILD_DIR/sources
git clone https://github.com/xenu256/qprob_goapi
cd $TRAVIS_BUILD_DIR/sources/qprob_goapi/api_server
go build
