package sqldb

import (
	"context"
	"fmt"
)

func QueryItems[
M any,           // Model struct
MP Scannable[M], // *Model Implementing Scannable[M]
](
	ctx context.Context,
	DBHandle DBHandle,
	rawStmt string,
	args ...any, // variadic
) ([]*M, error) { // Returns a Slice of Model-Pointers
	rows, err := DBHandle.QueryRows(ctx, rawStmt, args...)
	if err != nil {
		return nil, err
	}
	return RowsToNewItems[M, MP](rows)
}

func RowsToNewItems[
M any,           // Model struct
MP Scannable[M], // *Model Implementing Scannable[M]
](rows Rows) ([]*M, error) { // Returns a Slice of Model-Pointers
	var itemPtrs []*M
	for rows.Next() {
		var item M     // struct with zero values for the fields
		p := MP(&item) // p is *M, which satisfies targetFieldsProvider interface
		// Scan the Fields of Each Row to the Fields of the new struct of the Model
		if err := rows.Scan(p.TargetFields()...); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		itemPtrs = append(itemPtrs, &item) // Collect the pointers
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows: %v", err)
	}
	return itemPtrs, nil
}

func QueryItem[
M any,           // Model struct
MP Scannable[M], // *Model Implementing Scannable[M]
](
	ctx context.Context,
	DBHandle DBHandle,
	rawStmt string,
	args ...any, // variadic
) (*M, error) { // Returns the Pointer to the Newly Created Item
	row := DBHandle.QueryRow(ctx, rawStmt, args...)
	return RowToNewItem[M, MP](row)
}

func RowToNewItem[
M any,           // Model struct
MP Scannable[M], // *Model Implementing Scannable[M]
](row Row) (*M, error) { // Returns the Pointer to the Newly Created Item
	var item M     // struct with zero values for the fields
	p := MP(&item) // p is *M, which satisfies targetFieldsProvider interface
	err := row.Scan(p.TargetFields()...)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
