# whiteboardcleaner

Application to clean up picture taken from whiteboard / paperboard

*Note: This is a pretty messy prototype at the moment*

## Get

    export GOPATH=`pwd`
    go get github.com/yml/whiteboardcleaner

## Install

    go install github.com/yml/whiteboardcleaner/...
    cd $GOPATH/src/github.com/yml/whiteboardcleaner/assets/js
    npm install
    browserify index.js -t reactify -o bundle.js -v -d

## Run

    cd $GOPATH  # do this so the static assets get picked up
    ./bin/whtbc-server
