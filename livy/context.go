package livy

type Context struct {
	Globals  map[string]Val
	Monadics map[string]MonadicFunc
}
