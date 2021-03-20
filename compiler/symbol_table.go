package compiler

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer          *SymbolTable
	store          map[string]Symbol
	FreeSymbols    []Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
}

func NewEnclosedSymbolTable(st *SymbolTable) *SymbolTable {
	symTable := NewSymbolTable()
	symTable.Outer = st
	return symTable
}

func (s *SymbolTable) Define(name string) Symbol {
	sym := Symbol{
		Name:  name,
		Scope: GlobalScope,
		Index: s.numDefinitions,
	}

	if s.Outer != nil {
		sym.Scope = LocalScope
	}

	s.store[name] = sym
	s.numDefinitions++

	return sym
}

func (s *SymbolTable) DefineBuiltin(i int, name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Scope: BuiltinScope,
		Index: i,
	}

	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Scope: FunctionScope,
		Index: 0,
	}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	sym, ok := s.store[name]
	if !ok && s.Outer != nil {
		sym, ok = s.Outer.Resolve(name)
		if !ok {
			return sym, ok
		}

		if sym.Scope == GlobalScope || sym.Scope == BuiltinScope {
			return sym, ok
		}

		free := s.defineFree(sym)
		return free, true
	}

	return sym, ok
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)

	symbol := Symbol{
		Name:  original.Name,
		Scope: FreeScope,
		Index: len(s.FreeSymbols) - 1,
	}

	s.store[original.Name] = symbol
	return symbol
}
