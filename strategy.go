package lseq

import "math/rand"

type strategy uint8

// StrategyMap -
// Which strategy was picked at a given digit depth.
type StrategyMap map[uint8]strategy

const (
	undefinedStrategy strategy = iota
	boundaryLoStrategy
	boundaryHiStrategy
	stategyCount
)

// get --
// Return the stategy for "depth", if needed by picking a random one and
// updating the map.
func (m StrategyMap) get(depth uint8) strategy {
	s, ok := m[depth]
	if !ok {
		s = strategy(rand.Intn(2) + 1)
		m[depth] = s
	}
	return s
}
