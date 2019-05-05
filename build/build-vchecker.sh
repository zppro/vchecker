#!/bin/bash
sh ./build/docker-build.sh -c vchecker -o darwin -a amd64
sh ./build/docker-build.sh -c vchecker -o linux -a amd64
