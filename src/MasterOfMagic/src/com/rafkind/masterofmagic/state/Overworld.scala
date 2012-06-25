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

  def create(neutralPlayer:Player):Overworld = {

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
    val myrror = Map(TerrainType.OCEAN -> 888,
      TerrainType.GRASSLAND -> 889,
      TerrainType.FOREST -> 1147,
      TerrainType.MOUNTAIN -> 1148,
      TerrainType.DESERT -> 1149,
      TerrainType.SWAMP -> 1150,
      TerrainType.SORCERY_NODE -> 1152,
      TerrainType.NATURE_NODE -> 1186,
      TerrainType.CHAOS_NODE -> 1160,
      TerrainType.HILLS -> 1164,
      TerrainType.VOLCANO -> 1172,
      TerrainType.RIVER -> 1464,
      TerrainType.TUNDRA -> 1605);

    var overworld = new Overworld(WIDTH, HEIGHT);

    overworld.fillPlane(Plane.ARCANUS, arcanus(TerrainType.OCEAN), TerrainType.OCEAN);
    overworld.fillPlane(Plane.MYRROR, myrror(TerrainType.OCEAN), TerrainType.OCEAN);

    overworld.buildPlane(Plane.ARCANUS, 5, 800, neutralPlayer);
    overworld.buildPlane(Plane.MYRROR, 5, 800, neutralPlayer);

    for ((plane, mapping) <- List((Plane.ARCANUS, arcanus), (Plane.MYRROR, myrror))) {
      for (y <- 0 until HEIGHT) {
        for (x <- 0 until WIDTH) {
          var t = overworld.get(plane, x, y);
            if (y > 0 && y < HEIGHT-1) {
              val recommendation =
                TerrainTileMetadata.recommendedTerrainChange(plane,
                  CardinalDirection.valuesAll map {(d) => overworld.get(plane, x + d.dx + WIDTH, y + d.dy).terrainType},
                  mapping(t.terrainType));
              recommendation match {
                case (terrain, sprite) => {
                  t.terrainType = terrain;
                  t.spriteNumber = sprite;
                }
              }
            } else {
              t.spriteNumber = mapping(t.terrainType);
            }
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

  var nodes = Array(List[Node](), List[Node]());
  var lairs = Array(List[Lair](), List[Lair]());
  var towers = Array(List[Lair](), List[Lair]());

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

  def scatter(random:Random,
              plane:Plane,
              count:Int,
              groundCheck:(Plane,Int,Int) => Boolean,
              placement:(Plane,Int,Int) => Unit):Unit = {
    for (n <- 0 until count) {
      findWhereTrue(random,
                    plane,
                    (x,y) => groundCheck(plane, x, y)) match {
        case (x,y) => placement(plane, x, y);
      }
    }
  }

  def grow(random:Random,
           plane:Plane,
           numSeeds:Int,
           groundCount:Int,
           base:TerrainType,
           topping:TerrainType):Unit = 
             grow(random,
                  plane,
                  numSeeds,
                  groundCount,
                  base,
                  topping, (x,y) => ());

  def grow(random:Random,
           plane:Plane,
           numSeeds:Int,
           groundCount:Int,
           base:TerrainType,
           topping:TerrainType,
           callback:(Int,Int) => Unit):Unit = {
    for(n <- 0 until numSeeds) {
      findWhereTrue(random,
                     plane,
                     (x,y) => {
                        get(plane, x, y).terrainType == base
                      }) match {
        case (x,y) => {
          var t = get(plane, x, y);
          t.terrainType = topping;
          callback(x, y);
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
    for (y <- 0 until height) {
      for (x <- 0 until width) {
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

  def buildPlane(plane:Plane, numSeeds:Int, groundCount:Int, neutralPlayer:Player):Unit = {
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
    val createNode = (random:Random, plane:Plane, x:Int, y:Int, n:TerrainType) => {
      val node = Node.createNode(random, neutralPlayer, x, y, n, 1);
      nodes(plane.id) ::= node;
      get(plane, x, y).place = Option(node);
    };

    grow(random, plane, numSeeds, 0, TerrainType.GRASSLAND, TerrainType.SORCERY_NODE,
      (x, y) => createNode(random, plane, x, y, TerrainType.SORCERY_NODE));
    grow(random, plane, numSeeds, 0, TerrainType.MOUNTAIN, TerrainType.CHAOS_NODE,
      (x, y) => createNode(random, plane, x, y, TerrainType.CHAOS_NODE));
    grow(random, plane, numSeeds, 0, TerrainType.FOREST, TerrainType.NATURE_NODE,
      (x, y) => createNode(random, plane, x, y, TerrainType.NATURE_NODE));

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

    println("Lairs");
    scatter(random, plane, numSeeds,
      (plane, x, y) => {
        val t = get(plane, x, y);
        (y > 1
          && y < height-1
          && t.terrainType != TerrainType.OCEAN
          && t.terrainType != TerrainType.SHORE
          && t.place == None);
      },
      (plane, x, y) => {
        val lair = Lair.createLair(
              random,
              neutralPlayer,
              x,
              y, 
              LairType.getRandom(random),
              1);
        lairs(plane.id) ::= lair;
        get(plane, x, y).place = Some(lair);
      });

    println("Towers");
    scatter(random, plane, numSeeds,
      (plane, x, y) => {
        val t = get(plane, x, y);
        (y > 1
          && y < height-1
          && t.terrainType != TerrainType.OCEAN
          && t.terrainType != TerrainType.SHORE
          && t.place == None);
      },
      (plane, x, y) => {
        val lair = Lair.createLair(
              random,
              neutralPlayer,
              x,
              y,
              LairType.TOWER,
              1);
        lairs(plane.id) ::= lair;
        get(plane, x, y).place = Some(lair);
      });

    println("Neutral Cities");
    scatter(random, plane, numSeeds,
      (plane, x, y) => {
        val t = get(plane, x, y);
        (y > 1
          && y < height-1
          && (t.terrainType == TerrainType.RIVER
          || t.terrainType == TerrainType.SWAMP
          || t.terrainType == TerrainType.TUNDRA
          || t.terrainType == TerrainType.MOUNTAIN
          || t.terrainType == TerrainType.HILLS
          || t.terrainType == TerrainType.GRASSLAND
          || t.terrainType == TerrainType.DESERT
          || t.terrainType == TerrainType.FOREST)
          && t.place == None);
      },
      (plane, x, y) => {
        val city = City.createNeutralCity(neutralPlayer, 
                                          x,
                                          y,
                                          "My City",
                                          Race.HIGH_MEN);
        neutralPlayer.cities ::= city;

        get(plane, x, y).place = Some(city);
      });
  }

  def isValidCityLocation(plane:Plane, x:Int, y:Int):Boolean = {
    val terrainSquare = get(plane, x, y);
    if (!terrainSquare.terrainType.canBuildCityOn) {
      return false;
    }

    for (j <- -2 to 2) {
      for (i <- -2 to 2) {
        val cx = (x + i + width) % width;
        val cy = y + j;
        if (cy >= 0 && cy < height) {
          (get(plane, cx, cy).place) match {
            case Some(city:City) => {
                return false;
            }
            case _ =>
          }
        }
      }
    }

    return true;
  }

  def getMaxCityPop(plane:Plane, x:Int, y:Int):Int = {
    var answer = 1;
    
    for (j <- -2 to 2) {
      for (i <- -2 to 2) {
        val cx = (x + i + width) % width;
        val cy = y + j;
        if (cy >= 0 && cy < height) {
          (get(plane, cx, cy).terrainType) match {
            case TerrainType.SHORE => answer += 1;
            case TerrainType.RIVER => answer += 3;
            case TerrainType.SWAMP => answer += 1;
            case TerrainType.HILLS => answer += 1;
            case TerrainType.GRASSLAND => answer += 2;
            case TerrainType.FOREST => answer += 1;
            case _ =>
          }
        }
      }
    }

    answer;
  }

  def findGoodCityLocation(random:Random, plane:Plane):Tuple2[Int, Int] = {
    var bestx:Int = 0;
    var besty:Int = 0;
    var bestpop:Int = 0;

    for (i <- 0 to 5) {
      findWhereTrue(random, plane, (x,y) => isValidCityLocation(plane, x, y)) match {
        case (x,y) =>
          val pop = getMaxCityPop(plane, x, y);
          if (pop > bestpop) {
            bestx = x;
            besty = y;
            bestpop = pop;
          }
      }
    }

    (bestx,besty);
  }

  def createCityAt(
    plane:Plane,
    x:Int,
    y:Int,
    player:Player,
    race:Race,
    startingPopulation:Int):Unit = {

    val city = new City(x, y, player, "New City", race);

    get(plane, x, y).place = Some(city);
  }
}