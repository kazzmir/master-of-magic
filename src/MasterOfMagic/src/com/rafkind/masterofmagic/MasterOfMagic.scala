/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic

import org.newdawn.slick.BasicGame;
import org.newdawn.slick.GameContainer;
import org.newdawn.slick.Graphics;
import org.newdawn.slick.SlickException;

import com.rafkind.masterofmagic.ui.TerrainTilePainter;

class MasterOfMagic(title:String) extends BasicGame(title) {

  var terrainTilePainter:TerrainTilePainter = null;

  override def init(gc:GameContainer):Unit = {
    terrainTilePainter = new TerrainTilePainter(
      TerrainTilePainter.createDummySpriteSheetImage());
  }

  override def update(gc:GameContainer, delta:Int):Unit = {
  }

  override def render(gc:GameContainer, graphics:Graphics):Unit = {
    terrainTilePainter.render(gc, graphics, 0, 0);
  }
}
