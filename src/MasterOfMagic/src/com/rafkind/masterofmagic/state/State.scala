/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

class Player {
  // id
  // name
  // picture
  // color scheme
  // music scheme
}

class TerrainType {
  // passability cost
  // base food
  // base money
  // base mana generation
}

class TerrainSquare {
  // what type of terrain
  // what terrain tile to use
  // bitset for fog of war
  // polluted?
  // what type of bonus
  // road?
  // what city is here?
  // what unit stack is here?
}

class Overworld {
  // terrain array
}

class State {
  // players
  // normal world, mirror world
  // cities
  // lairs
  // magic nodes
  // unit stacks
  // units
}
