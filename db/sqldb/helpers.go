package sqldb

import (
	"context"
	"fmt"
)

func QueryItems[
	M any, // Model Base Type
	MP Scannable[M], // Model Pointer Type Implementing Scannable[M]
](
	ctx context.Context,
	DBHandle DBHandle,
	rawStmt string,
	args ...any, // variadic
) ([]M, error) {
	rows, err := DBHandle.QueryRows(ctx, rawStmt, args...)
	if err != nil {
		return nil, err
	}
	return RowsToNewItems[M, MP](rows)
}

func RowsToNewItems[
	M any, // Model Base Type
	MP Scannable[M], // Model Pointer Type Implementing Scannable[M]
](rows Rows) ([]M, error) {
	var items []M
	for rows.Next() {
		var item M     // struct with zero values for the fields
		p := MP(&item) // p is *M, which satisfies targetFieldsProvider interface
		if err := rows.Scan(p.TargetFields()...); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows: %v", err)
	}
	return items, nil
}

func QueryItem[
	M any, // Model Base Type
	MP Scannable[M], // Model Pointer Type Implementing Scannable[M]
](
	ctx context.Context,
	DBHandle DBHandle,
	rawStmt string,
	args ...any, // variadic
) (*M, error) {
	row := DBHandle.QueryRow(ctx, rawStmt, args...)
	return RowToNewItem[M, MP](row)
}

func RowToNewItem[
	M any, // Model Base Type
	MP Scannable[M], // Model Pointer Type Implementing Scannable[M]
](row Row) (*M, error) {
	var item M     // struct with zero values for the fields
	p := MP(&item) // p is *M, which satisfies targetFieldsProvider interface
	err := row.Scan(p.TargetFields()...)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
