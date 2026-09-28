package slugbug

type Operation string

var AvailableOperations = []Operation{
	Invoke,
	Inspect,
	Monitor,
}

func (o Operation) String() string {
	return string(o)
}
