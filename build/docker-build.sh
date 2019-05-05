#!/bin/bash
usage()
{
    echo "usage: <command> -c <target> -o <GOOS> -a <GOARCH> (-s)"
}

hashable=""
while getopts "o:a:c:s" arg #选项后面的冒号表示该选项需要参数
do
        case $arg in
             o)
                GOOS="$OPTARG"
                ;;
             a)
                GOARCH="$OPTARG"
                ;;
             c)
                cmd="$OPTARG"
                ;;
             s)
                hashable="yes"
                ;;
             ?)  #当有不认识的选项的时候arg为?
                echo "unkonw argument"
                exit 1
                ;;
        esac
done

if [ -z "$cmd" ]
then
   usage
   exit 1
fi

if [ -z "$GOOS" ]
then
   GOOS="darwin"
fi

if [ -z "$GOARCH" ]
then
   GOARCH="amd64"
fi

echo "build for $GOOS-$GOARCH $cmd..."

project=$(basename `pwd`)
out="$cmd"
docker run -v "$GOPATH":/go --rm -v "$PWD":"/go/src/$project" -w "/go/src/$project" -e GOOS="$GOOS" -e GOARCH="$GOARCH" golang:1.11.6 go build -v -o "$out" ./cmd/$cmd
if [ -n "$hashable" ]
then
    fhash=`shasum -a 256 ./$out`
    mv "$out" "./build/package/$GOOS-$GOARCH/$out-${fhash:0:8}"
    unset fhash
else
    mv "$out" "./build/package/$GOOS-$GOARCH/$out"
fi

