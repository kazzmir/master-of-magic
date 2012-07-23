/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic

import org.newdawn.slick._;
import org.newdawn.slick.state._;

import com.rafkind.masterofmagic.ui._;
import com.rafkind.masterofmagic.state._;

import com.google.inject._;

class MasterOfMagic @Inject() (playingGameState:PlayingGameState) extends StateBasedGame("Master of Magic") {
  var overworld:Overworld = null;  

  override def initStatesList(container:GameContainer):Unit = {

    //overworld = Overworld.create(new Player("Raiders", FlagColor.BROWN, Race.HIGH_MEN));

    //addState(new OverworldMapState(0, overworld));

    addState(playingGameState);
  }
}
