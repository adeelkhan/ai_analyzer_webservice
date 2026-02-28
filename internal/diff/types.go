package diff

type ChangeType int

const (
	Add ChangeType = iota
	Remove
	Modify
	Unchanged
)

type Change struct {
	Line     string
	Type     ChangeType
	OldLine  int
	NewLine  int
	Position int
}

type DiffResult struct {
	Changes []Change
	Stats   Stats
}

type Stats struct {
	Additions int
	Removals  int
	Modifies  int
	Unchanged int
}

func (s *Stats) Calculate(changes []Change) {
	s.Additions = 0
	s.Removals = 0
	s.Modifies = 0
	s.Unchanged = 0

	for _, c := range changes {
		switch c.Type {
		case Add:
			s.Additions++
		case Remove:
			s.Removals++
		case Modify:
			s.Modifies++
		case Unchanged:
			s.Unchanged++
		}
	}
}

func (ct ChangeType) String() string {
	switch ct {
	case Add:
		return "Add"
	case Remove:
		return "Remove"
	case Modify:
		return "Modify"
	case Unchanged:
		return "Unchanged"
	default:
		return "Unknown"
	}
}
