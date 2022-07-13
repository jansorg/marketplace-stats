#!/bin/bash
# Executed on Linux
# uses https://github.com/JarvusInnovations/puppeteer-cli
# uses pdftoppm

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TARGET_DIR="$DIR/.."
cd "$DIR"

go build -o random-report-cli ./random-report
./random-report-cli > "$TARGET_DIR/random-report.html"

npx puppeteer-cli --wait-until "networkidle2" print "$TARGET_DIR/random-report.html" "$TARGET_DIR/random-report.pdf"
pdftoppm -jpeg -jpegopt "quality=80,optimize=y" -r 120 -singlefile "$TARGET_DIR/random-report.pdf" "$TARGET_DIR/random-report"
