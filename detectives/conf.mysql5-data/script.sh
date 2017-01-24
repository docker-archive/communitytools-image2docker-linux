#!/bin/sh

cd /v2c/disk/

FN=$(sed -n 's/^datadir[ \t]*= \(.*\)$/\1/p' etc/mysql/my.cnf 2> /dev/null | sed 's/^.//' 2> /dev/null)
tar c etc/mysql $FN 2> /dev/null
