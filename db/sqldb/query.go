package sqldb

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/zeptools/gw-core/orm"
)

func RawQueryItem[
	M any, // Model struct
	MP Scannable[M], // *Model Implementing Scannable[M]
](
	ctx context.Context,
	dbClient Client,
	rawSQLStmt string,
	args ...any, // variadic
) (*M, error) { // Returns the Pointer to the Newly Created Item
	row := dbClient.QueryRow(ctx, rawSQLStmt, args...)
	return ScanRowToItem[M, MP](row)
}

func RawQueryItems[
	M any, // Model struct
	MP Scannable[M], // *Model Implementing Scannable[M]
](
	ctx context.Context,
	dbClient Client,
	rawSQLStmt string,
	args ...any, // variadic
) ([]*M, error) { // Returns a Slice of Model-Pointers
	rows, err := dbClient.QueryRows(ctx, rawSQLStmt, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close() failed: %v", err)
		}
	}()
	return ScanRowsToItems[M, MP](rows)
}

// RawQueryMap queries items using rawSQLStmt and scan rows to a map[id]item
func RawQueryMap[
	M any, // Model struct
	MP ScannableIdentifiable[M, ID], // *Model Implementing ScannableIdentifiable[M, ID]
	ID comparable,
](
	ctx context.Context,
	dbClient Client,
	rawSQLStmt string,
	args ...any, // variadic
) (map[ID]*M, error) { // Returns a ItemsMap of ID to Model-Pointers
	rows, err := dbClient.QueryRows(ctx, rawSQLStmt, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close() failed: %v", err)
		}
	}()
	return ScanRowsToMap[M, MP, ID](rows)
}

// RawQueryCollection queries items using rawSQLStmt and scan rows to a collection
func RawQueryCollection[
	M any, // Model struct
	MP ScannableIdentifiable[M, ID], // *Model implementing ScannableIdentifiable[M, ID]
	ID comparable,
](
	ctx context.Context,
	dbClient Client,
	rawSQLStmt string,
	args ...any, // variadic
) (*orm.Collection[MP, ID], error) {
	rows, err := dbClient.QueryRows(ctx, rawSQLStmt, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close() failed: %v", err)
		}
	}()
	return ScanRowsToCollection[M, MP, ID](rows)
}

func QueryCollectionByColumn[
	M any, // Model struct
	MP ScannableIdentifiable[M, ID], // *Model implementing ScannableIdentifiable[M, ID]
	ID comparable,
	V any,
](
	ctx context.Context,
	dbClient Client,
	sqlSelectBase string,
	column Column,
	values ...V,
) (*orm.Collection[MP, ID], error) {
	if len(values) == 0 {
		return nil, errors.New("empty values")
	}
	var (
		rows Rows
		err  error
	)
	if len(values) == 1 {
		sqlStmt := sqlSelectBase + fmt.Sprintf(" WHERE %s = %s", column.Name(), dbClient.SinglePlaceholder())
		rows, err = dbClient.QueryRows(ctx, sqlStmt, values[0])
	} else {
		sqlStmt := sqlSelectBase + fmt.Sprintf(" WHERE %s IN (%s)", column.Name(), dbClient.Placeholders(len(values)))
		valuesAsAny := make([]any, len(values))
		for i, v := range values {
			valuesAsAny[i] = v
		}
		rows, err = dbClient.QueryRows(ctx, sqlStmt, valuesAsAny...)
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close() failed: %v", err)
		}
	}()
	return ScanRowsToCollection[M, MP, ID](rows)
}
