#!/bin/bash
pwd=`pwd`
cmd="$pwd/cmd/$1"
outfile=`basename $cmd`
echo $outfile
echo $cmd
go build ./cmd/$outfile
