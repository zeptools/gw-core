package orm

import (
	"encoding/json/v2"
	"fmt"
)

type ModelCollection[MP Identifiable[ID], ID comparable] struct {
	itemsMap   map[ID]MP
	orderedIDs []ID // optional (default = nil). only populated if you care about iteration order
}

func NewEmptyOrderedModelCollection[
	P Identifiable[ID],
	ID comparable,
]() *ModelCollection[P, ID] {
	return &ModelCollection[P, ID]{
		itemsMap:   make(map[ID]P),
		orderedIDs: make([]ID, 0),
	}
}

func NewEmptyUnorderedModelCollection[
	P Identifiable[ID],
	ID comparable,
]() *ModelCollection[P, ID] {
	return &ModelCollection[P, ID]{
		itemsMap: make(map[ID]P),
	}
}

func NewModelCollectionUnordered[
	P Identifiable[ID],
	ID comparable,
](items []P) *ModelCollection[P, ID] {
	coll := &ModelCollection[P, ID]{
		itemsMap: make(map[ID]P, len(items)),
	}
	for _, item := range items {
		coll.itemsMap[item.GetID()] = item
	}
	return coll
}

func NewModelCollectionOrdered[
	P Identifiable[ID],
	ID comparable,
](items []P) *ModelCollection[P, ID] {
	coll := &ModelCollection[P, ID]{
		itemsMap:   make(map[ID]P, len(items)),
		orderedIDs: make([]ID, len(items)),
	}
	for i, item := range items {
		id := item.GetID()
		coll.itemsMap[id] = item
		coll.orderedIDs[i] = id
	}
	return coll
}

func (c *ModelCollection[MP, ID]) Len() int {
	return len(c.itemsMap)
}

func (c *ModelCollection[MP, ID]) Has(id ID) bool {
	_, ok := c.itemsMap[id]
	return ok
}

func (c *ModelCollection[MP, ID]) Find(id ID) (MP, bool) {
	p, ok := c.itemsMap[id]
	return p, ok
}

func (c *ModelCollection[MP, ID]) Add(item MP) {
	id := item.GetID()
	_, already := c.itemsMap[id]
	c.itemsMap[id] = item
	// Preserve order if ordered collection
	if c.orderedIDs != nil && !already {
		c.orderedIDs = append(c.orderedIDs, id)
	}
}

func (c *ModelCollection[MP, ID]) IDs() []ID {
	if c.orderedIDs != nil {
		return append([]ID(nil), c.orderedIDs...) // preserve original order
	}
	ids := make([]ID, 0, len(c.itemsMap))
	for id := range c.itemsMap {
		ids = append(ids, id)
	}
	return ids
}

func (c *ModelCollection[MP, ID]) IDsAsAny() []any {
	if len(c.orderedIDs) > 0 {
		ids := make([]any, len(c.orderedIDs))
		for i, id := range c.orderedIDs {
			ids[i] = id
		}
		return ids
	}
	ids := make([]any, 0, len(c.itemsMap))
	for id := range c.itemsMap {
		ids = append(ids, id)
	}
	return ids
}

func (c *ModelCollection[MP, ID]) Items() []MP {
	if len(c.orderedIDs) > 0 {
		items := make([]MP, 0, len(c.orderedIDs))
		for _, id := range c.orderedIDs {
			items = append(items, c.itemsMap[id])
		}
		return items
	}
	items := make([]MP, 0, len(c.itemsMap))
	for _, item := range c.itemsMap {
		items = append(items, item)
	}
	return items
}

func (c *ModelCollection[MP, ID]) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}
	return json.Marshal(c.Items())
}

// ForEach calls fn for every model in the collection.
// If the collection has an order, it respects that order.
func (c *ModelCollection[MP, ID]) ForEach(fn func(MP)) {
	if len(c.orderedIDs) > 0 {
		for _, id := range c.orderedIDs {
			if mp, ok := c.itemsMap[id]; ok {
				fn(mp)
			}
		}
		return
	}
	for _, mp := range c.itemsMap {
		fn(mp)
	}
}

func (c *ModelCollection[MP, ID]) ForEachUnorderly(fn func(MP)) {
	for _, mp := range c.itemsMap {
		fn(mp)
	}
}

func (c *ModelCollection[MP, ID]) ForEachOrderly(fn func(MP)) error {
	if len(c.orderedIDs) == 0 {
		return fmt.Errorf("collection is unordered")
	}
	for _, id := range c.orderedIDs {
		if mp, ok := c.itemsMap[id]; ok {
			fn(mp)
		}
	}
	return nil
}

func (c *ModelCollection[MP, ID]) Filter(fn func(MP) bool) *ModelCollection[MP, ID] {
	// If ordered, keep the same order slice layout
	if len(c.orderedIDs) > 0 {
		filtered := &ModelCollection[MP, ID]{
			itemsMap:   make(map[ID]MP, len(c.itemsMap)),
			orderedIDs: make([]ID, 0, len(c.orderedIDs)),
		}
		for _, id := range c.orderedIDs {
			item := c.itemsMap[id]
			if fn(item) {
				filtered.itemsMap[id] = item
				filtered.orderedIDs = append(filtered.orderedIDs, id)
			}
		}
		return filtered
	}
	// Unordered — iterate directly over the map
	filtered := &ModelCollection[MP, ID]{
		itemsMap: make(map[ID]MP, len(c.itemsMap)),
	}
	for id, item := range c.itemsMap {
		if fn(item) {
			filtered.itemsMap[id] = item
		}
	}
	return filtered
}

