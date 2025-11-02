package orm

import (
	"fmt"
)

type ModelCollection[MP Identifiable[ID], ID comparable] struct {
	Map   map[ID]MP
	Order []ID // optional: only populated if you care about iteration order
}

func NewModelCollectionUnordered[
P Identifiable[ID],
ID comparable,
](items []P) *ModelCollection[P, ID] {
	coll := &ModelCollection[P, ID]{
		Map: make(map[ID]P, len(items)),
	}
	for _, item := range items {
		coll.Map[item.GetID()] = item
	}
	return coll
}

func NewModelCollectionOrdered[
P Identifiable[ID],
ID comparable,
](items []P) *ModelCollection[P, ID] {
	coll := &ModelCollection[P, ID]{
		Map:   make(map[ID]P, len(items)),
		Order: make([]ID, len(items)),
	}
	for i, item := range items {
		id := item.GetID()
		coll.Map[id] = item
		coll.Order[i] = id
	}
	return coll
}

func (c *ModelCollection[MP, ID]) Len() int {
	return len(c.Map)
}

func (c *ModelCollection[MP, ID]) Has(id ID) bool {
	_, ok := c.Map[id]
	return ok
}

func (c *ModelCollection[MP, ID]) Find(id ID) (MP, bool) {
	p, ok := c.Map[id]
	return p, ok
}

func (c *ModelCollection[MP, ID]) IDs() []ID {
	if len(c.Order) > 0 {
		return append([]ID(nil), c.Order...) // preserve original order
	}
	ids := make([]ID, 0, len(c.Map))
	for id := range c.Map {
		ids = append(ids, id)
	}
	return ids
}

func (c *ModelCollection[MP, ID]) IDsAsAny() []any {
	if len(c.Order) > 0 {
		ids := make([]any, len(c.Order))
		for i, id := range c.Order {
			ids[i] = id
		}
		return ids
	}
	ids := make([]any, 0, len(c.Map))
	for id := range c.Map {
		ids = append(ids, id)
	}
	return ids
}

func (c *ModelCollection[MP, ID]) Items() []MP {
	if len(c.Order) > 0 {
		items := make([]MP, 0, len(c.Order))
		for _, id := range c.Order {
			items = append(items, c.Map[id])
		}
		return items
	}
	items := make([]MP, 0, len(c.Map))
	for _, item := range c.Map {
		items = append(items, item)
	}
	return items
}

// ForEach calls fn for every model in the collection.
// If the collection has an order, it respects that order.
func (c *ModelCollection[MP, ID]) ForEach(fn func(MP)) {
	if len(c.Order) > 0 {
		for _, id := range c.Order {
			if mp, ok := c.Map[id]; ok {
				fn(mp)
			}
		}
		return
	}
	for _, mp := range c.Map {
		fn(mp)
	}
}

func (c *ModelCollection[MP, ID]) ForEachUnorderly(fn func(MP)) {
	for _, mp := range c.Map {
		fn(mp)
	}
}

func (c *ModelCollection[MP, ID]) ForEachOrderly(fn func(MP)) error {
	if len(c.Order) == 0 {
		return fmt.Errorf("collection is unordered")
	}
	for _, id := range c.Order {
		if mp, ok := c.Map[id]; ok {
			fn(mp)
		}
	}
	return nil
}

func (c *ModelCollection[MP, ID]) Filter(fn func(MP) bool) *ModelCollection[MP, ID] {
	// If ordered, keep the same order slice layout
	if len(c.Order) > 0 {
		filtered := &ModelCollection[MP, ID]{
			Map:   make(map[ID]MP, len(c.Map)),
			Order: make([]ID, 0, len(c.Order)),
		}
		for _, id := range c.Order {
			item := c.Map[id]
			if fn(item) {
				filtered.Map[id] = item
				filtered.Order = append(filtered.Order, id)
			}
		}
		return filtered
	}
	// Unordered — iterate directly over the map
	filtered := &ModelCollection[MP, ID]{
		Map: make(map[ID]MP, len(c.Map)),
	}
	for id, item := range c.Map {
		if fn(item) {
			filtered.Map[id] = item
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
	relationFieldPtr func(CP) *PP,    // on the child
) {
	for _, child := range children.Map {
		fkPtr := foreignKeyFieldPtr(child)
		if fkPtr == nil {
			continue
		}
		fk := *fkPtr
		if parent, ok := parents.Map[fk]; ok {
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
	foreignKeyFieldPtr func(CP) *PID, // on the child
	relationFieldPtr func(CP) *PP,    // on the child
) error {
	for _, child := range children.Map {
		fkPtr := foreignKeyFieldPtr(child)
		if fkPtr == nil {
			return fmt.Errorf(
				"LinkBelongsTo: nil foreign key for child ID %v",
				child.GetID(),
			)
		}
		fk := *fkPtr
		parent, ok := parents.Map[fk]
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
	foreignKeyFieldPtr func(CP) *PID, // on the child
	relationFieldPtr func(PP) *[]CP,  // on the parent, slice
) {
	grouped := make(map[PID][]CP, len(parents.Map))
	for _, child := range children.Map {
		fk := *foreignKeyFieldPtr(child)
		grouped[fk] = append(grouped[fk], child)
	}
	for id, parent := range parents.Map {
		if kids, ok := grouped[id]; ok {
			*relationFieldPtr(parent) = kids
		} else {
			*relationFieldPtr(parent) = []CP{}
		}
	}
}
