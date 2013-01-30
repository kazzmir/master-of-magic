/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.state

import scala.util.Random;
import com.rafkind.masterofmagic.util._

case class Alignment(val id:Int, val name:String)
object Alignment {
  val HORIZONTAL = Alignment(0, "HORIZONTAL");
  val VERTICAL = Alignment(1, "VERTICAL");

  val values = Array(HORIZONTAL, VERTICAL);
}

case class FlagColor(val id:Int, 
                     val name:String, 
                     val paletteIndexOffset:Int, 
                     val armyStackBackgroundSpriteIndex:Int)
object FlagColor {
  val RED = FlagColor(0, "Red", -15, 17);
  val BLUE = FlagColor(1, "Blue", 5, 14);
  val GREEN = FlagColor(2, "Green", 0, 15);
  val YELLOW = FlagColor(3, "Yellow", -5, 18);
  val PURPLE = FlagColor(4, "Purple", -10, 16);
  val BROWN = FlagColor(5, "Brown", -33, 19);

  val values = Array(BROWN, RED, GREEN, BLUE, YELLOW, PURPLE);
}

case class MagicColor(val id:Int, val name:String)
object MagicColor {
  val ARCANE  = MagicColor(0, "Arcane");
  val WHITE = MagicColor(1, "White");
  val GREEN = MagicColor(2, "Green");
  val RED = MagicColor(3, "Red");
  val BLUE = MagicColor(4, "Blue");
  val BLACK = MagicColor(5, "Black");
}

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
  val valuesAll = Array(NORTH, NORTH_EAST, EAST, SOUTH_EAST, SOUTH, SOUTH_WEST, WEST, NORTH_WEST, CENTER);

  def opposite(dir:CardinalDirection) =
    dir match {
      case NORTH => SOUTH;
      case NORTH_EAST => SOUTH_WEST;
      case EAST => WEST;
      case SOUTH_EAST => NORTH_WEST;
      case SOUTH => NORTH;
      case SOUTH_WEST => NORTH_EAST;
      case WEST => EAST;
      case NORTH_WEST => SOUTH_EAST;
      case _ => CENTER;
    }
}

case class Plane(val id:Int, val name:String)
object Plane {
  val ARCANUS = Plane(0, "Arcanus")
  val MYRROR = Plane(1, "Myrror")

  val values = Array(ARCANUS, MYRROR)

  implicit def plane2string(p:Plane) = p.name  
}

// http://www.dragonsword.com/magic/eljay/SaveGam.html

class GameInitializationParameters(val numberOfPlayers:Int) {
  
}

object State {
  def createGameState(initParams:GameInitializationParameters):State = {
    val state = new State(initParams);

    state;
  }
}

class State(initParams:GameInitializationParameters) {
  // players
  val allPlayers = new Array[Player](initParams.numberOfPlayers + 1);
  allPlayers(0) = new Player("Raiders", FlagColor.BROWN, Race.HIGH_MEN);

  val _overworld = Overworld.create(allPlayers);
  def overworld = _overworld;
  
  var random = new Random();
  for (p <- 1 to initParams.numberOfPlayers) {
    val player = new Player("Player" + p, FlagColor.values(p), Race.HIGH_MEN);

    // create a city for the player
    overworld.findGoodCityLocation(random, Plane.ARCANUS) match {
      case (x,y) =>
        overworld.createCityAt(Plane.ARCANUS, x, y, player, player.race, 3000);
    }
    allPlayers(p) = player;    
  }
}
