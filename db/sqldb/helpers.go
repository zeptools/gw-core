package sqldb

import (
	"context"
	"fmt"
	"log"

	"github.com/zeptools/gw-core/orm"
)

func QueryItem[
	M any, // Model struct
	P Scannable[M], // *Model Implementing Scannable[M]
](
	ctx context.Context,
	DBHandle DBHandle,
	rawSQLStmt string,
	args ...any, // variadic
) (*M, error) { // Returns the Pointer to the Newly Created Item
	row := DBHandle.QueryRow(ctx, rawSQLStmt, args...)
	return RowToItem[M, P](row)
}

func RowToItem[
	M any, // Model struct
	P Scannable[M], // *Model Implementing Scannable[M]
](row Row) (*M, error) { // Returns the Pointer to the Newly Created Item
	var item M    // struct with zero values for the fields
	p := P(&item) // p is *M, which satisfies targetFieldsProvider interface
	err := row.Scan(p.TargetFields()...)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func QueryItems[
	M any, // Model struct
	P Scannable[M], // *Model Implementing Scannable[M]
](
	ctx context.Context,
	DBHandle DBHandle,
	rawSQLStmt string,
	args ...any, // variadic
) ([]*M, error) { // Returns a Slice of Model-Pointers
	rows, err := DBHandle.QueryRows(ctx, rawSQLStmt, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close() failed: %v", err)
		}
	}()
	return RowsToItems[M, P](rows)
}

func RowsToItems[
	M any, // Model struct
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

// QueryMap queries items using rawSQLStmt and scan rows to a map[id]item
func QueryMap[
	M any, // Model struct
	P ScannableIdentifiable[M, ID], // *Model Implementing ScannableIdentifiable[M, ID]
	ID comparable,
](
	ctx context.Context,
	DBHandle DBHandle,
	rawSQLStmt string,
	args ...any, // variadic
) (map[ID]*M, error) { // Returns a Map of ID to Model-Pointers
	rows, err := DBHandle.QueryRows(ctx, rawSQLStmt, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close() failed: %v", err)
		}
	}()
	return RowsToMap[M, P, ID](rows)
}

// RowsToMap scan rows to a map[id]item
func RowsToMap[
	M any, // Model struct
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

// QueryCollection queries items using rawSQLStmt and scan rows to a collection
func QueryCollection[
	M any, // Model struct
	P ScannableIdentifiable[M, ID], // *Model implementing ScannableIdentifiable[M, ID]
	ID comparable,
](
	ctx context.Context,
	DBHandle DBHandle,
	rawSQLStmt string,
	args ...any, // variadic
) (*orm.ModelCollection[P, ID], error) {
	rows, err := DBHandle.QueryRows(ctx, rawSQLStmt, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close() failed: %v", err)
		}
	}()
	return RowsToCollection[M, P, ID](rows)
}

// RowsToCollection scan rows to a collection
func RowsToCollection[
	M any, // Model struct
	P ScannableIdentifiable[M, ID], // *Model implementing ScannableIdentifiable[M, ID]
	ID comparable,
](
	rows Rows,
) (*orm.ModelCollection[P, ID], error) {
	coll := &orm.ModelCollection[P, ID]{
		Map:   make(map[ID]P),
		Order: []ID{},
	}
	for rows.Next() {
		var item M
		p := P(&item) // *M implementing ScannableIdentifiable
		if err := rows.Scan(p.TargetFields()...); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		id := p.GetID()
		coll.Map[id] = p
		coll.Order = append(coll.Order, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows: %v", err)
	}
	return coll, nil
}
