# migration

The **migration** package provides a way to generate, list, and apply PostgreSQL
or Sqlite3 database migrations.
The package can be imported like so:

```go
import (
	"code.soquee.net/migration"
)
```

Two build tags are available to pick between PostgreSQL and Sqlite3:

- `pgx5` (default)
- `sqlite`

When building with the `pgx5` build tag (or no build tags at all) the
[`github.com/jackc/pgx/v5`][pgx5] is imported and used to sanitize inputs.
When building with the `sqlite` build tag no specific sqlite driver is implied
or imported and a generic sanitization method is used.

[pgx5]: https://pkg.go.dev/github.com/jackc/pgx/v5


## License

The package may be used under the terms of the BSD 2-Clause License a copy of
which may be found in the [`LICENSE`] file.

Unless you explicitly state otherwise, any contribution submitted for inclusion
in the work by you shall be licensed as above, without any additional terms or
conditions.

[`LICENSE`]: ./LICENSE
