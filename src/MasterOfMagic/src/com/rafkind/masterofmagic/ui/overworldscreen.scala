/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui

import com.rafkind.masterofmagic.state._;

import org.newdawn.slick._;
import org.newdawn.slick.state._;

import de.lessvoid.nifty._;
import de.lessvoid.nifty.slick._;
import de.lessvoid.nifty.screen._;
import de.lessvoid.nifty.input.keyboard.KeyboardInputEvent;

class OverworldMapScreenController extends ScreenController {
  override def onEndScreen():Unit = {}
  override def onStartScreen():Unit = {}
  override def bind(nifty:Nifty, screen:Screen):Unit = {}
}

class OverworldMapState(id:Int, overworld:Overworld) extends NiftyOverlayGameState {
  var terrainPainter:TerrainPainter = null;
  
  override def getID() = id;

  override def init(container:GameContainer, game:StateBasedGame):Unit = {
    //super.init(container, game);
    
    terrainPainter = new TerrainPainter(
      TerrainPainter.createDummySpriteSheetImage());
    
    this.initNifty();
    this.loadXml("com/rafkind/masterofmagic/ui/overworld-screen.xml");
  }

  override def processKeyboardEvent(e:KeyboardInputEvent):Boolean = {
    return true;
  }

  override def processMouseEvent(
    mouseX:Int,
    mouseY:Int,
    mouseWheel:Int,
    button:Int,
    buttonDown:Boolean):Boolean = {
    return true;
  }

  override def render(container:GameContainer, game:StateBasedGame, graphics:Graphics):Unit = {
    super.render(container, game, graphics);

    terrainPainter.render(container, graphics, 0, 0, 0, 0, overworld);
  }
}