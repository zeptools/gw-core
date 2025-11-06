//go:build !(debug && verbose)

package sqldb

import (
	"context"
	"fmt"

	"github.com/zeptools/gw-core/orm"
)

// LoadBelongsTo - Load Parents on Children from SQL DB and Link Child-BelongsTo-Parent Relation
// Returns the Parents
func LoadBelongsTo[
CP orm.Identifiable[CID],
CID comparable,
P any, // Model struct
PP ScannableIdentifiable[P, PID],
PID comparable,
](
	ctx context.Context,
	dbClient Client,
	children *orm.Collection[CP, CID],
	sqlSelectBase string,
	foreignKey func(c CP) PID,
	relationFieldPtr func(c CP) *PP,
) (*orm.Collection[PP, PID], error) {
	fKeysAsAny := orm.EnumerateToSlice(children, func(c CP) any { return foreignKey(c) })
	sqlStmt := sqlSelectBase + fmt.Sprintf(" WHERE id IN (%s)", dbClient.Placeholders(len(fKeysAsAny)))
	parents, err := RawQueryCollection[P, PP, PID](ctx, dbClient, sqlStmt, fKeysAsAny...)
	if err != nil {
		return nil, err
	}
	err = orm.LinkBelongsTo[CP, CID, PP, PID](children, parents, foreignKey, relationFieldPtr)
	if err != nil {
		return nil, err
	}
	return parents, nil
}

func LoadHasMany[
PP orm.Identifiable[PID],
PID comparable,
C any, // Model struct
CP ScannableIdentifiable[C, CID],
CID comparable,
](
	ctx context.Context,
	dbClient Client,
	parents *orm.Collection[PP, PID],
	sqlSelectBase string,
	foreignKeyColumn Column,                             // on the child
	foreignKey func(CP) PID,                             // on the child
	relationFieldPtr func(PP) **orm.Collection[CP, CID], // on the parent
	orderBy ...OrderBy,
) (*orm.Collection[CP, CID], error) {
	whereClause := fmt.Sprintf(" WHERE %s IN (%s)", foreignKeyColumn.Name(), dbClient.Placeholders(parents.Len()))
	sqlStmt := sqlSelectBase + whereClause + OrderByClause(orderBy)
	parentIDsAsAny := parents.IDsAsAny()
	children, err := RawQueryCollection[C, CP, CID](ctx, dbClient, sqlStmt, parentIDsAsAny...)
	if err != nil {
		return nil, err
	}
	orm.LinkHasMany[PP, PID, CP, CID](
		parents,
		children,
		foreignKey,
		relationFieldPtr,
	)
	return children, nil
}
