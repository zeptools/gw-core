package orm

import (
	"fmt"
)

type ModelCollection[MP Identifiable[ID], ID comparable] struct {
	ItemsMap   map[ID]MP
	OrderedIDs []ID // optional: only populated if you care about iteration order
}

func NewModelCollectionUnordered[
	P Identifiable[ID],
	ID comparable,
](items []P) *ModelCollection[P, ID] {
	coll := &ModelCollection[P, ID]{
		ItemsMap: make(map[ID]P, len(items)),
	}
	for _, item := range items {
		coll.ItemsMap[item.GetID()] = item
	}
	return coll
}

func NewModelCollectionOrdered[
	P Identifiable[ID],
	ID comparable,
](items []P) *ModelCollection[P, ID] {
	coll := &ModelCollection[P, ID]{
		ItemsMap:   make(map[ID]P, len(items)),
		OrderedIDs: make([]ID, len(items)),
	}
	for i, item := range items {
		id := item.GetID()
		coll.ItemsMap[id] = item
		coll.OrderedIDs[i] = id
	}
	return coll
}

func (c *ModelCollection[MP, ID]) Len() int {
	return len(c.ItemsMap)
}

func (c *ModelCollection[MP, ID]) Has(id ID) bool {
	_, ok := c.ItemsMap[id]
	return ok
}

func (c *ModelCollection[MP, ID]) Find(id ID) (MP, bool) {
	p, ok := c.ItemsMap[id]
	return p, ok
}

func (c *ModelCollection[MP, ID]) IDs() []ID {
	if len(c.OrderedIDs) > 0 {
		return append([]ID(nil), c.OrderedIDs...) // preserve original order
	}
	ids := make([]ID, 0, len(c.ItemsMap))
	for id := range c.ItemsMap {
		ids = append(ids, id)
	}
	return ids
}

func (c *ModelCollection[MP, ID]) IDsAsAny() []any {
	if len(c.OrderedIDs) > 0 {
		ids := make([]any, len(c.OrderedIDs))
		for i, id := range c.OrderedIDs {
			ids[i] = id
		}
		return ids
	}
	ids := make([]any, 0, len(c.ItemsMap))
	for id := range c.ItemsMap {
		ids = append(ids, id)
	}
	return ids
}

func (c *ModelCollection[MP, ID]) Items() []MP {
	if len(c.OrderedIDs) > 0 {
		items := make([]MP, 0, len(c.OrderedIDs))
		for _, id := range c.OrderedIDs {
			items = append(items, c.ItemsMap[id])
		}
		return items
	}
	items := make([]MP, 0, len(c.ItemsMap))
	for _, item := range c.ItemsMap {
		items = append(items, item)
	}
	return items
}

// ForEach calls fn for every model in the collection.
// If the collection has an order, it respects that order.
func (c *ModelCollection[MP, ID]) ForEach(fn func(MP)) {
	if len(c.OrderedIDs) > 0 {
		for _, id := range c.OrderedIDs {
			if mp, ok := c.ItemsMap[id]; ok {
				fn(mp)
			}
		}
		return
	}
	for _, mp := range c.ItemsMap {
		fn(mp)
	}
}

func (c *ModelCollection[MP, ID]) ForEachUnorderly(fn func(MP)) {
	for _, mp := range c.ItemsMap {
		fn(mp)
	}
}

func (c *ModelCollection[MP, ID]) ForEachOrderly(fn func(MP)) error {
	if len(c.OrderedIDs) == 0 {
		return fmt.Errorf("collection is unordered")
	}
	for _, id := range c.OrderedIDs {
		if mp, ok := c.ItemsMap[id]; ok {
			fn(mp)
		}
	}
	return nil
}

func (c *ModelCollection[MP, ID]) Filter(fn func(MP) bool) *ModelCollection[MP, ID] {
	// If ordered, keep the same order slice layout
	if len(c.OrderedIDs) > 0 {
		filtered := &ModelCollection[MP, ID]{
			ItemsMap:   make(map[ID]MP, len(c.ItemsMap)),
			OrderedIDs: make([]ID, 0, len(c.OrderedIDs)),
		}
		for _, id := range c.OrderedIDs {
			item := c.ItemsMap[id]
			if fn(item) {
				filtered.ItemsMap[id] = item
				filtered.OrderedIDs = append(filtered.OrderedIDs, id)
			}
		}
		return filtered
	}
	// Unordered — iterate directly over the map
	filtered := &ModelCollection[MP, ID]{
		ItemsMap: make(map[ID]MP, len(c.ItemsMap)),
	}
	for id, item := range c.ItemsMap {
		if fn(item) {
			filtered.ItemsMap[id] = item
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
		if v := yield(mp); v != nil {
			sl = append(sl, *v)
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
		if k, v := yield(mp); k != nil && v != nil {
			m[*k] = *v
		}
	})
	return m
}

// LinkOptionalBelongsTo connects ChildCollection-ParentCollection where aChild-BelongsTo-aParent
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
	for _, child := range children.ItemsMap {
		fkPtr := foreignKeyFieldPtr(child)
		if fkPtr == nil {
			continue
		}
		fk := *fkPtr
		if parent, ok := parents.ItemsMap[fk]; ok {
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
	for _, child := range children.ItemsMap {
		fk := foreignKey(child)
		parent, ok := parents.ItemsMap[fk]
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
	relationFieldPtr func(PP) *[]CP, // on the parent, slice
) {
	grouped := make(map[PID][]CP, len(parents.ItemsMap))
	for _, child := range children.ItemsMap {
		fk := foreignKey(child) // child's FK to parent
		grouped[fk] = append(grouped[fk], child)
	}
	for id, parent := range parents.ItemsMap {
		if kids, ok := grouped[id]; ok {
			*relationFieldPtr(parent) = kids
		} else {
			*relationFieldPtr(parent) = []CP{}
		}
	}
}
