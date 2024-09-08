package combat

/* various path finding algorithms */

import (
    "image"
)

const Infinity = 1000000

/* returns an array of points that is the shortest/cheapest path from start->end and true, or false if no such path exists
 * basically djikstra's shortest path algorithm
 */
func FindPath(start image.Point, end image.Point, maxPath int, tileCost func(int, int) int, neighbors (int, int) []image.Point) ([]image.Point, bool) {

    // set distance to start as 0
    // put start in open list
    // find all neighbors, choose a neighbor with the lowest distance to start
    // update neighbor cost with current cost + cost to neighbor, or keep current neighbor cost if lower

    type Node struct {
        point image.Point
        previous *Node
        visited bool
        cost int
    }

    var endNode *Node

    nodes := make(map[image.Point]*Node)

    nodes[start] = &Node{
        point: start,
        cost: 0,
        visited: false,
    }

    lowestCost := Infinity

    // this should be a priority queue
    var openList []*Node
    openList = append(openList, nodes[start])

    for len(openList) > 0 {
        node := openList[0]
        openList = openList[1:]

        if node.visited {
            continue
        }

        nodes[node.point].visited = true

        // ignore paths that are already more expensive than the lowest cost path
        if node.cost > lowestCost {
            continue
        }

        if node.point.Eq(end) {
            endNode = node
            if node.cost < lowestCost {
                lowestCost = node.cost
            }
        }

        for _, neighbor := range neighbors(node.point.X, node.point.Y) {
            newNode, ok := nodes[neighbor]
            if !ok {
                newNode = &Node{point: neighbor, cost: Infinity}
                nodes[neighbor] = newNode
            }

            newCost := node.cost + tileCost(newNode.point.X, newNode.point.Y)
            if newCost < newNode.cost {
                newNode.cost = newCost
                newNode.previous = node
                openList = append(openList, newNode)
            }
        }
    }

    if endNode != nil {
        // max path to ensure we don't follow an infinite loop somehow
        var out []image.Point
        for len(out) < maxPath && endNode != nil && !endNode.point.Equal(start) {
            out = append(out, endNode.point)
            endNode = endNode.previous
        }

        if endNode == nil || !endNode.point.Eq(start) {
            return nil, false
        }

        out = append(out, start)

        slices.Reverse(out)
        return out, true
    }

    return nil, false

}
