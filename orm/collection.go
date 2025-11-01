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

func (c *ModelCollection[MP, ID]) ForEach(fn func(MP)) {
	for _, item := range c.Map {
		fn(item)
	}
}

func (c *ModelCollection[MP, ID]) ForEachOrdered(fn func(MP)) error {
	if len(c.Order) == 0 {
		return fmt.Errorf("collection is unordered")
	}
	for _, id := range c.Order {
		fn(c.Map[id])
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

func Pluck[
MP Identifiable[ID],
ID comparable,
V any,
](
	c *ModelCollection[MP, ID],
	fieldPtr func(MP) *V,
) []V {
	sl := make([]V, 0, c.Len())
	if len(c.Order) > 0 {
		for _, id := range c.Order {
			mp, ok := c.Map[id]
			if !ok {
				continue
			}
			vp := fieldPtr(mp)
			if vp == nil {
				continue
			}
			sl = append(sl, *vp)
		}
		return sl
	}
	for _, mp := range c.Map {
		vp := fieldPtr(mp)
		if vp == nil {
			continue
		}
		sl = append(sl, *vp)
	}
	return sl
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
