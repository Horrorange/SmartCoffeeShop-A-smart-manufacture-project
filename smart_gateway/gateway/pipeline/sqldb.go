package pipeline

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "time"
    _ "github.com/go-sql-driver/mysql"
)

type SQLDB struct {
    db     *sql.DB
    driver string
}

func NewSQLDB() (*SQLDB, error) {
    host := getenv("PG_HOST", "localhost")
    port := getenv("PG_PORT", "5432")
    dbname := getenv("PG_DB", "smartshop")
    user := getenv("PG_USER", "postgres")
    pass := getenv("PG_PASS", "")
    driver := getenv("DB_TYPE", "postgres")
    var (
        db  *sql.DB
        err error
    )
    if driver == "mysql" {
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local", user, pass, host, port, dbname)
        db, err = sql.Open("mysql", dsn)
    } else {
        dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)
        db, err = sql.Open("postgres", dsn)
    }
    if err != nil { return nil, err }
    if err = db.Ping(); err != nil { return nil, err }
    s := &SQLDB{db: db, driver: driver}
    if err = s.ensureSchema(context.Background()); err != nil { return nil, err }
    return s, nil
}

func (s *SQLDB) ensureSchema(ctx context.Context) error {
    if s.driver == "mysql" {
        if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS orders (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  coffee_type VARCHAR(32) NOT NULL,
  bool_ice TINYINT(1) NOT NULL,
  table_num INT NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'pending',
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  finish_time TIMESTAMP NULL,
  error_msg TEXT
)`); err != nil { return err }
        if _, err := s.db.ExecContext(ctx, `CREATE INDEX idx_orders_status_ctime ON orders(status, create_time)`); err != nil { return err }
        return nil
    }
    ddl := `
CREATE TABLE IF NOT EXISTS orders (
  id BIGSERIAL PRIMARY KEY,
  coffee_type VARCHAR(32) NOT NULL,
  bool_ice BOOLEAN NOT NULL,
  table_num INT NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'pending',
  create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  finish_time TIMESTAMPTZ,
  error_msg TEXT
);
CREATE INDEX IF NOT EXISTS idx_orders_status_ctime ON orders(status, create_time DESC);`
    _, err := s.db.ExecContext(ctx, ddl)
    return err
}

func (s *SQLDB) PollPending(limit int) ([]Order, error) {
    tx, err := s.db.Begin()
    if err != nil { return nil, err }
    var rows *sql.Rows
    if s.driver == "mysql" {
        rows, err = tx.Query(`SELECT id, coffee_type, bool_ice, table_num, create_time FROM orders WHERE status='pending' ORDER BY create_time ASC LIMIT ? FOR UPDATE`, limit)
    } else {
        rows, err = tx.Query(`SELECT id, coffee_type, bool_ice, table_num, create_time FROM orders WHERE status='pending' ORDER BY create_time ASC FOR UPDATE SKIP LOCKED LIMIT $1`, limit)
    }
    if err != nil { tx.Rollback(); return nil, err }
    var res []Order
    var ids []int
    for rows.Next() {
        var o Order
        var ct time.Time
        if err = rows.Scan(&o.ID, &o.CoffeeType, &o.BoolIce, &o.TableNum, &ct); err != nil { rows.Close(); tx.Rollback(); return nil, err }
        o.CreateTime = ct
        o.Status = "pending"
        res = append(res, o)
        ids = append(ids, o.ID)
    }
    rows.Close()
    for _, id := range ids {
        if s.driver == "mysql" {
            if _, err = tx.Exec(`UPDATE orders SET status='queued' WHERE id=?`, id); err != nil { tx.Rollback(); return nil, err }
        } else {
            if _, err = tx.Exec(`UPDATE orders SET status='queued' WHERE id=$1`, id); err != nil { tx.Rollback(); return nil, err }
        }
    }
    if err = tx.Commit(); err != nil { return nil, err }
    return res, nil
}

func (s *SQLDB) MarkQueued(ids []int) error { return nil }

func (s *SQLDB) UpdateResult(res Result) error {
    status := "error"
    if res.BoolIsDone { status = "done" }
    if s.driver == "mysql" {
        _, err := s.db.Exec(`UPDATE orders SET status=?, finish_time=?, error_msg=? WHERE id=?`, status, time.Now(), nullIfEmpty(res.ErrorMsg), res.ID)
        return err
    }
    _, err := s.db.Exec(`UPDATE orders SET status=$1, finish_time=$2, error_msg=$3 WHERE id=$4`, status, time.Now(), nullIfEmpty(res.ErrorMsg), res.ID)
    return err
}

func (s *SQLDB) Seed(orders []Order) {
    for i := range orders {
        if s.driver == "mysql" {
            _, _ = s.db.Exec(`INSERT INTO orders(coffee_type, bool_ice, table_num, status, create_time) VALUES(?,?,?,'pending',NOW())`, orders[i].CoffeeType, orders[i].BoolIce, orders[i].TableNum)
        } else {
            _, _ = s.db.Exec(`INSERT INTO orders(coffee_type, bool_ice, table_num, status, create_time) VALUES($1,$2,$3,'pending',NOW())`, orders[i].CoffeeType, orders[i].BoolIce, orders[i].TableNum)
        }
    }
}

func (s *SQLDB) All() []Order {
    rows, err := s.db.Query(`SELECT id, coffee_type, bool_ice, table_num, status, create_time, finish_time, COALESCE(error_msg,'') FROM orders ORDER BY id`)
    if err != nil { return nil }
    defer rows.Close()
    var res []Order
    for rows.Next() {
        var o Order
        var ct time.Time
        var ft sql.NullTime
        var em string
        if err = rows.Scan(&o.ID, &o.CoffeeType, &o.BoolIce, &o.TableNum, &o.Status, &ct, &ft, &em); err != nil { return res }
        o.CreateTime = ct
        if ft.Valid { o.FinishTime = ft.Time }
        o.ErrorMsg = em
        res = append(res, o)
    }
    return res
}

func getenv(k, d string) string { v := os.Getenv(k); if v == "" { return d }; return v }

func nullIfEmpty(s string) interface{} { if s == "" { return nil }; return s }