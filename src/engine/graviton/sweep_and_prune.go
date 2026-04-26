package graviton

import (
	"sort"

	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/klib/bitmap"
	"kaijuengine.com/matrix"
)

// Interval represents a projected AABB onto one axis
type Interval struct {
	Min  matrix.Float
	Max  matrix.Float
	Body *RigidBody
}

// ActivePair represents a potential collision pair found by SAP
type ActivePair struct {
	BodyA *RigidBody
	BodyB *RigidBody
}

type SweepPrune struct {
	// One sorted list per axis
	intervals [3][]Interval
	// Pairs that overlap on all three axes
	activePairs []ActivePair
	// Reusable buffer to avoid allocations
	pairBuffer []ActivePair
	// Pre-allocated buffers for sweep to avoid alloc
	activeListBuff [3][]activeEntry
	// Cached sweep vars for re-use
	sweepCandidates   [32]byte
	sweepCandidateIds [256]int
}

type activeEntry struct {
	body int
	max  matrix.Float
}

type axisBounds struct {
	min matrix.Float
	max matrix.Float
}

func (s *SweepPrune) Initialize(initialBodyCount int) {
	for i := range s.intervals {
		s.intervals[i] = make([]Interval, 0, initialBodyCount)
	}
	s.activePairs = make([]ActivePair, 0, initialBodyCount*2)
	s.pairBuffer = make([]ActivePair, 0, initialBodyCount*2)
	for i := range s.activeListBuff {
		s.activeListBuff[i] = make([]activeEntry, 0, initialBodyCount)
	}
}

func (s *SweepPrune) Rebuild(bodies *pooling.PoolGroup[RigidBody]) {
	for i := range s.intervals {
		s.intervals[i] = s.intervals[i][:0]
	}
	bodies.Each(func(body *RigidBody) {
		if !body.Active {
			return
		}
		worldAABB := body.WorldAABB()
		s.intervals[AxisX] = append(s.intervals[AxisX], Interval{
			Min:  matrix.Float(worldAABB.Min().X()),
			Max:  matrix.Float(worldAABB.Max().X()),
			Body: body,
		})
		s.intervals[AxisY] = append(s.intervals[AxisY], Interval{
			Min:  matrix.Float(worldAABB.Min().Y()),
			Max:  matrix.Float(worldAABB.Max().Y()),
			Body: body,
		})
		s.intervals[AxisZ] = append(s.intervals[AxisZ], Interval{
			Min:  matrix.Float(worldAABB.Min().Z()),
			Max:  matrix.Float(worldAABB.Max().Z()),
			Body: body,
		})
	})
	for i := range s.intervals {
		sort.Slice(s.intervals[i], func(a, b int) bool {
			return s.intervals[i][a].Min < s.intervals[i][b].Min
		})
	}
}

func (s *SweepPrune) Sweep() []ActivePair {
	s.pairBuffer = s.pairBuffer[:0]
	var activeLists [3][]activeEntry
	for i := range activeLists {
		activeLists[i] = s.activeListBuff[i][:0]
	}
	candidates := bitmap.Bitmap(s.sweepCandidates[:])
	var sweepNumCandidates int
	// Initialize candidates from X axis
	for _, interval := range s.intervals[AxisX] {
		id := interval.Body.poolLocation()
		candidates.Set(id)
		s.sweepCandidateIds[sweepNumCandidates] = id
		sweepNumCandidates++
	}
	// Intersect candidates across all three axes
	for axis := AxisX; axis <= AxisZ; axis++ {
		active := activeLists[axis]
		intervals := s.intervals[axis]
		// Reset candidates for this axis intersection
		axisCandidates := bitmap.New(256)
		var axisCandidateIds [256]int
		var numAxisCandidates int
		for _, interval := range intervals {
			id := interval.Body.poolLocation()
			// Remove entries from active list whose max < current min
			validActive := active[:0]
			for _, entry := range active {
				if entry.max >= interval.Min {
					validActive = append(validActive, entry)
					// This body overlaps on this axis, check if it's a candidate
					if candidates.IsSet(entry.body) {
						axisCandidates.Set(entry.body)
						axisCandidateIds[numAxisCandidates] = entry.body
						numAxisCandidates++
					}
				}
			}
			active = validActive
			active = append(active, activeEntry{
				body: id,
				max:  interval.Max,
			})
			axisCandidates.Set(id)
			axisCandidateIds[numAxisCandidates] = id
			numAxisCandidates++
		}
		activeLists[axis] = active
		candidates = axisCandidates
		s.sweepCandidateIds = axisCandidateIds
		sweepNumCandidates = numAxisCandidates
	}
	var bounds [256]axisBounds
	for axis := AxisX; axis <= AxisZ; axis++ {
		for _, interval := range s.intervals[axis] {
			bounds[interval.Body.poolLocation()] = axisBounds{
				min: interval.Min,
				max: interval.Max,
			}
		}
	}
	for i := range sweepNumCandidates {
		idA := s.sweepCandidateIds[i]
		for j := i + 1; j < sweepNumCandidates; j++ {
			idB := s.sweepCandidateIds[j]
			if s.overlapsOnAllAxes(bounds, idA, idB) {
				s.pairBuffer = append(s.pairBuffer, ActivePair{
					BodyA: s.intervals[AxisX][0].Body, // TODO: store body reference
					BodyB: s.intervals[AxisX][0].Body, // TODO: store body reference
				})
			}
		}
	}
	return s.pairBuffer
}

func (s *SweepPrune) overlapsOnAllAxes(bounds [256]axisBounds, idA, idB int) bool {
	for axis := AxisX; axis <= AxisZ; axis++ {
		a := bounds[idA]
		b := bounds[idB]
		if a.max < b.min || b.max < a.min {
			return false
		}
	}
	return true
}
