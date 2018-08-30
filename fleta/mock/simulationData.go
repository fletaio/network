package simulationdata

//node type
const (
	ObserverNodeCount   = 5
	FormulatorNodeCount = 4000
	NormalNodeCount     = 0
)

//simulation init count of node
const (
	InitNodeCount = ObserverNodeCount + 1
)

//node start index
const (
	ObserverNodeStartIndex   = 0
	FormulatorNodeStartIndex = ObserverNodeCount
	NormalNodeStartIndex     = FormulatorNodeStartIndex + FormulatorNodeCount
)

//mocknet delay param
const (
	Delay     = true
	DelayUnit = 1 // 1(0~32) 2(0~64) ~
)
