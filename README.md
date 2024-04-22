# SQLite Schema Syncer

This tool will extract and sync sqlite database schema's non-destructivly.
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
Since the tools uses cgo sqlite package you need to install `zig` if you want to cross compile for other platforms, everthing is defined in `build.sh`.