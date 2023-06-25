package db

import "context"

func ExecClickHouseQuery(ctx context.Context, db *ClickHouse, query string, args ...interface{}) error {
	return db.conn.Exec(ctx, query, args...)
}
