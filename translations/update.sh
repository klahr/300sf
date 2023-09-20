#!/usr/bin/env bash

crowdin download sources
crowdin pull
./json2ts.sh en sv > "../src/Translations.ts"
