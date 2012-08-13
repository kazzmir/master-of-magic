/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.main
import org.newdawn.slick.state._;
import org.newdawn.slick._;

import com.rafkind.masterofmagic.util._;
import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.ui.framework._;
import com.rafkind.masterofmagic.ui.main._;
import com.rafkind.masterofmagic._;
import com.rafkind.masterofmagic.system._;

import com.google.inject._;

class PlayingGameState @Inject() (imageLibrarian:ImageLibrarian, mainScreen:MainScreen) extends BasicGameState {

  var currentScreen:Screen = null;
  var terrainPainter:TerrainPainter = null;
  
  def getID() = 1;

  //var text:Image = null;

  def init(container:GameContainer, game:StateBasedGame):Unit = {

    /*val font = imageLibrarian.getFont(FontIdentifier.HUGE);
    val textib = new ImageBuffer(320, 50);
    val c = new Color(255, 255, 255, 255);
    val colors = Array(c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c);

    font.render(textib, 0, 0, colors, "THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG. PACK MY BOX WITH FIVE DOZEN LIQUOR JUGS.");
    text = textib.getImage();*/


    /*terrainPainter = new TerrainPainter(
      TerrainLbxReader.read(
        Data.originalDataPath(
          OriginalGameAsset.TERRAIN.fileName)));*/

    mainScreen.init(game.asInstanceOf[MasterOfMagic].getOverworld);
    
    currentScreen = mainScreen;
  }

  def render(container:GameContainer, game:StateBasedGame, graphics:Graphics):Unit = {
    currentScreen.render(graphics);
    //text.draw(0, 0);
  }

  def update(container:GameContainer, game:StateBasedGame, delta:Int):Unit = { }

  override def keyPressed(key:Int, c:Char):Unit = {
    currentScreen.notifyOf(Component.KEY_PRESSED, 0, 0, KeyPressedEvent(null, key, c));
  }  
}
