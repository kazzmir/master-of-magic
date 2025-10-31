package pathfinding

/* various path finding algorithms */

import (
    "image"
    "slices"
    "math"
    _ "log"
    "cmp"

    "github.com/kazzmir/master-of-magic/lib/priority"
)

type Path []image.Point

// cost to move from (x1,y1) -> (x2,y2)
type TileCostFunc func(int, int, int, int) float64
// neighbor points given a point x,y
type NeighborsFunc func (int, int) []image.Point
// compare two points for equality. notably this should handle wrapping around the map
type PointEqFunc func (image.Point, image.Point) bool

var Infinity = math.Inf(1)

// default point equality function
func PointEqual(a image.Point, b image.Point) bool {
    return a == b
}

/* returns an array of points that is the shortest/cheapest path from start->end and true, or false if no such path exists
 * basically djikstra's shortest path algorithm
 */
func FindPath(start image.Point, end image.Point, maxPath float64, tileCost TileCostFunc, neighbors NeighborsFunc, samePoint PointEqFunc) (Path, bool) {

    // set distance to start as 0
    // put start in open list
    // find all neighbors, choose a neighbor with the lowest distance to start
    // update neighbor cost with current cost + cost to neighbor, or keep current neighbor cost if lower

    type Node struct {
        point image.Point
        previous *Node
        visited bool
        cost float64
        // keep track of when nodes were added to enforce ordering for equal cost nodes
        time uint64
    }

    var endNode *Node

    nodes := make(map[image.Point]*Node)

    nodes[start] = &Node{
        point: start,
        cost: 0,
        visited: false,
        time: 0,
    }

    lowestCost := Infinity

    compare := func (a *Node, b *Node) int {
        if a.cost < b.cost {
            return -1
        }

        if a.cost > b.cost {
            return 1
        }

        return cmp.Compare(a.time, b.time)
    }

    unvisited := priority.MakePriorityQueue[*Node](compare)
    unvisited.Insert(nodes[start])

    var globalTime uint64 = 0
    for !unvisited.IsEmpty() {
        node := unvisited.ExtractMin()

        if node.visited {
            continue
        }

        nodes[node.point].visited = true

        // ignore paths that are already more expensive than the lowest cost path
        if node.cost > lowestCost || node.cost > maxPath {
            continue
        }

        if samePoint(node.point, end) && node.cost < Infinity {
            endNode = node
            if node.cost < lowestCost {
                lowestCost = node.cost
            }
            continue
        }

        for _, neighbor := range neighbors(node.point.X, node.point.Y) {
            globalTime += 1
            newNode, ok := nodes[neighbor]
            if !ok {
                newNode = &Node{point: neighbor, cost: Infinity, time: globalTime}
                nodes[neighbor] = newNode
            }

            newCost := node.cost + tileCost(node.point.X, node.point.Y, newNode.point.X, newNode.point.Y)
            if newCost < newNode.cost {
                newNode.cost = newCost
                newNode.previous = node
                unvisited.Insert(newNode)
            }
        }
    }

    if endNode != nil {
        var out []image.Point
        seen := make(map[image.Point]struct{})
        for endNode != nil && !samePoint(endNode.point, start) {
            /* make sure we never see the same node twice, otherwise we entered into a loop somehow */
            _, ok := seen[endNode.point]
            if ok {
                return nil, false
            }

            seen[endNode.point] = struct{}{}

            out = append(out, endNode.point)
            endNode = endNode.previous
        }

        if endNode == nil || !samePoint(endNode.point, start) {
            return nil, false
        }

        out = append(out, start)

        slices.Reverse(out)

        return out, true
    }

    return nil, false
}