// EnumerateToSlice iterates over every model in the collection and calls yield for each.
// Every model contributes exactly one value. No skipping.
// Conceptually equivalent to: [yield(m) for m in c].
func EnumerateToSlice[
	MP Identifiable[ID],
	ID comparable,
	V any,
](
	c *ModelCollection[MP, ID],
	yield func(MP) V,
) []V {
	sl := make([]V, 0, c.Len()) // new slice
	c.ForEach(func(mp MP) {
		// we don't mutate, but yield can. caller's responsibility
		sl = append(sl, yield(mp))
	})
	return sl
}

// EnumerateToMap iterates over every model in the collection and calls yield for each.
// Every model contributes exactly one key–value pair. No skipping.
// Conceptually equivalent to: {k: v for m in c}.
func EnumerateToMap[
	MP Identifiable[ID],
	ID comparable,
	K comparable,
	V any,
](
	c *ModelCollection[MP, ID],
	yield func(MP) (K, V),
) map[K]V {
	m := make(map[K]V, c.Len()) // new map
	c.ForEachUnorderly(func(mp MP) {
		// we don't mutate, but yield can. caller's responsibility
		k, v := yield(mp)
		m[k] = v
	})
	return m
}

// CollectToSlice iterates over the collection and calls yield for each model.
// If yield returns nil, the element is skipped (conditional yield).
// Returns a slice of yielded values.
// Equivalent to a list comprehension: [yield(m) for m in c if yield(m) != nil].
func CollectToSlice[
	MP Identifiable[ID],
	ID comparable,
	V any,
](
	c *ModelCollection[MP, ID],
	yield func(MP) *V,
) []V {
	sl := make([]V, 0, c.Len()) // new slice
	c.ForEach(func(mp MP) {
		// we don't mutate, but yield can. caller's responsibility
		if vp := yield(mp); vp != nil {
			sl = append(sl, *vp)
		}
	})
	return sl
}

// CollectToMap iterates over the collection and calls yield for each model.
// If yield returns nil, the element is skipped (conditional yield).
// The yielded key–value pair determines each map entry.
func CollectToMap[
	MP Identifiable[ID],
	ID comparable,
	K comparable,
	V any,
](
	c *ModelCollection[MP, ID],
	yield func(MP) (*K, *V),
) map[K]V {
	m := make(map[K]V, c.Len()) // new map
	c.ForEachUnorderly(func(mp MP) {
		// we don't mutate, but yield can. caller's responsibility
		if kp, vp := yield(mp); kp != nil && vp != nil {
			m[*kp] = *vp
		}
	})
	return m
}

func CollectUniqueToSlice[
	MP Identifiable[ID],
	ID comparable,
	V comparable,
](
	c *ModelCollection[MP, ID],
	yield func(MP) *V,
) []V {
	sl := make([]V, 0, c.Len())
	uniqueCollectedAsKeys := make(map[V]struct{}, c.Len())

	if len(c.orderedIDs) > 0 {
		// Ordered iteration: preserve first-occurrence order
		for _, id := range c.orderedIDs {
			item, ok := c.itemsMap[id]
			if !ok {
				continue
			}
			if vp := yield(item); vp != nil {
				v := *vp
				if _, exists := uniqueCollectedAsKeys[v]; !exists {
					uniqueCollectedAsKeys[v] = struct{}{}
					sl = append(sl, v)
				}
			}
		}
		return sl
	}

	// Unordered iteration
	for _, item := range c.itemsMap {
		if vp := yield(item); vp != nil {
			v := *vp
			if _, exists := uniqueCollectedAsKeys[v]; !exists {
				uniqueCollectedAsKeys[v] = struct{}{}
				sl = append(sl, v)
			}
		}
	}
	return sl
}

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
	children *ModelCollection[CP, CID],
	parents *ModelCollection[PP, PID],
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
	children *ModelCollection[CP, CID],
	parents *ModelCollection[PP, PID],
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
	parents *ModelCollection[PP, PID],
	children *ModelCollection[CP, CID],
	foreignKey func(CP) PID, // on the child
	relationFieldPtr func(PP) **ModelCollection[CP, CID], // on the parent, slice
) {
	childCollGrpByPID := make(map[PID]*ModelCollection[CP, CID], parents.Len())
	for _, child := range children.itemsMap {
		pid := foreignKey(child) // child's FK to parent id
		childColl, ok := childCollGrpByPID[pid]
		if !ok {
			childColl = NewEmptyOrderedModelCollection[CP, CID]()
			childCollGrpByPID[pid] = childColl
		}
		childColl.Add(child)
	}
	for pid, parent := range parents.itemsMap {
		if childColl, ok := childCollGrpByPID[pid]; ok {
			*relationFieldPtr(parent) = childColl
		} else {
			*relationFieldPtr(parent) = NewEmptyOrderedModelCollection[CP, CID]()
		}
	}
}
