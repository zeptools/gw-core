package orm

type Identifiable[ID comparable] interface {
	GetID() ID
}

type PointableIdentifiable[M any, ID comparable] interface {
	~*M
	Identifiable[ID]
}

func ModelPtrsToIDMap[
	MP Identifiable[ID], // *Model struct
	ID comparable,
](itemptrs []MP) map[ID]MP {
	idItemptrs := make(map[ID]MP, len(itemptrs))
	for _, p := range itemptrs {
		idItemptrs[p.GetID()] = p
	}
	return idItemptrs
}
