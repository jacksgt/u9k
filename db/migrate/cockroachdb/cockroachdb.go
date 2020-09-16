package cockroachdb

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	nurl "net/url"
	"regexp"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/hashicorp/go-multierror"
	"github.com/jackc/pgx/v4"
)

func init() {
	db := CockroachDb{}
	database.Register("cockroach", &db)
	database.Register("cockroachdb", &db)
	database.Register("crdb-postgres", &db)
}

var DefaultMigrationsTable = "schema_migrations"
var DefaultLockTable = "schema_lock"

var (
	ErrNilConfig      = fmt.Errorf("no config")
	ErrNoDatabaseName = fmt.Errorf("no database name")
)

type Config struct {
	MigrationsTable string
	LockTable       string
	ForceLock       bool
	DatabaseName    string
}

type CockroachDb struct {
	db       *pgx.Conn
	isLocked bool

	// Open and WithInstance need to guarantee that config is never nil
	config *Config
}

func WithInstance(instance *pgx.Conn, config *Config) (database.Driver, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if err := instance.Ping(context.Background()); err != nil {
		return nil, err
	}

	if config.DatabaseName == "" {
		query := `SELECT current_database()`
		var databaseName string
		if err := instance.QueryRow(context.Background(), query).Scan(&databaseName); err != nil {
			return nil, &database.Error{OrigErr: err, Query: []byte(query)}
		}

		if len(databaseName) == 0 {
			return nil, ErrNoDatabaseName
		}

		config.DatabaseName = databaseName
	}

	if len(config.MigrationsTable) == 0 {
		config.MigrationsTable = DefaultMigrationsTable
	}

	if len(config.LockTable) == 0 {
		config.LockTable = DefaultLockTable
	}

	px := &CockroachDb{
		db:     instance,
		config: config,
	}

	// ensureVersionTable is a locking operation, so we need to ensureLockTable before we ensureVersionTable.
	if err := px.ensureLockTable(); err != nil {
		return nil, err
	}

	if err := px.ensureVersionTable(); err != nil {
		return nil, err
	}

	return px, nil
}

func (c *CockroachDb) Open(url string) (database.Driver, error) {
	purl, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	// As Cockroach uses the postgres protocol, and 'postgres' is already a registered database, we need to replace the
	// connect prefix, with the actual protocol, so that the library can differentiate between the implementations
	re := regexp.MustCompile("^(cockroach(db)?|crdb-postgres)")
	connectString := re.ReplaceAllString(migrate.FilterCustomQuery(purl).String(), "postgres")

	db, err := pgx.Connect(context.Background(), connectString)
	if err != nil {
		return nil, err
	}

	migrationsTable := purl.Query().Get("x-migrations-table")
	if len(migrationsTable) == 0 {
		migrationsTable = DefaultMigrationsTable
	}

	lockTable := purl.Query().Get("x-lock-table")
	if len(lockTable) == 0 {
		lockTable = DefaultLockTable
	}

	forceLockQuery := purl.Query().Get("x-force-lock")
	forceLock, err := strconv.ParseBool(forceLockQuery)
	if err != nil {
		forceLock = false
	}

	px, err := WithInstance(db, &Config{
		DatabaseName:    purl.Path,
		MigrationsTable: migrationsTable,
		LockTable:       lockTable,
		ForceLock:       forceLock,
	})
	if err != nil {
		return nil, err
	}

	return px, nil
}

func (c *CockroachDb) Close() error {
	return c.db.Close(context.Background())
}

