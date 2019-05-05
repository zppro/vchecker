#!/bin/bash
sh ./build/docker-build.sh -c vpublisher -o darwin -a amd64
sh ./build/docker-build.sh -c vpublisher -o linux -a amd64
