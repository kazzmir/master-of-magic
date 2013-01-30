/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

import scala.util.Random;
import com.rafkind.masterofmagic.util._

case class LairType(val id:Int, val name:String)

object LairType {
  val CAVE = LairType(0, "Cave");
  val TEMPLE = LairType(1, "Temple");
  val KEEP = LairType(2, "Keep");
  val TOWER = LairType(3, "Tower");
  val NODE = LairType(4, "Node");
  
  val values = Array(CAVE, TEMPLE, KEEP);

  def getRandom(random:Random):LairType =
    values(random.nextInt(values.length));

  def spriteGroup(id:Int) = id match {
    case CAVE.id => 71
    case TEMPLE.id => 72
    case KEEP.id => 73
    // closed tower is 69, open is 70
    case TOWER.id => 69
    // nodes are drawn by the tile so this shouldn't really be here
    case NODE.id => 64
  }

  // all lair types use sprite 0
  def spriteItem(id:Int) = 0
}

class Place(val x:Int, val y:Int) {

}

object City {
  def createNeutralCity(
                 owner:Player,
                 x:Int,
                 y:Int,
                 name:String,
                 race:Race):City = {
    return new City(x, y, owner, name, race);
  }

  val city1Group = 20
  val city1Item = 0

  val city2Group = 20
  val city2Item = 1

  val city3Group = 20
  val city3Item = 2

  val city4Group = 20
  val city4Item = 3
  
  val city5Group = 20
  val city5Item = 5

  val village1Group = 21
  val village1Item = 0
  
  val village2Group = 21
  val village2Item = 1
  
  val village3Group = 21
  val village3Item = 2
  
  val village4Group = 21
  val village4Item = 3
  
  val village5Group = 21
  val village5Item = 4
}

class City(x:Int, y:Int, val owner:Player, val name:String, val race:Race) extends Place(x, y) {
  def getSprite(librarian:ImageLibrarian) = {
    //SpriteReaderHelper.turnLoggingOn;
    val s = librarian.getFlaggedSprite(OriginalGameAsset.MAPBACK, City.city1Group, City.city1Item, owner.flag)
    //SpriteReaderHelper.turnLoggingOff;
    s
  }
}

object LairReward {
  
  // powerlevel from 0 to 10
  def createReward(random:Random, powerLevel:Int):LairReward = {
    return new LairReward(10, 0, 0, 0, 0, 0, 0, 0);
  }
}
class LairReward(
  val gold:Int,
  val mana:Int,
  val spellLevel:Int,
  val spellBooksCount:Int,
  val prisonerHeroLevel:Int,
  val itemLevel:Int,
  val itemCount:Int,
  val retortCount:Int
  ) {

}

object Lair {
  def createLair(random:Random,
                 owner:Player,
                 x:Int,
                 y:Int,
                 lairType:LairType,
                 powerLevel:Int):Lair = {
    val nativeUnits = 
      new ArmyUnitStack(
        x, 
        y, 
        owner, 
        List(
          new ArmyUnit(ArmyUnitType.dummy)));
    val reward = LairReward.createReward(random, powerLevel);
    return new Lair(lairType, nativeUnits, reward, x, y);
  }
}
class Lair(val lairType:LairType,
           val nativeUnits:ArmyUnitStack,
           val reward:LairReward,
           x:Int,
           y:Int) extends Place(x, y) {

  def getSprite(librarian:ImageLibrarian) =
    librarian.getRawSprite(OriginalGameAsset.MAPBACK,
                           LairType.spriteGroup(lairType.id),
                           LairType.spriteItem(lairType.id))
}

object Node {
  def createNode(random:Random, owner:Player, x:Int, y:Int, nodeType:TerrainType, powerLevel:Int):Node = {
    val nativeUnits = 
      new ArmyUnitStack(
        x, 
        y, 
        owner, 
        List(
          new ArmyUnit(ArmyUnitType.dummy)));
    val reward = LairReward.createReward(random, powerLevel);
    return new Node(nodeType, nativeUnits, reward, x, y);
  }
}
class Node(val nodeType:TerrainType,
           nativeUnits:ArmyUnitStack,
           reward:LairReward,
           x:Int,
           y:Int) extends Lair(LairType.NODE, nativeUnits, reward, x, y) {
  
}
