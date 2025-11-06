package sqldb

import (
	"fmt"

	"github.com/zeptools/gw-core/orm"
)

func ScanRowToItem[
	M any, // Model struct
	MP Scannable[M], // *Model Implementing Scannable[M]
](row Row) (*M, error) { // Returns the Pointer to the Newly Created Item
	var item M     // struct with zero values for the fields
	p := MP(&item) // p is *M, which satisfies scanFieldsProvider interface
	err := row.Scan(p.FieldsToScan()...)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func ScanRowsToItems[
	M any, // Model struct
	MP Scannable[M], // *Model Implementing Scannable[M]
](rows Rows) ([]*M, error) { // Returns a Slice of Model-Pointers
	var itemptrs []*M
	for rows.Next() {
		var item M     // struct with zero values for the fields
		p := MP(&item) // p is *M, which satisfies scanFieldsProvider interface
		// Scan the Fields of Each Row to the Fields of the new struct of the Model
		if err := rows.Scan(p.FieldsToScan()...); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		itemptrs = append(itemptrs, &item) // Collect the pointers
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows: %v", err)
	}
	return itemptrs, nil
}

// ScanRowsToMap scan rows to a map[id]item
func ScanRowsToMap[
	M any, // Model struct
	MP ScannableIdentifiable[M, ID], // *Model Implementing ScannableIdentifiable[M, ID]
	ID comparable,
](rows Rows) (map[ID]*M, error) { // Returns a ItemsMap of ID to Model-Pointers
	idItemptrs := map[ID]*M{}
	for rows.Next() {
		var item M     // struct with zero values for the fields
		p := MP(&item) // p is *M, which satisfies scanFieldsProvider interface
		// Scan the Fields of Each Row to the Fields of the new struct of the Model
		if err := rows.Scan(p.FieldsToScan()...); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		idItemptrs[p.GetID()] = &item
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows: %v", err)
	}
	return idItemptrs, nil
}

// ScanRowsToCollection scan rows to a collection
func ScanRowsToCollection[
	M any, // Model struct
	MP ScannableIdentifiable[M, ID], // *Model implementing ScannableIdentifiable[M, ID]
	ID comparable,
](
	rows Rows,
) (*orm.Collection[MP, ID], error) {
	coll := orm.NewEmptyOrderedCollection[MP, ID]()

	for rows.Next() {
		var item M
		p := MP(&item) // *M implementing ScannableIdentifiable
		if err := rows.Scan(p.FieldsToScan()...); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		coll.Add(p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iterating rows: %v", err)
	}
	return coll, nil
}
