package database

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fan/safe-mysql-mcp/internal/config"
)

// newMockRouter creates a Router backed by a sqlmock *sql.DB.
// Returns the router, the mock, and a cleanup function.
func newMockRouter(t *testing.T, dbName string) (*Router, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	pool := &Pool{
		clusters: map[string]*managedDB{
			"primary": {db: db},
		},
		configs: config.ClustersConfig{
			"primary": {Host: "localhost", Port: 3306},
		},
	}

	databases := config.DatabasesConfig{
		dbName: {Cluster: "primary"},
	}

	router := NewRouter(pool, databases)
	cleanup := func() { db.Close() }

	return router, mock, cleanup
}

func TestRouter_Query_Success(t *testing.T) {
	router, mock, cleanup := newMockRouter(t, "testdb")
	defer cleanup()

	// Expect: Conn() → USE → QueryContext → Rows
	mock.ExpectExec("USE `testdb`").WillReturnResult(sqlmock.NewResult(0, 0))

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")
	mock.ExpectQuery("SELECT .* FROM users").WillReturnRows(rows)

	result, err := router.Query(context.Background(), "testdb", "SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	defer result.Close()

	var id int
	var name string
	count := 0
	for result.Next() {
		if err := result.Scan(&id, &name); err != nil {
			t.Fatalf("Scan() error = %v", err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("got %d rows, want 2", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRouter_Query_UseDBError(t *testing.T) {
	router, mock, cleanup := newMockRouter(t, "testdb")
	defer cleanup()

	// USE fails → conn should be closed
	mock.ExpectExec("USE `testdb`").WillReturnError(errors.New("unknown database"))

	_, err := router.Query(context.Background(), "testdb", "SELECT 1")
	if err == nil {
		t.Fatal("expected error when USE fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRouter_Query_QueryError(t *testing.T) {
	router, mock, cleanup := newMockRouter(t, "testdb")
	defer cleanup()

	mock.ExpectExec("USE `testdb`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("syntax error"))

	_, err := router.Query(context.Background(), "testdb", "SELECT 1")
	if err == nil {
		t.Fatal("expected error when query fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRouter_Query_UnknownDatabase(t *testing.T) {
	router, _, cleanup := newMockRouter(t, "testdb")
	defer cleanup()

	_, err := router.Query(context.Background(), "nonexistent", "SELECT 1")
	if err == nil {
		t.Fatal("expected error for unknown database")
	}
}

func TestRouter_Query_CancelledContext(t *testing.T) {
	router, _, cleanup := newMockRouter(t, "testdb")
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Context is already cancelled, Conn() should fail
	_, err := router.Query(ctx, "testdb", "SELECT 1")
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

func TestRouter_Exec_Success(t *testing.T) {
	router, mock, cleanup := newMockRouter(t, "testdb")
	defer cleanup()

	mock.ExpectExec("USE `testdb`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO users").WithArgs("Alice").WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := router.Exec(context.Background(), "testdb", "INSERT INTO users (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	affected, _ := result.RowsAffected()
	if affected != 1 {
		t.Errorf("rows affected = %d, want 1", affected)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRouter_Exec_UseDBError(t *testing.T) {
	router, mock, cleanup := newMockRouter(t, "testdb")
	defer cleanup()

	mock.ExpectExec("USE `testdb`").WillReturnError(errors.New("access denied"))

	_, err := router.Exec(context.Background(), "testdb", "INSERT INTO users (name) VALUES ('test')")
	if err == nil {
		t.Fatal("expected error when USE fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRouter_Exec_UnknownDatabase(t *testing.T) {
	router, _, cleanup := newMockRouter(t, "testdb")
	defer cleanup()

	_, err := router.Exec(context.Background(), "missingdb", "INSERT INTO t VALUES (1)")
	if err == nil {
		t.Fatal("expected error for unknown database")
	}
}

func TestRouter_GetDB_Success(t *testing.T) {
	router, _, cleanup := newMockRouter(t, "myapp")
	defer cleanup()

	db, err := router.GetDB("myapp")
	if err != nil {
		t.Fatalf("GetDB() error = %v", err)
	}
	if db == nil {
		t.Fatal("GetDB() returned nil db")
	}
}

func TestRouter_GetDB_Unknown(t *testing.T) {
	router, _, cleanup := newMockRouter(t, "myapp")
	defer cleanup()

	_, err := router.GetDB("other")
	if err == nil {
		t.Fatal("expected error for unknown database")
	}
}

func TestRouter_Query_SpecialCharsInDBName(t *testing.T) {
	// Database names with backticks or special chars should be properly quoted
	router, mock, cleanup := newMockRouter(t, "my-db")
	defer cleanup()

	mock.ExpectExec("USE `my-db`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	rows, err := router.Query(context.Background(), "my-db", "SELECT 1")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	rows.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRouter_ConcurrentAccess(t *testing.T) {
	// Verify Router methods are safe for concurrent use.
	// We interleave reads (GetCluster, ListDatabases) which don't need mock.
	pool := &Pool{
		clusters: map[string]*managedDB{"primary": {db: nil}},
		configs:  config.ClustersConfig{"primary": {Host: "localhost"}},
	}
	databases := config.DatabasesConfig{
		"db1": {Cluster: "primary"},
		"db2": {Cluster: "primary"},
	}
	router := NewRouter(pool, databases)

	const n = 100
	errCh := make(chan error, n*3)

	for i := 0; i < n; i++ {
		go func() {
			_, err := router.GetCluster("db1")
			errCh <- err
		}()
		go func() {
			_ = router.ListDatabases()
			errCh <- nil
		}()
		go func() {
			_, err := router.GetCluster("db2")
			errCh <- err
		}()
	}

	for i := 0; i < n*3; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent access failed: %v", err)
		}
	}
}
