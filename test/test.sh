#!/bin/bash
set -euxo pipefail

time composer install --ignore-platform-reqs --no-scripts --no-dev --no-autoloader --no-progress #-vvv

echo done

# vendor summary
# find vendor | wc -l
# du -sch vendor
