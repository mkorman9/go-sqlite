### Database replication

See [Litestream Install](https://litestream.io/install/debian/) and 
[Litestream Getting Started](https://litestream.io/getting-started/).

Run with:
```bash
./go-sqlite --dsn "data.db?_pragma=journal_mode(WAL)"
```
