/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

object Overworld {
  val WIDTH = 60;
  val HEIGHT = 40;

  /*
   * Arcanus:
   * OCEAN = 0
   * SHORE = XXX
   * GRASSLAND = 1
   * FOREST = 274
   * MOUNTAIN = 275
   * DESERT = 276
   * SWAMP = 277
   * TUNDRA = 278
   * DEEP_TUNDRA = XXX
   * SORCERY NODE = 279
   * NATURE NODE = 283
   * CHAOS NODE = 287
   * HILLS = 291
   * VOLCANO = 299
   * RIVER = 591
   *
   * Myrror:
   * OCEAN = 888
   * GRASSLAND = 889
   * FOREST = 1147
   * MOUNTAIN = 1148
   * DESERT = 1149
   * SWAMP = 1150
   * SORCERY NODE = 1152
   * FOREST NODE = 1186
   * CHAOS NODE = 1160
   * HILLS = 1164
   * VOLCANO = 1172
   * RIVER = 1464
   * TUNDRA = 1605
   */

  def create():Overworld = {

    val arcanus = Map(TerrainType.OCEAN -> 0,
      TerrainType.GRASSLAND -> 1,
      TerrainType.FOREST -> 274,
      TerrainType.MOUNTAIN -> 275,
      TerrainType.DESERT -> 276,
      TerrainType.SWAMP -> 277,
      TerrainType.TUNDRA -> 278,
      TerrainType.SORCERY_NODE -> 279,
      TerrainType.NATURE_NODE -> 283,
      TerrainType.CHAOS_NODE -> 287,
      TerrainType.HILLS -> 291,
      TerrainType.VOLCANO -> 299,
      TerrainType.RIVER -> 591);
    val myrror = Map(TerrainType.OCEAN -> 888);

    var overworld = new Overworld(WIDTH, HEIGHT);

    overworld.fillPlane(Plane.ARCANUS, arcanus(TerrainType.OCEAN), TerrainType.OCEAN);
    overworld.fillPlane(Plane.MYRROR, myrror(TerrainType.OCEAN), TerrainType.OCEAN);

    overworld.buildPlane(Plane.ARCANUS, 5, 800);

    for (y <- 0 until HEIGHT) {
      for (x <- 0 until WIDTH) {
        var t = overworld.get(Plane.ARCANUS, x, y);
        if (y > 0 && y < HEIGHT-1) {
          t.spriteNumber =
            TerrainTileMetadata.recommendedTerrainSprite(Plane.ARCANUS,
              CardinalDirection.valuesAll map {(d) => overworld.get(Plane.ARCANUS, x + d.dx + WIDTH, y + d.dy).terrainType});
        } else {
          t.spriteNumber = arcanus(t.terrainType);
        }
      }
    }


    /*overworld.buildPlane(Plane.MYRROR, 3, 10);
    for (y <- 0 until HEIGHT) {
      for (x <- 0 until WIDTH) {
        var t = overworld.get(Plane.MYRROR, x, y);
        t.spriteNumber = myrror(t.terrainType);
      }
    }*/

    return overworld;
  }
}

class Overworld(val width:Int, val height:Int) {
  import scala.util.Random;

  var terrain:Array[TerrainSquare] =
    new Array[TerrainSquare](2 * width * height);

  def get(plane:Plane, x:Int, y:Int):TerrainSquare = {
    val xx = x % width;

    if (y >= 0 && y <= height) {
      return terrain(plane.id * width * height
                     + y * width
                     + xx);
    } else {
      throw new IllegalArgumentException("Bad coordinates");
    }
  }

  def put(plane:Plane, x:Int, y:Int, terrainSquare:TerrainSquare):Unit = {
    val xx = x % width;
    terrain(plane.id * width * height
            + y * width
            + xx) = terrainSquare;
  }

  def fillPlane(plane:Plane, tileId:Int, terrain:TerrainType):Unit = {
    for (y <- 0 until height) {
      for (x <- 0 until width) {
        put(plane, x, y,
                      new TerrainSquare(
                        tileId,
                        terrain,
                        0,
                        false,
                        0,
                        None,
                        None));
      }
    }
  }

  def findWhereTrue(random:Random,
                    plane:Plane,
                    predicate:(Int,Int) => Boolean):Tuple2[Int,Int] = {

    var x:Int = 0;
    var y:Int = 0;
    do {
      x = random.nextInt(width);
      y = random.nextInt(height);
    } while (!predicate(x,y));

    return (x,y);
  }

