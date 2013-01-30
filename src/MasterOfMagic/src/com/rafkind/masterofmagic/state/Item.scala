/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

case class ItemType(val id:Int, val name:String)
object ItemType {
  val WEAPON = ItemType(0, "Weapon");
  val ARMOR = ItemType(1, "Armor");
  val MISC = ItemType(2, "Miscellaneous");
  
  val values = Array(WEAPON, ARMOR, MISC);
}

class Item {

}
