package sqldb

import (
	"context"
	"fmt"
)

func QueryItems[
M any,          // Model struct
P Scannable[M], // *Model Implementing Scannable[M]
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
	return RowsToItems[M, P](rows)
}

func RowsToItems[
M any,          // Model struct
P Scannable[M], // *Model Implementing Scannable[M]
](rows Rows) ([]*M, error) { // Returns a Slice of Model-Pointers
	var itemptrs []*M
	for rows.Next() {
		var item M    // struct with zero values for the fields
		p := P(&item) // p is *M, which satisfies targetFieldsProvider interface
		// Scan the Fields of Each Row to the Fields of the new struct of the Model
		if err := rows.Scan(p.TargetFields()...); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		itemptrs = append(itemptrs, &item) // Collect the pointers
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows: %v", err)
	}
	return itemptrs, nil
}

func QueryIDItemsMap[
M any,                          // Model struct
P ScannableIdentifiable[M, ID], // *Model Implementing ScannableIdentifiable[M, ID]
ID comparable,
](
	ctx context.Context,
	DBHandle DBHandle,
	rawStmt string,
	args ...any, // variadic
) (map[ID]*M, error) { // Returns a Map of ID to Model-Pointers
	rows, err := DBHandle.QueryRows(ctx, rawStmt, args...)
	if err != nil {
		return nil, err
	}
	return RowsToIDItemsMap[M, P, ID](rows)
}

func RowsToIDItemsMap[
M any,                          // Model struct
P ScannableIdentifiable[M, ID], // *Model Implementing ScannableIdentifiable[M, ID]
ID comparable,
](rows Rows) (map[ID]*M, error) { // Returns a Map of ID to Model-Pointers
	idItemptrs := map[ID]*M{}
	for rows.Next() {
		var item M    // struct with zero values for the fields
		p := P(&item) // p is *M, which satisfies targetFieldsProvider interface
		// Scan the Fields of Each Row to the Fields of the new struct of the Model
		if err := rows.Scan(p.TargetFields()...); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		idItemptrs[p.GetID()] = &item
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows: %v", err)
	}
	return idItemptrs, nil
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
