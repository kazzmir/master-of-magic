/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

case class CardinalDirection(val id:Int, val dx:Int, val dy:Int)
object CardinalDirection{

  val NORTH       = CardinalDirection(0, 0, -1);
  val NORTH_EAST  = CardinalDirection(1, 1, -1);
  val EAST        = CardinalDirection(2, 1, 0);
  val SOUTH_EAST  = CardinalDirection(3, 1, 1);
  val SOUTH       = CardinalDirection(4, 0, 1);
  val SOUTH_WEST  = CardinalDirection(5, -1, 1);
  val WEST        = CardinalDirection(6, -1, 0);
  val NORTH_WEST  = CardinalDirection(7, -1, -1);
  val CENTER      = CardinalDirection(8, 0, 0);

  val values = Array(NORTH, NORTH_EAST, EAST, SOUTH_EAST, SOUTH, SOUTH_WEST, WEST, NORTH_WEST);
  val valuesStraight = Array(NORTH, EAST, SOUTH, WEST);
  val valuesDiagonal = Array(NORTH_EAST, SOUTH_EAST, SOUTH_WEST, NORTH_WEST);
  val valuesAll = Array(CENTER, NORTH, NORTH_EAST, EAST, SOUTH_EAST, SOUTH, SOUTH_WEST, WEST, NORTH_WEST);
}

case class Plane(val id:Int, val name:String)
object Plane {
  val ARCANUS = Plane(0, "Arcanus")
  val MYRROR = Plane(1, "Myrror")

  val values = Array(ARCANUS, MYRROR)

  implicit def plane2string(p:Plane) = p.name
    
}

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

case class TerrainType(val id:Int, val name:String)
object TerrainType {
  val OCEAN = TerrainType(0, "Ocean");
  val SHORE = TerrainType(1, "Shore");
  val RIVER = TerrainType(2, "River");
  val SWAMP = TerrainType(3, "Swamp");
  val TUNDRA = TerrainType(4, "Tundra");
  val DEEP_TUNDRA = TerrainType(5, "Deep Tundra");
  val MOUNTAIN = TerrainType(6, "Mountain");
  val VOLCANO = TerrainType(7, "Volcano");
  val CHAOS_NODE = TerrainType(8, "Chaos Node");
  val HILLS = TerrainType(9, "Hills");
  val GRASSLAND = TerrainType(10, "Grassland");
  val SORCERY_NODE = TerrainType(11, "Sorcery Node");
  val DESERT = TerrainType(12, "Desert");
  val FOREST = TerrainType(13, "Forest");
  val NATURE_NODE = TerrainType(14, "Nature Node");

  val values = Array(
    OCEAN,
    SHORE,
    RIVER,
    SWAMP,
    TUNDRA,
    DEEP_TUNDRA,
    MOUNTAIN,
    VOLCANO,
    CHAOS_NODE,
    HILLS,
    GRASSLAND,
    SORCERY_NODE,
    DESERT,
    FOREST,
    NATURE_NODE);

  implicit def terrainType2string(t:TerrainType) = t.name
}

case class TerrainTileMetadata(
  val id:Int,
  val terrainType:TerrainType,
  val borderingTerrainTypes:Array[Option[TerrainType]],
  val plane:Plane,
  val parentId:Option[TerrainTileMetadata])

object TerrainTileMetadata {
  import com.rafkind.masterofmagic.util.TerrainLbxReader._;
  
  var data:Array[TerrainTileMetadata] = 
    new Array[TerrainTileMetadata](TILE_COUNT);

  def read(fn:String):Unit = {
    
  }
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
