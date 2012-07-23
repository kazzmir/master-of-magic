/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui
import org.newdawn.slick.state._;
import org.newdawn.slick._;

import com.rafkind.masterofmagic.util._;

import com.google.inject._;

class PlayingGameState @Inject() (imageLibrarian:ImageLibrarian) extends BasicGameState {

  def getID() = 1;

  var b:Image = null;
  
  def init(container:GameContainer, game:StateBasedGame):Unit = {
    b = imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 0, 0);

  }

  def render(container:GameContainer, game:StateBasedGame, graphics:Graphics):Unit = {
    b.draw(0, 0);

  }

  def update(container:GameContainer, game:StateBasedGame, delta:Int):Unit = { }
}
