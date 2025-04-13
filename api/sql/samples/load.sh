#!/bin/bash

if [ $# -eq 0 ]; then
    	echo "Example: $0 <sqlite.db>"
    	exit 1
fi    


echo "==================Loading db================"
echo "===============Loading associations========="

sqlite3 $1 < associations.sql
echo "ok"

echo "===============Loading buildings============"

sqlite3 $1 < buildings.sql

echo "ok"

echo "===============Loading accounts============="

sqlite3 $1 < accounts.sql

echo "ok"
echo "===============Loading categories============"

sqlite3 $1 < categories.sql

echo "ok"

echo "===============Loading expenses=============="

sqlite3 $1 < expenses.sql

echo "ok"

echo "===============Loading units=============="

sqlite3 $1 < units.sql

echo "ok"

echo "===============Loading owners=============="

sqlite3 $1 < owners.sql

echo "ok"

echo "===============Loading ownerships=========="

sqlite3 $1 < ownerships.sql

echo "ok"

echo "===============Loading users==============="

sqlite3 $1 < users.sql

echo "ok"

echo "===============Loading users associations=="

sqlite3 $1 < users_associations.sql

echo "ok"




