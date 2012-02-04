/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic

import org.newdawn.slick._;
import org.newdawn.slick.state._;

import com.rafkind.masterofmagic.ui._;
import com.rafkind.masterofmagic.state._;


class MasterOfMagic(title:String) extends StateBasedGame(title) {
  var overworld:Overworld = null;  

  override def initStatesList(container:GameContainer):Unit = {

    overworld = Overworld.create;

    addState(new OverworldMapState(0, overworld));
  }
}
