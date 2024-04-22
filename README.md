# SQLite Schema Syncer

This tool will extract and sync sqlite database schema's non-destructivly.

```sh
# extract schema from test.db
./sqliteschema extract test.db > test.json

# create or sync test2.db with schema in test.json
./sqliteschema sync test2.db test.json

```