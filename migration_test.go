package migration_test

import (
	"context"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"code.soquee.net/migration"
)

func newMapFS(m map[string]string) fstest.MapFS {
	mfs := make(fstest.MapFS)
	for k, v := range m {
		mfs[k] = &fstest.MapFile{
			Data: []byte(v),
		}
	}
	return mfs
}

const migrationsTable = "__migrations"

var genTestCases = [...]struct {
	name string
	err  error
	dir  string
}{
	0: {
		name: "test",
		err:  nil,
		dir:  "test",
	},
	1: {
		name: "test me'\"\tagain",
		err:  nil,
		dir:  "test_me_again",
	},
}

func TestLastRun(t *testing.T) {
	for i, tc := range genTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			migrationDir := "001_" + tc.dir
			fs := newMapFS(map[string]string{
				path.Join(migrationDir, "up.sql"): "-- up.sql",
			})
			ident, name, err := migration.LastRun(context.Background(), migrationsTable, fs, nil)
			if ident != "" {
				t.Errorf("Wrong ident: %q", ident)
			}
			if name != migrationDir {
				t.Errorf("Wrong name: want=%q, got=%q", migrationDir, name)
			}
			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	for i, tc := range genTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir, err := os.MkdirTemp("", "migration_test")
			if err != nil {
				t.Fatalf("Error creating temp directory for tests: %q", err)
			}
			defer func() {
				err = os.RemoveAll(dir)
				if err != nil {
					t.Logf("Error cleaning up temp dir %q: %q", dir, err)
				}
			}()

			err = migration.Generator(dir)(tc.name)
			if err != tc.err {
				t.Errorf("Unexpected error: want=%q, got=%q", tc.err, err)
			}

			walked := 0
			walker, err := migration.NewWalker(context.Background(), migrationsTable, nil)
			if err != nil {
				t.Fatalf("error creating walker: %q", err)
			}
			err = walker(os.DirFS(dir), func(name string, info fs.DirEntry, status migration.RunStatus) error {
				dirName := info.Name()
				if dirName == path.Base(dir) {
					t.Fatalf("Walk included top level directory but should only hit migrations")
				}
				walked++
				if walked > 1 {
					t.Fatalf("Too many files created in temp dir %q, is cleanup working?", dir)
				}

				// TODO: test name generation
				if status != migration.StatusUnknown {
					t.Errorf("Unexpected status: want=%d, got=%d", migration.StatusUnknown, status)
				}

				idx := strings.Index(dirName, "_")
				if idx < 0 {
					idx = 0
				}
				dirName = dirName[idx+1:]
				if dirName != tc.dir {
					t.Errorf("Unexpected migration dir: want=%q, got=%q", tc.dir, dirName)
				}
				return nil
			})
			if err != nil {
				t.Errorf("Unexpected error walking test output: %q", err)
			}
		})
	}
}
