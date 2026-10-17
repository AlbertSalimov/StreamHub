package content

type ListFilter struct {
	Search string
	Genre  string
	Year   int
	Sort   string
	Order  string
	Limit  int
	Offset int
}
