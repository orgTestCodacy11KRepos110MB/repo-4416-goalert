#!/bin/sh
set -e

su postgres -c "/usr/lib/postgresql/13/bin/pg_ctl start -w -l /var/log/postgresql/server.log"
