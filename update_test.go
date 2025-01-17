package dbr

import (
	"testing"

	"github.com/lianchengwu/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

func TestUpdateStmt(t *testing.T) {
	buf := NewBuffer()
	builder := Update("table").Set("a", 1).Where(Eq("b", 2))
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)

	assert.Equal(t, "UPDATE `table` SET `a` = ? WHERE (`b` = ?)", buf.String())
	assert.Equal(t, []interface{}{1, 2}, buf.Value())
}

func TestUpdateStmtSetRecord(t *testing.T) {
	record := struct{ A int }{A: 1}
	buf := NewBuffer()
	builder := Update("table").SetRecord(&record).Where(Eq("b", 2))
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)

	assert.Equal(t, "UPDATE `table` SET `a` = ? WHERE (`b` = ?)", buf.String())
	assert.Equal(t, []interface{}{1, 2}, buf.Value())
}

func BenchmarkUpdateValuesSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Update("table").Set("a", 1).Build(dialect.MySQL, buf)
	}
}

func BenchmarkUpdateMapSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Update("table").SetMap(map[string]interface{}{"a": 1, "b": 2}).Build(dialect.MySQL, buf)
	}
}
