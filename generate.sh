#!/bin/bash

# generate blockchain
rivinecg generate blockchain

# update the vendor deps
dep ensure -v -update