// Locking is done manually with a separate lock table.  Implementing advisory locks in CRDB is being discussed
// See: https://github.com/cockroachdb/cockroach/issues/13546
func (c *CockroachDb) Lock() error {
	tx, err := c.db.Begin(context.Background())
	if err != nil {
		return err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	aid, err := database.GenerateAdvisoryLockId(c.config.DatabaseName)
	if err != nil {
		return err
	}

	query := "SELECT * FROM " + c.config.LockTable + " WHERE lock_id = $1"
	rows, err := tx.Query(context.Background(), query, aid)
	if err != nil {
		return database.Error{OrigErr: err, Err: "failed to fetch migration lock", Query: []byte(query)}
	}
	defer rows.Close()

	// If row exists at all, lock is present
	locked := rows.Next()
	if locked && !c.config.ForceLock {
		c.isLocked = true
		return database.ErrLocked
	}

	query = "INSERT INTO " + c.config.LockTable + " (lock_id) VALUES ($1)"
	if _, err := tx.Exec(context.Background(), query, aid); err != nil {
		return database.Error{OrigErr: err, Err: "failed to set migration lock", Query: []byte(query)}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// Locking is done manually with a separate lock table.  Implementing advisory locks in CRDB is being discussed
// See: https://github.com/cockroachdb/cockroach/issues/13546
func (c *CockroachDb) Unlock() error {
	aid, err := database.GenerateAdvisoryLockId(c.config.DatabaseName)
	if err != nil {
		return err
	}

	// In the event of an implementation (non-migration) error, it is possible for the lock to not be released.  Until
	// a better locking mechanism is added, a manual purging of the lock table may be required in such circumstances
	query := "DELETE FROM " + c.config.LockTable + " WHERE lock_id = $1"

	_, err = c.db.Exec(context.Background(), query, aid)
	switch {
	case err == nil:
		c.isLocked = false
		return nil
	case err == pgx.ErrNoRows: // pgx also handles "UndefinedTableError" as ErrNoRows
		c.isLocked = false
		return nil
	default:
		return database.Error{OrigErr: err, Err: "failed to release migration lock", Query: []byte(query)}
	}
}

func (c *CockroachDb) Run(migration io.Reader) error {
	migr, err := ioutil.ReadAll(migration)
	if err != nil {
		return err
	}

	// run migration
	query := string(migr[:])
	if _, err := c.db.Exec(context.Background(), query); err != nil {
		return database.Error{OrigErr: err, Err: "migration failed", Query: migr}
	}

	return nil
}

func (c *CockroachDb) SetVersion(version int, dirty bool) error {
	tx, err := c.db.Begin(context.Background())
	if err != nil {
		return err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	if _, err := tx.Exec(context.Background(), `DELETE FROM "`+c.config.MigrationsTable+`"`); err != nil {
		return err
	}

	// Also re-write the schema version for nil dirty versions to prevent
	// empty schema version for failed down migration on the first migration
	// See: https://github.com/golang-migrate/migrate/issues/330
	if version >= 0 || (version == database.NilVersion && dirty) {
		if _, err := tx.Exec(context.Background(), `INSERT INTO "`+c.config.MigrationsTable+`" (version, dirty) VALUES ($1, $2)`, version, dirty); err != nil {
			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (c *CockroachDb) Version() (version int, dirty bool, err error) {
	query := `SELECT version, dirty FROM "` + c.config.MigrationsTable + `" LIMIT 1`
	err = c.db.QueryRow(context.Background(), query).Scan(&version, &dirty)

	switch {
	// pgx also handles "UndefinedTableError" as ErrNoRows
	case err == pgx.ErrNoRows:
		return database.NilVersion, false, nil
	default:
		return version, dirty, nil
	}
}

func (c *CockroachDb) Drop() (err error) {
	// select all tables in current schema
	query := `SELECT table_name FROM information_schema.tables WHERE table_schema=(SELECT current_schema())`
	tables, err := c.db.Query(context.Background(), query)
	if err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	defer tables.Close()

	// delete one table after another
	tableNames := make([]string, 0)
	for tables.Next() {
		var tableName string
		if err := tables.Scan(&tableName); err != nil {
			return err
		}
		if len(tableName) > 0 {
			tableNames = append(tableNames, tableName)
		}
	}

	if len(tableNames) > 0 {
		// delete one by one ...
		for _, t := range tableNames {
			query = `DROP TABLE IF EXISTS ` + t + ` CASCADE`
			if _, err := c.db.Exec(context.Background(), query); err != nil {
				return &database.Error{OrigErr: err, Query: []byte(query)}
			}
		}
	}

	return nil
}

// ensureVersionTable checks if versions table exists and, if not, creates it.
// Note that this function locks the database, which deviates from the usual
// convention of "caller locks" in the CockroachDb type.
func (c *CockroachDb) ensureVersionTable() (err error) {
	if err = c.Lock(); err != nil {
		return err
	}

	defer func() {
		if e := c.Unlock(); e != nil {
			if err == nil {
				err = e
			} else {
				err = multierror.Append(err, e)
			}
		}
	}()

	// check if migration table exists
	var count int
	query := `SELECT COUNT(1) FROM information_schema.tables WHERE table_name = $1 AND table_schema = (SELECT current_schema()) LIMIT 1`
	if err := c.db.QueryRow(context.Background(), query, c.config.MigrationsTable).Scan(&count); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	if count == 1 {
		return nil
	}

	// if not, create the empty migration table
	query = `CREATE TABLE "` + c.config.MigrationsTable + `" (version INT NOT NULL PRIMARY KEY, dirty BOOL NOT NULL)`
	if _, err := c.db.Exec(context.Background(), query); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	return nil
}

func (c *CockroachDb) ensureLockTable() error {
	// check if lock table exists
	var count int
	query := `SELECT COUNT(1) FROM information_schema.tables WHERE table_name = $1 AND table_schema = (SELECT current_schema()) LIMIT 1`
	if err := c.db.QueryRow(context.Background(), query, c.config.LockTable).Scan(&count); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	if count == 1 {
		return nil
	}

	// if not, create the empty lock table
	query = `CREATE TABLE "` + c.config.LockTable + `" (lock_id INT NOT NULL PRIMARY KEY)`
	if _, err := c.db.Exec(context.Background(), query); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	return nil
}
