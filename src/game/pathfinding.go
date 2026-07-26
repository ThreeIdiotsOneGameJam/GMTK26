package game

import "container/heap"

// PathCostFunc returns the cost of entering destination from source.
// A cost less than or equal to zero makes that step impassable.
//
// The callback must not mutate destination. It is shared by client previews
// and authoritative server pathfinding, so it should be deterministic.
type PathCostFunc func(source, destination Hex, cell *Cell) int

// PathOptions controls how FindPathWithOptions traverses the map.
type PathOptions struct {
	// Cost controls passability and movement cost. Cells outside the grid and
	// void cells are always impassable. When Cost is nil, land cells cost one
	// and water cells are also impassable.
	Cost PathCostFunc

	// MinimumCost is the lowest positive value Cost can return. It strengthens
	// the A* heuristic for weighted maps. Values below one default to one.
	// Setting this higher than the actual minimum can produce a non-optimal path.
	MinimumCost int
}

// FindPath finds the cheapest land path from start to goal using A*.
// The returned path includes both start and goal. Water and void cells are
// impassable; use FindPathWithOptions to customize traversal.
func (m *Map) FindPath(start, goal Hex) ([]Hex, bool) {
	return m.FindPathWithOptions(start, goal, PathOptions{})
}

// FindPathWithOptions finds the cheapest path from start to goal using A*.
// The returned path includes both endpoints. It returns false when either
// endpoint is outside the map or no path exists.
func (m *Map) FindPathWithOptions(start, goal Hex, options PathOptions) ([]Hex, bool) {
	if !m.pathEndpointValid(start) || !m.pathEndpointValid(goal) {
		return nil, false
	}

	cost := options.Cost
	if cost == nil {
		cost = landPathCost
		if m.GetCell(start).Tile == TileWater || m.GetCell(goal).Tile == TileWater {
			return nil, false
		}
	}

	if start == goal {
		return []Hex{start}, true
	}

	minimumCost := options.MinimumCost
	if minimumCost < 1 {
		minimumCost = 1
	}

	frontier := &pathPriorityQueue{}
	heap.Init(frontier)
	sequence := 0
	heap.Push(frontier, pathQueueEntry{
		hex:      start,
		cost:     0,
		priority: int64(start.Distance(goal)) * int64(minimumCost),
		sequence: sequence,
	})

	cameFrom := make(map[Hex]Hex)
	costSoFar := map[Hex]int64{start: 0}

	for frontier.Len() > 0 {
		current := heap.Pop(frontier).(pathQueueEntry)
		bestCost, exists := costSoFar[current.hex]
		if !exists || current.cost != bestCost {
			continue
		}
		if current.hex == goal {
			return reconstructPath(cameFrom, start, goal), true
		}

		for _, next := range pathNeighborHexes(current.hex) {
			cell := m.GetCell(next)
			if cell == nil || cell.Tile == TileVoid {
				continue
			}

			stepCost := cost(current.hex, next, cell)
			if stepCost <= 0 {
				continue
			}

			nextCost := current.cost + int64(stepCost)
			previousCost, visited := costSoFar[next]
			if visited && nextCost >= previousCost {
				continue
			}

			costSoFar[next] = nextCost
			cameFrom[next] = current.hex
			sequence++
			heap.Push(frontier, pathQueueEntry{
				hex:      next,
				cost:     nextCost,
				priority: nextCost + int64(next.Distance(goal))*int64(minimumCost),
				sequence: sequence,
			})
		}
	}

	return nil, false
}

func (m *Map) pathEndpointValid(hex Hex) bool {
	if m == nil {
		return false
	}
	cell := m.GetCell(hex)
	return cell != nil && cell.Tile != TileVoid
}

func landPathCost(_, _ Hex, cell *Cell) int {
	if cell.Tile == TileWater || cell.Tile == TileVoid {
		return 0
	}
	return 1
}

func pathNeighborHexes(pos Hex) [6]Hex {
	if pos.X&1 == 0 {
		return [6]Hex{
			pos.Add(NewHex(-1, -1)),
			pos.Add(NewHex(0, -1)),
			pos.Add(NewHex(1, -1)),
			pos.Add(NewHex(1, 0)),
			pos.Add(NewHex(0, 1)),
			pos.Add(NewHex(-1, 0)),
		}
	}

	return [6]Hex{
		pos.Add(NewHex(-1, 0)),
		pos.Add(NewHex(0, -1)),
		pos.Add(NewHex(1, 0)),
		pos.Add(NewHex(1, 1)),
		pos.Add(NewHex(0, 1)),
		pos.Add(NewHex(-1, 1)),
	}
}

func reconstructPath(cameFrom map[Hex]Hex, start, goal Hex) []Hex {
	path := []Hex{goal}
	for current := goal; current != start; {
		current = cameFrom[current]
		path = append(path, current)
	}

	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}

type pathQueueEntry struct {
	hex      Hex
	cost     int64
	priority int64
	sequence int
}

type pathPriorityQueue []pathQueueEntry

func (queue pathPriorityQueue) Len() int {
	return len(queue)
}

func (queue pathPriorityQueue) Less(i, j int) bool {
	if queue[i].priority == queue[j].priority {
		return queue[i].sequence < queue[j].sequence
	}
	return queue[i].priority < queue[j].priority
}

func (queue pathPriorityQueue) Swap(i, j int) {
	queue[i], queue[j] = queue[j], queue[i]
}

func (queue *pathPriorityQueue) Push(value any) {
	*queue = append(*queue, value.(pathQueueEntry))
}

func (queue *pathPriorityQueue) Pop() any {
	old := *queue
	last := len(old) - 1
	value := old[last]
	*queue = old[:last]
	return value
}
