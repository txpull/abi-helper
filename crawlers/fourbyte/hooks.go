package fourbyte

import (
	"github.com/txpull/bytecode/db"
	"github.com/txpull/bytecode/signatures"
)

func PostWriteClickHouseHook(db *db.ClickHouse) PostWriteHook {
	return func(signature *signatures.Signature) error {
		panic("OK")
		return nil
	}
}
