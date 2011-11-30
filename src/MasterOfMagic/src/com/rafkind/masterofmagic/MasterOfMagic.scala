/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic

import org.newdawn.slick.BasicGame;
import org.newdawn.slick.GameContainer;
import org.newdawn.slick.Graphics;
import org.newdawn.slick.SlickException;

import com.rafkind.masterofmagic.ui.TerrainPainter;
import com.rafkind.masterofmagic.state._;

class MasterOfMagic(title:String) extends BasicGame(title) {

  var terrainPainter:TerrainPainter = null;
  var overworld:Overworld = null;

  override def init(gc:GameContainer):Unit = {
    terrainPainter = new TerrainPainter(
      TerrainPainter.createDummySpriteSheetImage());

    overworld = Overworld.createExampleWorld;
  }

  override def update(gc:GameContainer, delta:Int):Unit = {
  }

  override def render(gc:GameContainer, graphics:Graphics):Unit = {

    terrainPainter.render(gc, graphics, 0, 0, 0, 0, overworld);
  }
}
