/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic

import org.newdawn.slick._;
import org.newdawn.slick.state._;

import de.lessvoid.nifty._;
import de.lessvoid.nifty.tools._;
import de.lessvoid.nifty.lwjglslick.render._;
import de.lessvoid.nifty.lwjglslick.sound._;
import de.lessvoid.nifty.lwjglslick.input._;

import com.rafkind.masterofmagic.ui._;
import com.rafkind.masterofmagic.state._;

class OverworldMapState extends BasicGameState {

  var nifty:Nifty = null;

  override def getID() = 0;

  override def init(container:GameContainer, game:StateBasedGame):Unit = {
    nifty = new Nifty(
      new RenderDeviceLwjgl(),
      new SlickSoundDevice(),
      new LwjglInputSystem(),
      new TimeProvider()
    );
  }

  override def update(
    container:GameContainer,
    game:StateBasedGame,
    delta:Int):Unit = {

  }

  override def render(
    container:GameContainer,
    game:StateBasedGame,
    graphics:Graphics):Unit = {

  }
}

class MasterOfMagic(title:String) extends StateBasedGame(title) {

  /*var terrainPainter:TerrainPainter = null;
  var overworld:Overworld = null;
*/

  addState(new OverworldMapState());

  /*override def init(gc:GameContainer):Unit = {
    terrainPainter = new TerrainPainter(
      TerrainPainter.createDummySpriteSheetImage());

    overworld = Overworld.createExampleWorld;
  }

  override def update(gc:GameContainer, delta:Int):Unit = {
  }

  override def render(gc:GameContainer, graphics:Graphics):Unit = {

    terrainPainter.render(gc, graphics, 0, 0, 0, 0, overworld);
  }*/

  override def initStatesList(container:GameContainer):Unit = {
    
  }
}
