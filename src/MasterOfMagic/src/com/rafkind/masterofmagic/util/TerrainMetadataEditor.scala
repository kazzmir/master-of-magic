/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util

import org.newdawn.slick._;

import com.rafkind.masterofmagic.system._;
import com.rafkind.masterofmagic.state._;

class TerrainMetadataEditor(title:String) extends BasicGame(title) {
  import com.rafkind.masterofmagic.util.TerrainLbxReader._;
  
  var terrainTileSheet:Image = null;

  var currentTile:Int = 0;

  var currentDirection:CardinalDirection = CardinalDirection.CENTER;

  val BIG_WIDTH = 133;
  val BIG_HEIGHT = BIG_WIDTH * TILE_HEIGHT / TILE_WIDTH;

  val uiColor = Color.white;

  override def init(container:GameContainer):Unit = {
    terrainTileSheet = TerrainLbxReader.read(Data.originalDataPath("TERRAIN.LBX"));
  }

  override def update(container:GameContainer, delta:Int):Unit = {

  }

  override def render(container:GameContainer, graphics:Graphics):Unit = {

    val tX = (currentTile % SPRITE_SHEET_WIDTH) * TILE_WIDTH;
    val tY = (currentTile / SPRITE_SHEET_WIDTH) * TILE_HEIGHT;

    val dX = (container.getWidth()-BIG_WIDTH)/2;
    val dY = (container.getHeight()-BIG_HEIGHT)/2;

    terrainTileSheet.draw(
      dX, dY, dX + BIG_WIDTH, dY + BIG_HEIGHT,
      tX, tY, tX + TILE_WIDTH, tY + TILE_HEIGHT
    );

    // draw direction box
    val cX = container.getWidth()/2;
    val cY = container.getHeight()/2;
    
    val x1 = cX + BIG_WIDTH * currentDirection.dx - (10 + BIG_WIDTH/2);
    val y1 = cY + BIG_HEIGHT * currentDirection.dy - (10 + BIG_HEIGHT / 2);
    
    val x2 = cX + BIG_WIDTH * currentDirection.dx + (10 + BIG_WIDTH/2);
    val y2 = cY + BIG_HEIGHT * currentDirection.dy + (10 + BIG_HEIGHT / 2);

    graphics.setColor(uiColor);
    graphics.drawRect(x1, y1, x2-x1, y2-y1);

  }
}

object TerrainMetadataEditor {
  def main(args: Array[String]): Unit = {
    var app = new AppGameContainer(new TerrainMetadataEditor("Master of Magic: Terrain Metadata Editor"));
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(640, 400, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app.start();
  }
}
