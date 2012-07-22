/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui
/*
import com.rafkind.masterofmagic.state._
import com.rafkind.masterofmagic.system._
import com.rafkind.masterofmagic.util._

import org.newdawn.slick._
import org.newdawn.slick.state._

import de.lessvoid.nifty._
import de.lessvoid.nifty.slick._
import de.lessvoid.nifty.screen._
import de.lessvoid.nifty.input.keyboard.KeyboardInputEvent

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

  var x = 0;
  var y = 0;

  override def init(container:GameContainer, game:StateBasedGame):Unit = {
    terrainPainter = new TerrainPainter(
      TerrainLbxReader.read(Data.originalDataPath("TERRAIN.LBX")));

    minimap = new Minimap(overworld);
    minimap.generateMinimapImage(overworld);
    
    this.initNifty();
    this.loadXml("com/rafkind/masterofmagic/ui/overworld-screen.xml");

    backgroundImage = new Image(Data.path("img/overworld-example-double.png"));
    //backgroundImage = TerrainLbxReader.read(Data.originalDataPath("TERRAIN.LBX"));

  }

  override def processKeyboardEvent(e:KeyboardInputEvent):Boolean = {
    if (e.isKeyDown()) {
      e.getKey() match {
        case KeyboardInputEvent.KEY_UP => y = if (y > 0) y-1 else 0;
        case KeyboardInputEvent.KEY_DOWN => y = if (y < overworld.height-TerrainPainter.VIEW_HEIGHT) { y+1 } else {overworld.height-TerrainPainter.VIEW_HEIGHT};
        case KeyboardInputEvent.KEY_LEFT => x = if (x > 0) {x-1} else {overworld.width-1};
        case KeyboardInputEvent.KEY_RIGHT => x += 1;
        case _ => {}
      }
    }
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
      x,
      y, overworld);

    minimap.renderMiniMap(500, 40, 0, 0, 120, 64);
  }
}
*/