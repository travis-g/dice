package dice

var _ = Roller(&PolyhedralDie{})

var dieSets = []struct {
	size  int
	count int
}{
	{6, 2},
	{20, 2},
	{100, 2},
}
