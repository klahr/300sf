#!/usr/bin/env bash

crowdin pull
./json2ts.sh en sv > "../src/Translations.ts"
