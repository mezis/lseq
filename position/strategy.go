package position

import "math/rand"

type strategy uint8

// StrategyMap -
// Which strategy was picked at a given digit depth.
type StrategyMap [maxDigits]strategy

const (
	UndefinedStrategy strategy = iota
	boundaryLoStrategy
	boundaryHiStrategy
)
const strategyCount = int(boundaryHiStrategy)

// Get --
// Return the stategy for "depth", if needed by picking a random one and
// updating the map.
func (m *StrategyMap) Get(depth uint8) strategy {
	s := m[depth]
	if s == UndefinedStrategy {
		s = strategy(rand.Intn(strategyCount) + 1)
		m[depth] = s
	}
	return s
}
