package graviton

import (
	"sync"

	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

const (
	defaultVelocityIterations = 8
	defaultPositionIterations = 3
	defaultRestitution        = matrix.Float(0.05)
	defaultStaticFriction     = matrix.Float(0.6)
	defaultDynamicFriction    = matrix.Float(0.45)
	defaultBaumgarte          = matrix.Float(0.8)
	defaultPenetrationSlop    = matrix.Float(0.005)
	defaultMaxCorrection      = matrix.Float(0.25)
	solverMinIslandsPerJob    = 2
)

type collisionIsland struct {
	manifolds []int
}

// CollisionSolver resolves narrow-phase contacts with an iterative impulse
// solver. Contacts are grouped into independent dynamic islands so islands can
// be solved in parallel without concurrent writes to the same body.
type CollisionSolver struct {
	VelocityIterations int
	PositionIterations int
	Restitution        matrix.Float
	StaticFriction     matrix.Float
	DynamicFriction    matrix.Float
	Baumgarte          matrix.Float
	PenetrationSlop    matrix.Float
	MaxCorrection      matrix.Float

	islands          []collisionIsland
	writableBodies   []*RigidBody
	parents          []int
	ranks            []uint8
	bodyIndex        map[*RigidBody]int
	rootToIsland     map[int]int
	eligibleContacts []int
	initialized      bool
}

func (s *CollisionSolver) Initialize() {
	s.VelocityIterations = defaultVelocityIterations
	s.PositionIterations = defaultPositionIterations
	s.Restitution = defaultRestitution
	s.StaticFriction = defaultStaticFriction
	s.DynamicFriction = defaultDynamicFriction
	s.Baumgarte = defaultBaumgarte
	s.PenetrationSlop = defaultPenetrationSlop
	s.MaxCorrection = defaultMaxCorrection
	s.bodyIndex = make(map[*RigidBody]int, 256)
	s.rootToIsland = make(map[int]int, 64)
	s.initialized = true
}

func (s *CollisionSolver) Reset() {
	for key := range s.bodyIndex {
		delete(s.bodyIndex, key)
	}
	for key := range s.rootToIsland {
		delete(s.rootToIsland, key)
	}
	for i := range s.islands {
		s.islands[i].manifolds = s.islands[i].manifolds[:0]
	}
	s.islands = s.islands[:0]
	s.writableBodies = s.writableBodies[:0]
	s.parents = s.parents[:0]
	s.ranks = s.ranks[:0]
	s.eligibleContacts = s.eligibleContacts[:0]
}

func (s *CollisionSolver) Solve(manifolds []ContactManifold, threads *concurrent.Threads) {
	if len(manifolds) == 0 {
		return
	}
	s.ensureInitialized()
	s.buildIslands(manifolds)
	if len(s.islands) == 0 {
		return
	}
	workers := broadPhaseWorkerCount(threads, len(s.islands), solverMinIslandsPerJob)
	if workers == 1 {
		s.solveIslandRange(manifolds, 0, len(s.islands))
		return
	}
	runSolverJobs(threads, workers, len(s.islands), func(start, end, _ int) {
		s.solveIslandRange(manifolds, start, end)
	})
}

func (s *CollisionSolver) ensureInitialized() {
	if !s.initialized {
		s.Initialize()
	}
}

func (s *CollisionSolver) buildIslands(manifolds []ContactManifold) {
	for key := range s.bodyIndex {
		delete(s.bodyIndex, key)
	}
	for key := range s.rootToIsland {
		delete(s.rootToIsland, key)
	}
	for i := range s.islands {
		s.islands[i].manifolds = s.islands[i].manifolds[:0]
	}
	s.islands = s.islands[:0]
	s.writableBodies = s.writableBodies[:0]
	s.parents = s.parents[:0]
	s.ranks = s.ranks[:0]
	s.eligibleContacts = s.eligibleContacts[:0]

	for i := range manifolds {
		manifold := &manifolds[i]
		if !s.shouldResolve(manifold) {
			continue
		}
		aIndex, aWritable := s.addWritableBody(manifold.BodyA)
		bIndex, bWritable := s.addWritableBody(manifold.BodyB)
		if aWritable && bWritable {
			s.union(aIndex, bIndex)
		}
		s.eligibleContacts = append(s.eligibleContacts, i)
	}

	for _, manifoldIndex := range s.eligibleContacts {
		manifold := &manifolds[manifoldIndex]
		root := s.manifoldRoot(manifold)
		if root < 0 {
			continue
		}
		islandIndex, ok := s.rootToIsland[root]
		if !ok {
			islandIndex = len(s.islands)
			s.rootToIsland[root] = islandIndex
			s.addIsland()
		}
		s.islands[islandIndex].manifolds = append(s.islands[islandIndex].manifolds, manifoldIndex)
	}
}

func (s *CollisionSolver) addIsland() {
	if len(s.islands) < cap(s.islands) {
		s.islands = s.islands[:len(s.islands)+1]
		s.islands[len(s.islands)-1].manifolds = s.islands[len(s.islands)-1].manifolds[:0]
		return
	}
	s.islands = append(s.islands, collisionIsland{})
}

func (s *CollisionSolver) shouldResolve(manifold *ContactManifold) bool {
	if manifold == nil || manifold.Count == 0 || manifold.BodyA == nil || manifold.BodyB == nil {
		return false
	}
	if manifold.BodyA.Collision.IsTrigger || manifold.BodyB.Collision.IsTrigger {
		return false
	}
	return manifold.BodyA.inverseMass()+manifold.BodyB.inverseMass() > 0
}

func (s *CollisionSolver) addWritableBody(body *RigidBody) (int, bool) {
	if body == nil || body.inverseMass() == 0 {
		return -1, false
	}
	if index, ok := s.bodyIndex[body]; ok {
		return index, true
	}
	index := len(s.writableBodies)
	s.bodyIndex[body] = index
	s.writableBodies = append(s.writableBodies, body)
	s.parents = append(s.parents, index)
	s.ranks = append(s.ranks, 0)
	return index, true
}

func (s *CollisionSolver) manifoldRoot(manifold *ContactManifold) int {
	if index, ok := s.bodyIndex[manifold.BodyA]; ok {
		return s.find(index)
	}
	if index, ok := s.bodyIndex[manifold.BodyB]; ok {
		return s.find(index)
	}
	return -1
}

func (s *CollisionSolver) find(index int) int {
	parent := s.parents[index]
	if parent != index {
		parent = s.find(parent)
		s.parents[index] = parent
	}
	return parent
}

func (s *CollisionSolver) union(a, b int) {
	rootA := s.find(a)
	rootB := s.find(b)
	if rootA == rootB {
		return
	}
	if s.ranks[rootA] < s.ranks[rootB] {
		rootA, rootB = rootB, rootA
	}
	s.parents[rootB] = rootA
	if s.ranks[rootA] == s.ranks[rootB] {
		s.ranks[rootA]++
	}
}

func (s *CollisionSolver) solveIslandRange(manifolds []ContactManifold, start, end int) {
	for islandIndex := start; islandIndex < end; islandIndex++ {
		island := &s.islands[islandIndex]
		for range s.VelocityIterations {
			for _, manifoldIndex := range island.manifolds {
				s.solveVelocity(&manifolds[manifoldIndex])
			}
		}
		for range s.PositionIterations {
			for _, manifoldIndex := range island.manifolds {
				s.solvePosition(&manifolds[manifoldIndex])
			}
		}
	}
}

func (s *CollisionSolver) solveVelocity(manifold *ContactManifold) {
	bodyA := manifold.BodyA
	bodyB := manifold.BodyB
	normal := safeNormal(manifold.Normal, matrix.Vec3Right())
	for i := range manifold.Count {
		contact := manifold.Contacts[i]
		ra := contact.Point.Subtract(bodyA.Transform.WorldPosition())
		rb := contact.Point.Subtract(bodyB.Transform.WorldPosition())
		relativeVelocity := velocityAtContact(bodyB, rb).Subtract(velocityAtContact(bodyA, ra))
		normalVelocity := relativeVelocity.Dot(normal)
		if normalVelocity > 0 {
			continue
		}
		denominator := impulseDenominator(bodyA, bodyB, ra, rb, normal)
		if denominator <= contactEpsilon {
			continue
		}
		normalImpulseMagnitude := -(1 + s.Restitution) * normalVelocity / denominator
		normalImpulseMagnitude /= matrix.Float(manifold.Count)
		normalImpulse := normal.Scale(normalImpulseMagnitude)
		applyImpulse(bodyA, normalImpulse.Negative(), ra)
		applyImpulse(bodyB, normalImpulse, rb)

		relativeVelocity = velocityAtContact(bodyB, rb).Subtract(velocityAtContact(bodyA, ra))
		tangent := relativeVelocity.Subtract(normal.Scale(relativeVelocity.Dot(normal)))
		if tangent.LengthSquared() <= contactEpsilon*contactEpsilon {
			continue
		}
		tangent = tangent.Normal()
		tangentDenominator := impulseDenominator(bodyA, bodyB, ra, rb, tangent)
		if tangentDenominator <= contactEpsilon {
			continue
		}
		tangentImpulseMagnitude := -relativeVelocity.Dot(tangent) / tangentDenominator
		tangentImpulseMagnitude /= matrix.Float(manifold.Count)
		maxStaticFriction := normalImpulseMagnitude * s.StaticFriction
		var tangentImpulse matrix.Vec3
		if matrix.Abs(tangentImpulseMagnitude) <= maxStaticFriction {
			tangentImpulse = tangent.Scale(tangentImpulseMagnitude)
		} else {
			dynamicMagnitude := klib.Clamp(tangentImpulseMagnitude,
				-normalImpulseMagnitude*s.DynamicFriction,
				normalImpulseMagnitude*s.DynamicFriction)
			tangentImpulse = tangent.Scale(dynamicMagnitude)
		}
		applyImpulse(bodyA, tangentImpulse.Negative(), ra)
		applyImpulse(bodyB, tangentImpulse, rb)
	}
}

func (s *CollisionSolver) solvePosition(manifold *ContactManifold) {
	bodyA := manifold.BodyA
	bodyB := manifold.BodyB
	current, ok := CollideBodies(bodyA, bodyB)
	if !ok {
		return
	}
	manifold = &current
	invMassA := bodyA.inverseMass()
	invMassB := bodyB.inverseMass()
	invMassSum := invMassA + invMassB
	if invMassSum <= contactEpsilon {
		return
	}
	normal := safeNormal(manifold.Normal, matrix.Vec3Right())
	for i := range manifold.Count {
		penetration := manifold.Contacts[i].Penetration
		depth := matrix.Max(penetration-s.PenetrationSlop, 0)
		if depth <= 0 {
			continue
		}
		correctionMagnitude := depth * s.Baumgarte / invMassSum
		correctionMagnitude = matrix.Min(correctionMagnitude, s.MaxCorrection)
		correctionMagnitude /= matrix.Float(manifold.Count)
		correction := normal.Scale(correctionMagnitude)
		moveBody(bodyA, correction.Scale(-invMassA))
		moveBody(bodyB, correction.Scale(invMassB))
	}
}

func velocityAtContact(body *RigidBody, r matrix.Vec3) matrix.Vec3 {
	if body == nil {
		return matrix.Vec3Zero()
	}
	return body.MotionState.LinearVelocity.Add(body.MotionState.AngularVelocity.Cross(r))
}

func impulseDenominator(bodyA, bodyB *RigidBody, ra, rb, axis matrix.Vec3) matrix.Float {
	denominator := bodyA.inverseMass() + bodyB.inverseMass()
	denominator += angularImpulseDenominator(bodyA, ra, axis)
	denominator += angularImpulseDenominator(bodyB, rb, axis)
	return denominator
}

func angularImpulseDenominator(body *RigidBody, r, axis matrix.Vec3) matrix.Float {
	inverseInertia := body.inverseInertia()
	if inverseInertia.IsZero() {
		return 0
	}
	angular := r.Cross(axis).Multiply(inverseInertia)
	return angular.Cross(r).Dot(axis)
}

func applyImpulse(body *RigidBody, impulse, r matrix.Vec3) {
	invMass := body.inverseMass()
	if invMass == 0 {
		return
	}
	body.MotionState.LinearVelocity.AddAssign(impulse.Scale(invMass))
	invInertia := body.inverseInertia()
	if invInertia.IsZero() {
		return
	}
	angularImpulse := r.Cross(impulse).Multiply(invInertia)
	body.MotionState.AngularVelocity.AddAssign(angularImpulse)
}

func moveBody(body *RigidBody, correction matrix.Vec3) {
	if body == nil || body.inverseMass() == 0 || correction.LengthSquared() <= contactEpsilon*contactEpsilon {
		return
	}
	body.Transform.AddPosition(correction)
}

func runSolverJobs(threads *concurrent.Threads, workers, items int, work func(start, end, worker int)) {
	jobs := make([]func(int), workers)
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for worker := range workers {
		start := worker * items / workers
		end := (worker + 1) * items / workers
		worker := worker
		jobs[worker] = func(int) {
			defer wg.Done()
			work(start, end, worker)
		}
	}
	threads.AddWork(jobs)
	wg.Wait()
}
