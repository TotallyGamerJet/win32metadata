package md

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestHasConstantIndexSize checks that the HasConstant composite index is sized
// from the tables it actually encodes.
//
// II.24.2.6 defines HasConstant over Field, Param and Property. Sizing it from
// any other set only diverges when exactly one of the two crosses the 2^14 row
// limit that widens a three-table composite index from two bytes to four.
// Neither winmd in _testdata does, but Windows.UI.Xaml.winmd has 21293 Param
// rows against at most 9371 in TypeDef, TypeRef and Property: the Constant
// table there was read two bytes too narrow, which misaligned every table
// stored after it.
func TestHasConstantIndexSize(t *testing.T) {
	// A three-table composite index spends 2 bits on the tag, so it stays
	// narrow only while every table it encodes has fewer than 1<<14 rows.
	const widensIndex = 1 << 14

	tests := []struct {
		name     string
		rows     map[TableType]uint32
		wantSize uint32
	}{
		{
			name:     "narrow when every encoded table is small",
			rows:     map[TableType]uint32{Param: widensIndex - 1},
			wantSize: 2,
		},
		{
			name:     "wide when Param is large",
			rows:     map[TableType]uint32{Param: widensIndex},
			wantSize: 4,
		},
		{
			name:     "wide when Field is large",
			rows:     map[TableType]uint32{Field: widensIndex},
			wantSize: 4,
		},
		{
			name:     "wide when Property is large",
			rows:     map[TableType]uint32{Property: widensIndex},
			wantSize: 4,
		},
		{
			name: "unaffected by tables it does not encode",
			rows: map[TableType]uint32{
				TypeDef: widensIndex,
				TypeRef: widensIndex,
			},
			wantSize: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)

			var h TablesHeader // HeapSizes 0: two byte string, GUID and blob indexes
			for table, rows := range tt.rows {
				h.Tables[table].RowCount = rows
			}
			h.computeIndexes()

			// Constant is Type, Parent (HasConstant), Value (blob index).
			a.Equal(tt.wantSize, h.Tables[Constant].Columns[1].Size)
			a.Equal(2+tt.wantSize+2, h.Tables[Constant].RowSize)
		})
	}
}
