package comparer

type Ider interface {
	GetId() uint64
}

func CompareSlice[T1, T2 Ider](current []T1, mod []T2) ([]T1, []T2, []T1) {
	// Delete
	currentMap := make(map[uint64]struct{})
	modMap := make(map[uint64]struct{})
	for _, elem := range current {
		currentMap[elem.GetId()] = struct{}{}
	}
	for _, elem := range mod {
		modMap[elem.GetId()] = struct{}{}
	}

	// DELETE ( которые есть в current но нет в mod )
	var del []T1
	for _, elem := range current {
		if _, ok := modMap[elem.GetId()]; !ok {
			del = append(del, elem)
		}
	}

	// CREATE ( которые есть в mod но нет в current )
	var create []T2
	for _, elem := range mod {
		if _, ok := currentMap[elem.GetId()]; !ok {
			create = append(create, elem)
		}
	}

	// MOD ( которые есть в mod И current )
	var modify []T1
	for _, elem := range current {
		if _, ok := modMap[elem.GetId()]; ok {
			modify = append(modify, elem)
		}
	}
	return del, create, modify
}
