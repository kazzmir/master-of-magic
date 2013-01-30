/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state


class Retort {
  
}



class Player(val name:String, val flag:FlagColor, val race:Race) {
  // id
  // name
  // picture
  // color scheme
  // music scheme

  var armyUnits:List[ArmyUnit] = List();

  var armyUnitStacks:List[ArmyUnitStack] = List();
  
  var cities:List[City] = List();

  var gold:Int = 0;
  var mana:Int = 0;

  var deltaGold:Int = 0;
  var deltaMana:Int = 0;
}
