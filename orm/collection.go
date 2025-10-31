package orm

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

func (c *ModelCollection[MP, ID]) Find(id ID) (MP, bool) {
	p, ok := c.Map[id]
	return p, ok
}

func LinkBelongsTo[
	CP Identifiable[CID], CID comparable,
	PP Identifiable[PID], PID comparable,
](
	children *ModelCollection[CP, CID],
	parents *ModelCollection[PP, PID],
	foreignKeyFieldPtr func(CP) *PID, // on the child
	relationFieldPtr func(CP) *PP, // on the child
) {
	for _, child := range children.Map {
		if parent, ok := parents.Map[*foreignKeyFieldPtr(child)]; ok {
			*relationFieldPtr(child) = parent
		}
	}
}
