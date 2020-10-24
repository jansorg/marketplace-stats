#!/bin/bash
# Executed on Linux
# uses https://github.com/JarvusInnovations/puppeteer-cli
# uses pdftoppm

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$DIR/.."
go build -o random-report ./cmds/random-report
./random-report > ./random-report.html
puppeteer --wait-until "networkidle2" print ./random-report.html ./random-report.pdf
pdftoppm -jpeg -r 160 -singlefile ./random-report.pdf random-report
