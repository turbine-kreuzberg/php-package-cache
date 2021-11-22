#!/bin/sh
echo `which php-package-cache` | entr -nr `which php-package-cache` $@
