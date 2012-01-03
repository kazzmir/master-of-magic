/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui

import java.awt.Color;
import org.newdawn.slick._;

import com.rafkind.masterofmagic.state._;

object TerrainPainter {

  import com.rafkind.masterofmagic.util.TerrainLbxReader._;

  // size of the big viewport, in tiles
  val VIEW_WIDTH = 12;
  val VIEW_HEIGHT = 11;

  def createDummySpriteSheetImage():Image = {
    
    var imageBuffer = new ImageBuffer(
      TILE_WIDTH * 3,
      TILE_HEIGHT * 3);

    List(
      Color.BLACK,
      Color.BLUE,
      Color.GREEN,
      Color.RED,
      Color.YELLOW,
      Color.GREEN.brighter().brighter(),
      Color.BLUE.brighter().brighter(),
      Color.WHITE,
      Color.MAGENTA).zipWithIndex.foreach{
      case (color, index) => {
        var tx = index % 3;
        var ty = index / 3;
        for (y <- 0 until TILE_HEIGHT) {
          for (x <- 0 until TILE_WIDTH) {
            imageBuffer.setRGBA(
              tx * TILE_WIDTH + x,
              ty * TILE_HEIGHT + y,
              color.getRed(),
              color.getGreen(),
              color.getBlue(),
              color.getAlpha()
            );
          }
        }
      }
      };

    return imageBuffer.getImage(Image.FILTER_NEAREST);
  }
}

class Minimap(overworld:Overworld) {

  var minimapImage:Image = null;
  var minimapImageData:ImageBuffer = null;

  def miniPutpix(x:Int, y:Int, c:Color):Unit = {
    for (j <- 0 until 2) {
      for (i <- 0 until 2) {
        minimapImageData.setRGBA(x+i, y+j, c.getRed(), c.getGreen(), c.getBlue(), c.getAlpha());
      }
    }
  }

  def generateMinimapImage(overworld:Overworld):Unit = {
    minimapImageData = new ImageBuffer(Overworld.WIDTH * 2,
                                       Overworld.HEIGHT * 2);
    for (j <- 0 until Overworld.HEIGHT) {
      for (i <- 0 until Overworld.WIDTH) {
        miniPutpix(i, j, Color.GREEN);
      }
    }
    minimapImage = new Image(minimapImageData);
  }


  def renderMiniMap(startX:Int, startY:Int, offx:Int, offy:Int, width:Int, height:Int):Unit = {
    minimapImage.draw(startX, startY, startX+width, startY+height, offx, offy, offx+width, offy+height);
  }
}

class TerrainPainter(baseTileImage:Image) {  
  import com.rafkind.masterofmagic.util.TerrainLbxReader._;

  def render(
    gc:GameContainer,
    graphics:Graphics,
    startX:Int,
    startY:Int,
    startTileX:Int,
    startTileY:Int,
    overworld:Overworld):Unit = {
    
    val DOUBLE_WIDTH = TILE_WIDTH * 2;
    val DOUBLE_HEIGHT = TILE_HEIGHT * 2;

    baseTileImage.startUse();

    for (tileY <- 0 until TerrainPainter.VIEW_HEIGHT) {
      for (tileX <- 0 until TerrainPainter.VIEW_WIDTH) {

        val terrainSquare:TerrainSquare = overworld.get(
          tileX + startTileX,
          tileY + startTileY);

        val whichTile:Int = terrainSquare.spriteNumber;

        val tX = (whichTile % 3) * TILE_WIDTH;
        val tY = (whichTile / 3) * TILE_HEIGHT;
        val dX = startX + tileX * DOUBLE_WIDTH;
        val dY = startY + tileY * DOUBLE_HEIGHT;

        baseTileImage.drawEmbedded(
          dX, dY, dX + DOUBLE_WIDTH, dY + DOUBLE_HEIGHT,
          tX, tY, tX + TILE_WIDTH, tY + TILE_HEIGHT

        );
      }
    }

    baseTileImage.endUse();
  }
}