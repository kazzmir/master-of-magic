/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui
import org.newdawn.slick.state._;
import org.newdawn.slick._;

import com.rafkind.masterofmagic.util._;
import com.rafkind.masterofmagic.ui.framework._;

import com.google.inject._;

class PlayingGameState @Inject() (imageLibrarian:ImageLibrarian) extends BasicGameState {

  var mainScreen:Screen = null;
  var currentScreen:Screen = null;
  
  def getID() = 1;
  
  def init(container:GameContainer, game:StateBasedGame):Unit = {

    mainScreen = new Screen(imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 0, 0));
    currentScreen = mainScreen;
  }

  def render(container:GameContainer, game:StateBasedGame, graphics:Graphics):Unit = {
    currentScreen.render(graphics);
  }

  def update(container:GameContainer, game:StateBasedGame, delta:Int):Unit = { }
}
