package dbr

import (
	"github.com/lianchengwu/dbr/dialect"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionCommit(t *testing.T) {
	for _, sess := range testSession {
		tx, err := sess.Begin()
		assert.NoError(t, err)
		defer tx.RollbackUnlessCommitted()

		id := nextID()

		result, err := tx.InsertInto("dbr_people").Columns("id", "name", "email").Values(id, "Barack", "obama@whitehouse.gov").Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		// not all drivers supports RowsAffected
		if err == nil {
			assert.EqualValues(t, 1, rowsAffected)
		}

		err = tx.Commit()
		assert.NoError(t, err)

		var person person
		err = tx.Select("*").From("dbr_people").Where(Eq("id", id)).LoadStruct(&person)
		assert.Error(t, err)
	}
}

func TestTransactionRollback(t *testing.T) {
	for _, sess := range testSession {
		if sess.Dialect == dialect.ClickHouse {
			// clickhouse does not support transactions
			continue
		}
		tx, err := sess.Begin()
		assert.NoError(t, err)
		defer tx.RollbackUnlessCommitted()

		id := nextID()

		result, err := tx.InsertInto("dbr_people").Columns("id", "name", "email").Values(id, "Barack", "obama@whitehouse.gov").Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		err = tx.Rollback()
		assert.NoError(t, err)

		var person person
		err = tx.Select("*").From("dbr_people").Where(Eq("id", id)).LoadStruct(&person)
		assert.Error(t, err)
	}
}
