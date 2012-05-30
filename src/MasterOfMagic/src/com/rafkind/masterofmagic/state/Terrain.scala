/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

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
  val parentId:Option[TerrainTileMetadata]) {

  def matches(plane:Plane, terrains:Array[TerrainType]):Boolean = {
    terrainType match {
      case TerrainType.OCEAN | TerrainType.SHORE => {          
        return (borderingTerrainTypes zip terrains foldLeft true){
          (accum, pair) =>            
            pair match {
              case (Some(TerrainType.OCEAN), TerrainType.OCEAN) => accum && true;
              case (Some(TerrainType.SHORE), TerrainType.OCEAN) => accum && true;
              case (Some(TerrainType.OCEAN), TerrainType.SHORE) => accum && true;
              case (Some(TerrainType.SHORE), TerrainType.SHORE) => accum && true;
              case (Some(TerrainType.RIVER), TerrainType.RIVER) => accum && true;
              case (Some(s1), s2) if (s1 != TerrainType.OCEAN
                                      && s1 != TerrainType.SHORE
                                      && s1 != TerrainType.RIVER
                                      && s2 != TerrainType.OCEAN
                                      && s2 != TerrainType.SHORE
                                      && s2 != TerrainType.RIVER) => accum && true;

              case _ => accum && false;

            }       
        };
      }
      case TerrainType.RIVER => {
          return ((CardinalDirection.values, borderingTerrainTypes, terrains).zipped foldLeft true) {
            (accum, triplet) =>
              triplet match {
                case (dir, x, y) if (dir == CardinalDirection.NORTH_EAST
                                    || dir == CardinalDirection.NORTH_WEST
                                    || dir == CardinalDirection.SOUTH_EAST
                                    || dir == CardinalDirection.SOUTH_WEST) => accum && true;
                case (dir, Some(TerrainType.RIVER), TerrainType.RIVER)  => accum && true;
                case (dir, Some(TerrainType.RIVER), TerrainType.SHORE)  => accum && true;
                case (dir, Some(TerrainType.RIVER), TerrainType.OCEAN)  => accum && true;
                case (dir, Some(s1), s2) if (s1 != TerrainType.RIVER
                                        && s2 != TerrainType.RIVER) => accum && true;
                case _ => accum && false;
              }
          }
      }
      case r if (r == TerrainType.HILLS
                 || r == TerrainType.MOUNTAIN) => {

          return ((CardinalDirection.values, borderingTerrainTypes, terrains).zipped foldLeft true) {
            (accum, triplet) =>
              triplet match {
                case (dir, x, y) if (dir == CardinalDirection.NORTH_EAST
                                    || dir == CardinalDirection.NORTH_WEST
                                    || dir == CardinalDirection.SOUTH_EAST
                                    || dir == CardinalDirection.SOUTH_WEST) => accum && true;
                case (dir, Some(t1), t2) if (t1 == r && t2 == r) => accum && true;
                case (dir, Some(s1), s2) if (s1 != r
                                        && s2 != r) => accum && true;
                case _ => accum && false;
              }
          }
      }
        
      case target =>
        return (borderingTerrainTypes zip terrains foldLeft true){
          (accum, pair) =>
            pair match {
              case (Some(t1), t2) if (t1 == target && t2 == target) => accum && true;
              case (Some(s1), s2) if (s1 != target
                                      && s2 != target) => accum && true;
              case _ => accum && false;

            }
        };
    }
    return false;
  }
}

object TerrainTileMetadata {
  import com.rafkind.masterofmagic.util.TerrainLbxReader._;
  import scala.xml._;
  import com.rafkind.masterofmagic.system._;
  import com.rafkind.masterofmagic.util._;
  import scala.collection.mutable._;

  /*var data:Array[TerrainTileMetadata] =
    new Array[TerrainTileMetadata](TILE_COUNT);
  */
  var data = new CustomMultiMap[Tuple2[Plane, Int], TerrainTileMetadata];


  read(Data.path("terrainMetaData.xml"));

  def read(fn:String):Unit = {
    XML.load(fn) \ "metadata" foreach { (m) =>
      val borders = new Array[Option[TerrainType]](CardinalDirection.values.length);
      m \ "borders" foreach { (b) =>
        borders(Integer.parseInt((b \ "@direction").text)) =
          Some(TerrainType.values(Integer.parseInt((b \ "@terrain").text)));
      }

      val id = Integer.parseInt((m \ "@id").text)
      val terrainType = Integer.parseInt((m \ "@terrainType").text);
      val plane = Integer.parseInt((m \ "@plane").text);
      var metadata = new TerrainTileMetadata(id,
                                          TerrainType.values(terrainType),
                                          borders,
                                          Plane.values(plane), None);

      plane match {
        case Plane.ARCANUS.id => data.put((Plane.ARCANUS, terrainType), metadata);
        case Plane.MYRROR.id => data.put((Plane.MYRROR, terrainType), metadata);
        case _ => 
      }      
    }
  }

  def setCombine[T](s1:Option[Set[T]], s2:Option[Set[T]]):Set[T] = {
    (s1, s2) match {
      case (None, None) =>
        new HashSet[T];
      case (Some(set1), None) =>
        set1;
      case (None, Some(set2)) =>
        set2;
      case (Some(set1), Some(set2)) =>
        set1 ++ set2;
    }
  }

  def recommend(plane:Plane,
          terrain:Array[TerrainType],
          meta:Set[TerrainTileMetadata],
          target:TerrainType,
          default:Int):Tuple2[TerrainType, Int] = {

    val soFar = ((target, default), false);
    return meta.foldLeft(soFar)((acc, metadata) =>
      (acc) match {
        case((terr, id), true) =>
          ((terr, id), true)
        case((terr, id), false) =>
          if (metadata.matches(plane, terrain)) {
            ((metadata.terrainType, metadata.id), true)
          } else {
            ((terr, id), false)
          }
      })._1;
  }

  def recommendedTerrainChange(plane:Plane, terrain:Array[TerrainType], default:Int):Tuple2[TerrainType, Int] = {
    terrain(CardinalDirection.CENTER.id) match {
      case TerrainType.OCEAN => {
        val oceans = setCombine(data.get((plane, TerrainType.OCEAN.id)),
                                data.get((plane, TerrainType.SHORE.id)));
         return recommend(plane, terrain, oceans, TerrainType.OCEAN, default);
      }
      case t if (t == TerrainType.TUNDRA 
                 || t == TerrainType.HILLS
                 || t == TerrainType.MOUNTAIN
                 || t == TerrainType.RIVER
                 || t == TerrainType.DESERT)  => {
          return recommend(plane, 
                           terrain,
                           setCombine(data.get((plane, t.id)), None),
                           t,
                           default);
      }
      case x => return (x, default);
    }
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
  var unitStack:Option[ArmyUnitStack]) {

  // what type of terrain
  // what terrain tile to use
  // bitset for fog of war
  // polluted?
  // what type of bonus
  // road?
  // what city is here?
  // what unit stack is here?
}
