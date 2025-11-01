package orm

type Identifiable[ID comparable] interface {
	comparable
	GetID() ID
}

func ModelPtrsToIDMap[
	P Identifiable[ID], // *Model struct
	ID comparable,
](itemptrs []P) map[ID]P {
	idItemptrs := make(map[ID]P, len(itemptrs))
	for _, p := range itemptrs {
		idItemptrs[p.GetID()] = p
	}
	return idItemptrs
}
