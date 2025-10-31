package orm

import "fmt"

type ModelCollection[MP Identifiable[ID], ID comparable] struct {
	Map   map[ID]MP
	Order []ID // optional: only populated if you care about iteration order
}

func NewModelCollectionUnordered[P Identifiable[ID], ID comparable](items []P) *ModelCollection[P, ID] {
	coll := &ModelCollection[P, ID]{
		Map: make(map[ID]P, len(items)),
	}
	for _, item := range items {
		coll.Map[item.GetID()] = item
	}
	return coll
}

func NewModelCollectionOrdered[P Identifiable[ID], ID comparable](items []P) *ModelCollection[P, ID] {
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

func LinkBelongsTo[
CP Identifiable[CID], CID comparable,
PP Identifiable[PID], PID comparable,
](
	children *ModelCollection[CP, CID],
	parents *ModelCollection[PP, PID],
	foreignKeyFieldPtr func(CP) *PID, // on the child
	relationFieldPtr func(CP) *PP,    // on the child
) {
	for _, child := range children.Map {
		if parent, ok := parents.Map[*foreignKeyFieldPtr(child)]; ok {
			*relationFieldPtr(child) = parent
		}
	}
}
