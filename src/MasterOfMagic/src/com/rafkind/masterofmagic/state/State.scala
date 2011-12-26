/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

class Place {

}

class UnitStack {

}

class Player {
  // id
  // name
  // picture
  // color scheme
  // music scheme
}

// http://www.dragonsword.com/magic/eljay/SaveGam.html
class TerrainType(val id:Int) {
}

object TerrainType {
  val OCEAN = new TerrainType(0);
  val SHORE = new TerrainType(1);
  val RIVER = new TerrainType(2);
  val SWAMP = new TerrainType(3);
  val TUNDRA = new TerrainType(4);
  val DEEP_TUNDRA = new TerrainType(5);
  val MOUNTAIN = new TerrainType(6);
  val VOLCANO = new TerrainType(7);
  val CHAOS_NODE = new TerrainType(8);
  val HILLS = new TerrainType(9);
  val GRASSLAND = new TerrainType(10);
  val SORCERY_NODE = new TerrainType(11);
  val DESERT = new TerrainType(12);
  val FOREST = new TerrainType(13);
  val NATURE_NODE = new TerrainType(14);
}



object TerrainSquare {
  
}

class TerrainSquare(
  var spriteNumber:Int,
  var terrainType:TerrainType,
  var fogOfWarBitset:Int,
  var pollutionFlag:Boolean,
  var roadBitset:Int,
  var building:Option[Place],
  var unitStack:Option[UnitStack]) {
    
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
  val WIDTH = 60;
  val HEIGHT = 40;

  def createExampleWorld:Overworld = {
    var overworld = new Overworld(WIDTH, HEIGHT);

    for (y <- 0 until HEIGHT) {
      for (x <- 0 until WIDTH) {
        var distx = (WIDTH/2) - x;
        var disty = (HEIGHT/2) - y;
        var dist = distx*distx + disty*disty;
        overworld.put(x, y, 
                      new TerrainSquare(/* dist / 1000 */ x % 9,
                        TerrainType.OCEAN,
                        0,
                        false,
                        0,
                        None,
                        None));
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
      throw new IllegalArgumentException("Bad coordinates");
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
