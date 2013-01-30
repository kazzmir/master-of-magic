/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

import com.rafkind.masterofmagic.util._

class Characteristic(val name:String, val spriteIndex:Int) {
  
}

object ArmyUnitType {
  // todo(dave): take this out
  val dummy = new ArmyUnitType(
    "Dummy", 
    new SpriteKey(OriginalGameAsset.UNITS1, 0, 0),
    new SpriteKey(OriginalGameAsset.UNITS1, 0, 0),
    1,
    null,
    1,
    0,
    0,
    1,
    1,
    1,
    0,
    0,
    0
  );
}

class ArmyUnitType(
  val name:String, 
  val overworldSprite:SpriteKey, 
  val combatBaseSprite:SpriteKey, 
  val combatBaseFigures:Int,
  val defaultCharacteristics:Map[Characteristic, Int],
  val upkeepGold:Int,
  val upkeepMana:Int,
  val upkeepFood:Int,
  val movement:Int,
  val attack:Int,
  val defense:Int,
  val loRange:Int,
  val midRange:Int,
  val hiRange:Int) {
}

class ArmyUnit(val unitType:ArmyUnitType) {  
  def overworldSprite = unitType.overworldSprite;
}

class Hero(unitType:ArmyUnitType, 
           val itemSlots:List[ItemType]) extends
  ArmyUnit(unitType) {

}

class ArmyUnitStack(var _x:Int, var _y:Int, var _owner:Player, var _units:List[ArmyUnit]) {
  def x = _x;
  def x_=(v:Int):Unit = _x = v;

  def y = _y;
  def y_=(v:Int):Unit = _y = v;

  def owner = _owner;
  def owner_=(v:Player):Unit = _owner = v;
  
  def units = _units;
  
  def getBackgroundSprite(librarian:ImageLibrarian) = {    
    librarian.getRawSprite(
      OriginalGameAsset.MAPBACK, 
      owner.flag.armyStackBackgroundSpriteIndex, 
      0);
  }
  
  def getForegroundSprite(librarian:ImageLibrarian) = {
    val first = _units(0);
    
    librarian.getFlaggedSprite(first.overworldSprite, owner.flag);
  }
}
