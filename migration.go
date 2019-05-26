// Package migration contains functions for generating and finding PostgreSQL
// database migrations.
package migration // import "code.soquee.net/migration"

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/tools/godoc/vfs"
)

// Generator returns a function that creates migration files at the given base
// path.
func Generator(basePath string, debug *log.Logger) func(name string) error {
	replacer := strings.NewReplacer(
		" ", "_",
		"\t", "_",
		"'", "",
		"\"", "",
	)

	return func(name string) error {
		name = time.Now().Format("2006-01-02-150405_") + replacer.Replace(strings.TrimSpace(name))
		relPath := path.Join(basePath, name)

		// TODO: perform file creation operations in a temporary directory and then
		// move everything to the final location.
		err := os.MkdirAll(relPath, 0750)
		if err != nil {
			return err
		}

		upfile, err := os.Create(path.Join(relPath, "up.sql"))
		if err != nil {
			return err
		}
		defer func() {
			err := upfile.Close()
			if err != nil && debug != nil {
				debug.Printf("Error closing new up.sql: %q", err)
			}
		}()

		_, err = fmt.Fprintf(upfile, "-- Your SQL goes here")
		if err != nil {
			return err
		}

		downfile, err := os.Create(path.Join(relPath, "down.sql"))
		if err != nil {
			return err
		}
		defer func() {
			err := downfile.Close()
			if err != nil && debug != nil {
				debug.Printf("Error closing new down.sql: %q", err)
			}
		}()

		_, err = fmt.Fprintf(downfile, "-- This file should undo anything in `up.sql'")
		return err
	}
}

// RunStatus is a type that indicates if a migration has been run, not run, or
// if we can't determine the status.
type RunStatus int

// Valid RunStatus values. For more information see RunStatus.
const (
	StatusUnknown RunStatus = iota
	StatusNotRun
	StatusRun
)

func contains(sl []string, s string) int {
	for i, ss := range sl {
		if ss == s {
			return i
		}
	}
	return -1
}

func getRunMigrations(ctx context.Context, tx *sql.Tx, migrationsTable string) ([]string, error) {
	var ran []string
	err := tx.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT array_agg(version ORDER BY version ASC) FROM %s`, pq.QuoteIdentifier(migrationsTable)),
	).Scan(pq.Array(&ran))
	return ran, err
}

// LastRun returns the last migration directory by lexical order that exists in
// the database and on disk.
func LastRun(ctx context.Context, migrationsTable string, fs vfs.FileSystem, tx *sql.Tx, logger *log.Logger) (ident, name string, err error) {
	var version string
	if tx != nil {
		err = tx.QueryRowContext(ctx,
			fmt.Sprintf(`SELECT version FROM %s ORDER BY version DESC LIMIT 1`, pq.QuoteIdentifier(migrationsTable)),
		).Scan(&version)
		if err != nil {
			return version, "", err
		}
	}

	var fpath string
	err = Walk(ctx, migrationsTable, tx, fs, logger, func(name string, info os.FileInfo, status RunStatus) error {
		if tx != nil && name != version {
			return nil
		}
		fpath = info.Name()
		if tx != nil {
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return version, fpath, err
	}

	return version, fpath, nil
}

// WalkFunc is the type of the function called for each file or directory
// visited by a Walker.
type WalkFunc func(name string, info os.FileInfo, status RunStatus) error

// Walk walks the migrations it finds on the filesystem in lexical order
// (mostly, keep reading) and calls a function for each discovered migration,
// passing in its name, status, and file information.
// If a migration exists in the database but not on the filesystem, info will be
// nil and f will be called for it after the migrations that exist on the
// filesystem.
// No particular order is guaranteed for calls to f for migrations that do not
// exist on the filesystem.
func Walk(ctx context.Context, migrationsTable string, tx *sql.Tx, fs vfs.FileSystem, logger *log.Logger, f WalkFunc) error {
	var err error
	var ran []string
	if tx != nil {
		ran, err = getRunMigrations(ctx, tx, migrationsTable)
		if err != nil {
			logger.Printf("Error querying existing migrations: %q", err)
			tx = nil
		}
	}

	files, err := fs.ReadDir("/")
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, info := range files {
		if !info.IsDir() {
			continue
		}
		name := info.Name()
		idx := strings.Index(name, "_")
		if idx == -1 {
			continue
		}
		name = strings.Replace(name[:idx], "-", "", -1)
		var status RunStatus
		if tx != nil {
			if n := contains(ran, name); n != -1 {
				// The migration exists on the filesystem and in the database.
				// Since we found it, remove it from the list of previously ran
				// migrations.
				ran = append(ran[:n], ran[n+1:]...)
				status = StatusRun
			} else {
				// The migration only exists on the filesystem.
				status = StatusNotRun
			}
		}
		err = f(name, info, status)
		if err != nil {
			return err
		}
	}

	for _, missing := range ran {
		err = f(missing, nil, StatusRun)
		if err != nil {
			return err
		}
	}
	return nil
}
