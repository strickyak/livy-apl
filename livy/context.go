package livy

type Context struct {
	Globals    map[string]Val
	Monadics   map[string]MonadicFunc
	Dyadics    map[string]DyadicFunc
	LocalStack []map[string]Val
}
