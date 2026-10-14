#!/bin/bash
# Loads sample data into an APC SQLite database.
#
# The *.sql data files are NOT in git (api/.gitignore: sql/samples/*.sql):
# they hold real association data and are shared privately. Put them next to
# this script; missing files are skipped. Without them, create data in the
# app instead (Associations, then Units -> Management -> Import units).
#
# Usage: sql/samples/load.sh <sqlite.db>   (start the backend once first, so
# the schema exists)

set -euo pipefail

if [ $# -eq 0 ]; then
	echo "Usage: $0 <sqlite.db>"
	exit 1
fi

db=$(realpath "$1")
cd "$(dirname "$0")"

if ! command -v sqlite3 >/dev/null; then
	echo "sqlite3 is not installed"
	exit 1
fi

loaded=0
# In dependency order.
for name in associations buildings accounts categories expenses units owners ownerships users users_associations; do
	if [ -f "$name.sql" ]; then
		echo "Loading $name"
		sqlite3 "$db" < "$name.sql"
		loaded=$((loaded + 1))
	else
		echo "Skipping $name ($name.sql not found)"
	fi
done

if [ $loaded -eq 0 ]; then
	echo "No sample files found: they are not in git; ask a maintainer, or create data in the app."
	exit 1
fi
echo "Done: $loaded file(s) loaded."
