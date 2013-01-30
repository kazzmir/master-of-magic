/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic

import org.newdawn.slick._;
import org.newdawn.slick.state._;

import com.rafkind.masterofmagic.ui._;
import com.rafkind.masterofmagic.ui.main._;
import com.rafkind.masterofmagic.state._;

import com.google.inject._;

class MasterOfMagic @Inject() (playingGameState:PlayingGameState) extends StateBasedGame("Master of Magic") {
  var overworld:Overworld = null;
  
  var players:Array[Player] = null;
  
  def getOverworld = overworld;

  override def initStatesList(container:GameContainer):Unit = {
    
    players = Array(
      new Player("Raiders", FlagColor.BROWN, Race.HIGH_MEN),
      new Player("Abe", FlagColor.RED, Race.HIGH_MEN),
      new Player("Bob", FlagColor.BLUE, Race.HIGH_MEN),
      new Player("Cam", FlagColor.GREEN, Race.HIGH_MEN),
      new Player("Don", FlagColor.YELLOW, Race.HIGH_MEN),
      new Player("Erl", FlagColor.PURPLE, Race.HIGH_MEN)
    );
      
    overworld = Overworld.create(players);

  
    //addState(new OverworldMapState(0, overworld));

    addState(playingGameState);
  }
}
