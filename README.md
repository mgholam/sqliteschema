# SQLite Schema Syncer

This tool will extract and sync sqlite database schema's non-destructively.
- Tables will not be removed.
- Columns will not be removed, only added.
- If a database does not exist, it will be created.

```sh
# extract schema from test.db into test.json
./sqliteschema extract test.db > test.json

# create or sync test2.db with schema in test.json
./sqliteschema sync test2.db test.json

```
## Building and Cross Compiling
Since the tools uses cgo sqlite package you need to install `zig` if you want to cross compile for other platforms, everything is defined in `build.sh`.

## usage
Given you have the following directory structure:
- data : sqlite database folder (i.e. test.db, ...)
- schema : schema files (i.e. test.db.json, ...)

The following shell script will update the database files:
```sh
#!/bin/bash
# update schema
for f in $(ls schema/*);
do
  l=`basename "$f" | sed "s/.json//"`
  db=$l
  ./sqliteschema sync data/$db $f
done

```