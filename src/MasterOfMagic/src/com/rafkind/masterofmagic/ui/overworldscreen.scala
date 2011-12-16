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
  var backgroundImage:Image = null;
  var minimap:Minimap = null;
  
  override def getID() = id;

  override def init(container:GameContainer, game:StateBasedGame):Unit = {
    terrainPainter = new TerrainPainter(
      TerrainPainter.createDummySpriteSheetImage());

    minimap = new Minimap(overworld);
    minimap.generateMinimapImage(overworld);
    
    this.initNifty();
    this.loadXml("com/rafkind/masterofmagic/ui/overworld-screen.xml");

    backgroundImage = new Image("../../data/img/overworld-example-double.png");
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

  override def render(
    container:GameContainer,
    game:StateBasedGame,
    graphics:Graphics):Unit = {

    backgroundImage.draw(0, 0);

    super.render(container, game, graphics);

    terrainPainter.render(
      container,
      graphics,
      0,
      40,
      0,
      0, overworld);

    minimap.renderMiniMap(500, 40, 0, 0, 120, 64);
  }
}