/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state


case class Race(val id:Int, val name:String)
object Race {
  val BARBARIAN = Race(0, "Barbarian");
  val GNOLL = Race(1, "Gnoll");
  val HALFLING = Race(2, "Halfling");
  val HIGH_ELF = Race(3, "High Elf");
  val HIGH_MEN = Race(4, "High Man");
  val KLACKON = Race(5, "Klackon");
  val LIZARDMAN = Race(6, "Lizardman");
  val NOMAD = Race(7, "Nomad");
  val ORC = Race(8, "Orc");
  val BEASTMAN = Race(9, "Beastman");
  val DARK_ELF = Race(10, "Dark Elf");
  val DRACONIAN = Race(11, "Draconian");
  val DWARVEN = Race(12, "Dwarven");
  val TROLL = Race(13, "Troll");

  val values = Array(BARBARIAN,
                     GNOLL,
                     HALFLING,
                     HIGH_ELF,
                     HIGH_MEN,
                     KLACKON,
                     LIZARDMAN,
                     NOMAD,
                     ORC,
                     BEASTMAN,
                     DARK_ELF,
                     DRACONIAN,
                     DWARVEN,
                     TROLL);

  val valuesByPlane = Array(
    Array(BARBARIAN,
         GNOLL,
         HALFLING,
         HIGH_ELF,
         HIGH_MEN,
         KLACKON,
         LIZARDMAN,
         NOMAD,
         ORC),
    Array(BEASTMAN,
         DARK_ELF,
         DRACONIAN,
         DWARVEN,
         TROLL));

  implicit def race2string(r:Race) = r.name;
}