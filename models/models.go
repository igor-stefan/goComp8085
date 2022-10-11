package models

type Label struct {
	Address int
	Nline   int
	Name    string
}

type Mnemonic struct {
	Start int
	End   int
	Nline int
	Name  string
}

type Instruction struct {
	Size   int
	Opcode string
}
