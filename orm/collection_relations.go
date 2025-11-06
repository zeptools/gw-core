package orm

import "fmt"

// LinkOptionalBelongsTo connects ChildCollection-ParentCollection where Child-BelongsTo-Parent
// ForeignKeyField is on the Child
// RelationField is on the Child
// Optional Version
func LinkOptionalBelongsTo[
	CP Identifiable[CID],
	CID comparable,
	PP Identifiable[PID],
	PID comparable,
](
	children *Collection[CP, CID],
	parents *Collection[PP, PID],
	foreignKeyFieldPtr func(CP) *PID, // on the child
	relationFieldPtr func(CP) *PP, // on the child
) {
	for _, child := range children.itemsMap {
		fkPtr := foreignKeyFieldPtr(child)
		if fkPtr == nil {
			continue
		}
		fk := *fkPtr
		if parent, ok := parents.itemsMap[fk]; ok {
			*relationFieldPtr(child) = parent
		}
	}
}

// LinkBelongsTo - Strict Version
// ForeignKeyField is on the Child
// RelationField is on the Child
func LinkBelongsTo[
	CP Identifiable[CID],
	CID comparable,
	PP Identifiable[PID],
	PID comparable,
](
	children *Collection[CP, CID],
	parents *Collection[PP, PID],
	foreignKey func(CP) PID, // on the child
	relationFieldPtr func(CP) *PP, // on the child
) error {
	for _, child := range children.itemsMap {
		fk := foreignKey(child)
		parent, ok := parents.itemsMap[fk]
		if !ok {
			return fmt.Errorf(
				"LinkBelongsTo: parent with ID %v not found for child ID %v",
				fk, child.GetID(),
			)
		}
		*relationFieldPtr(child) = parent
	}
	return nil
}

// LinkHasMany connects ParentCollection-ChildCollection where a Parent-HasMany-Children
// ForeignKeyField is on the Child
// RelationField (a Slice) is on the Parent
func LinkHasMany[
	PP Identifiable[PID],
	PID comparable,
	CP Identifiable[CID],
	CID comparable,
](
	parents *Collection[PP, PID],
	children *Collection[CP, CID],
	foreignKey func(CP) PID, // on the child
	relationFieldPtr func(PP) **Collection[CP, CID], // on the parent
) {
	childCollGrpByPID := make(map[PID]*Collection[CP, CID], parents.Len())
	for _, child := range children.itemsMap {
		pid := foreignKey(child) // child's FK to parent id
		childColl, ok := childCollGrpByPID[pid]
		if !ok {
			childColl = NewEmptyOrderedCollection[CP, CID]()
			childCollGrpByPID[pid] = childColl
		}
		childColl.Add(child)
	}
	for pid, parent := range parents.itemsMap {
		if childColl, ok := childCollGrpByPID[pid]; ok {
			*relationFieldPtr(parent) = childColl
		} else {
			*relationFieldPtr(parent) = NewEmptyOrderedCollection[CP, CID]()
		}
	}
}
