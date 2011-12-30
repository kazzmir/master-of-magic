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

// Enumerations in scala are confusing!
// the following copied from here:
// http://downgra.de/2010/02/11/playing-with-scala-enumeration/
object TerrainType extends Enumeration {

  case class TerrainTypeVal(number:Int) extends Val(number) {
    // put definitions in here
  }

  val OCEAN = TerrainTypeVal(0);
  val SHORE = TerrainTypeVal(1);
  val RIVER = TerrainTypeVal(2);
  val SWAMP = TerrainTypeVal(3);
  val TUNDRA = TerrainTypeVal(4);
  val DEEP_TUNDRA = TerrainTypeVal(5);
  val MOUNTAIN = TerrainTypeVal(6);
  val VOLCANO = TerrainTypeVal(7);
  val CHAOS_NODE = TerrainTypeVal(8);
  val HILLS = TerrainTypeVal(9);
  val GRASSLAND = TerrainTypeVal(10);
  val SORCERY_NODE = TerrainTypeVal(11);
  val DESERT = TerrainTypeVal(12);
  val FOREST = TerrainTypeVal(13);
  val NATURE_NODE = TerrainTypeVal(14);

  // needed I think because Enumeration.elements is final and return the invariant
  // type Enumeration.Value :|
  implicit def valueToPlanet(v: Value): TerrainTypeVal =
    v.asInstanceOf[TerrainTypeVal]
}



object TerrainSquare {
  
}

class TerrainSquare(
  var spriteNumber:Int /*,
  var terrainType:TerrainType.TerrainTypeVal,
  var fogOfWarBitset:Int,
  var pollutionFlag:Boolean,
  var roadBitset:Int,
  var building:Option[Place],
  var unitStack:Option[UnitStack]*/ ) {
    
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
                      new TerrainSquare(/* dist / 1000 */ x % 9 /*,
                        TerrainType.OCEAN,
                        0,
                        false,
                        0,
                        None,
                        None */));
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
