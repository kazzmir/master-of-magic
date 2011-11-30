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

object TerrainSquare {
  val EMPTY:TerrainSquare = new TerrainSquare(0);
}

class TerrainSquare(val terrain:Int) {
  // what type of terrain
  // what terrain tile to use
  // bitset for fog of war
  // polluted?
  // what type of bonus
  // road?
  // what city is here?
  // what unit stack is here?
}


object Overworld {
  val WIDTH = 100;
  val HEIGHT = 50;

  def createExampleWorld:Overworld = {
    var overworld = new Overworld(WIDTH, HEIGHT);

    for (y <- 0 until HEIGHT) {
      for (x <- 0 until WIDTH) {
        var distx = (WIDTH/2) - x;
        var disty = (HEIGHT/2) - y;
        var dist = distx*distx + disty*disty;
        overworld.put(x, y, new TerrainSquare(dist / 350));
      }
    }

    return overworld;
  }
}

class Overworld(width:Int, height:Int) {
  var terrain:Array[TerrainSquare] = 
    new Array[TerrainSquare](width * height);

  def get(x:Int, y:Int):TerrainSquare = {
    var xx = x % Overworld.WIDTH;

    if (y >= 0 && y <= Overworld.HEIGHT) {
      return terrain(y * Overworld.WIDTH + x);
    } else {
      return TerrainSquare.EMPTY;
    }
  }

  def put(x:Int, y:Int, terrainSquare:TerrainSquare):Unit = {
    terrain(y * Overworld.WIDTH + x) = terrainSquare;
  }
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