  def grow(random:Random,
           plane:Plane,
           numSeeds:Int,
           groundCount:Int,
           base:TerrainType,
           topping:TerrainType):Unit = {
    for(n <- 0 until numSeeds) {
      findWhereTrue(random,
                     plane,
                     (x,y) => {
                        get(plane, x, y).terrainType == base
                      }) match {
        case (x,y) => {
          var t = get(plane, x, y);
          t.terrainType = topping;
        }
      }
    }

    for (n <- 0 until groundCount) {
      findWhereTrue(random,
                     plane,
                     (x,y) => {
                        var found = false;
                        if (get(plane, x, y).terrainType == base) {
                          for (dir <- CardinalDirection.valuesStraight) {
                            val nx = x + dir.dx + width;
                            val ny = y + dir.dy;
                            if (ny >=0 && ny < height) {
                              if (get(plane, nx, ny).terrainType == topping) {
                                found = true;
                              }
                            }
                          }
                        }
                        found;
                      }) match {
        case (x,y) => {
          var t = get(plane, x, y);
          t.terrainType = topping;
        }
      }
    }

    // fill in the gaps
    for (y <- 0 to height) {
      for (x <- 0 to width) {
        if (get(plane, x, y).terrainType == base) {
          var count = 0;
          for (dir <- CardinalDirection.valuesStraight) {
            val nx = x + dir.dx + width;
            val ny = y + dir.dy;
            if (ny >=0 && ny < height) {
              if (get(plane, nx, ny).terrainType == topping) {
                count += 1;
              }
            }
          }
          if (count == 4) {
            var t = get(plane, x, y);
            t.terrainType = topping;
          }
        }
      }
    }
  }

  def growRiver(random:Random, plane:Plane) {
    findWhereTrue(random, plane, (x,y) => {
        var t = get(plane, x, y);
        (t.terrainType != TerrainType.OCEAN)
      }) match {
      case (x,y) => {
        findWhereTrue(random, plane, (x2, y2) => {
          get(plane, x2, y2).terrainType == TerrainType.OCEAN;
        }) match {
          case (ox, oy) => {
            var rx = x;
            var ry = y;
            while (get(plane, rx, ry).terrainType != TerrainType.OCEAN) {
              get(plane, rx, ry).terrainType = TerrainType.RIVER;

              if (random.nextInt(2) == 1) {
                if (ry < oy) {
                  ry += 1;
                } else if (ry > oy) {
                  ry -= 1;
                }

              } else {
                if ((rx-ox+width) % width > (ox-rx+width) % width) {
                  rx += 1;
                }
                if ((rx-ox+width) % width < (ox-rx+width) % width) {
                  rx = (rx - 1 + width) % width;
                }
              }
            }
          }
        }
      }
    }
  }

  def buildPlane(plane:Plane, numSeeds:Int, groundCount:Int):Unit = {
    var random = new Random();

    println("Land");
    grow(random, plane, numSeeds, groundCount, TerrainType.OCEAN, TerrainType.GRASSLAND);
    println("Forest");
    grow(random, plane, numSeeds * 3, groundCount / 13, TerrainType.GRASSLAND, TerrainType.FOREST);
    println("Hills");
    grow(random, plane, numSeeds * 3, groundCount / 18, TerrainType.GRASSLAND, TerrainType.HILLS);
    println("Mountain");
    grow(random, plane, numSeeds * 4, groundCount / 24, TerrainType.GRASSLAND, TerrainType.MOUNTAIN);
    println("Desert");
    grow(random, plane, numSeeds * 3, groundCount / 15, TerrainType.GRASSLAND, TerrainType.DESERT);
    println("Swamp");
    grow(random, plane, numSeeds * 3, groundCount / 15, TerrainType.GRASSLAND, TerrainType.SWAMP);

    println("River");
    for (n <- 0 until numSeeds * 2 / 3) {
      growRiver(random, plane);
    }

    println("Nodes");
    grow(random, plane, numSeeds, 0, TerrainType.GRASSLAND, TerrainType.SORCERY_NODE);
    grow(random, plane, numSeeds, 0, TerrainType.MOUNTAIN, TerrainType.CHAOS_NODE);
    grow(random, plane, numSeeds, 0, TerrainType.FOREST, TerrainType.NATURE_NODE);


    println("Tundra");
    // north and south poles, and tundra
    for (x <- 0 until width) {
      get(plane, x, 0).terrainType = TerrainType.TUNDRA;
      get(plane, x, height-1).terrainType = TerrainType.TUNDRA;
      for (y <- 0 until height / 8) {
        if (get(plane, x, y).terrainType == TerrainType.GRASSLAND ||
          get(plane, x, y).terrainType == TerrainType.HILLS ||
          get(plane, x, y).terrainType == TerrainType.DESERT) {
          get(plane, x, y).terrainType = TerrainType.TUNDRA;
        }
        val yy = height - (y+1);
        if (get(plane, x, yy).terrainType == TerrainType.GRASSLAND ||
          get(plane, x, yy).terrainType == TerrainType.HILLS ||
          get(plane, x, yy).terrainType == TerrainType.DESERT) {
          get(plane, x, yy).terrainType = TerrainType.TUNDRA;
        }
      }
    }


  }

}
