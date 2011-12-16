/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui

import java.awt.Color;
import org.newdawn.slick._;

import com.rafkind.masterofmagic.state._;

object TerrainPainter {

  // double sized
  val TILE_WIDTH = 40;
  val TILE_HEIGHT = 36;

  val VIEW_WIDTH = 12;
  val VIEW_HEIGHT = 11;

  def createDummySpriteSheetImage():Image = {
    var imageBuffer = new ImageBuffer(TILE_WIDTH * 3, TILE_HEIGHT * 3);

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

    return imageBuffer.getImage();
  }
}

class TerrainPainter(baseTileImage:Image) {
  var minimapImage:Image = null;
  var minimapImageData:ImageBuffer = null;
  
  var baseTileSpriteSheet = new SpriteSheet(
    baseTileImage,
    TerrainPainter.TILE_WIDTH,
    TerrainPainter.TILE_HEIGHT);

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

  def render(
    gc:GameContainer,
    graphics:Graphics,
    startX:Int,
    startY:Int,
    startTileX:Int,
    startTileY:Int,
    overworld:Overworld):Unit = {

    baseTileSpriteSheet.startUse();

    for (tileY <- 0 until TerrainPainter.VIEW_HEIGHT) {
      for (tileX <- 0 until TerrainPainter.VIEW_WIDTH) {

        var terrainSquare = overworld.get(
          tileX + startTileX,
          tileY + startTileY);

        var whichTile = terrainSquare.terrain;

        baseTileSpriteSheet.renderInUse(
          startX + tileX * TerrainPainter.TILE_WIDTH,
          startY + tileY * TerrainPainter.TILE_HEIGHT,
          whichTile % 3,
          whichTile / 3
        );
      }
    }

    baseTileSpriteSheet.endUse();
  }

  def renderMiniMap(startX:Int, startY:Int, offx:Int, offy:Int, width:Int, height:Int):Unit = {
    minimapImage.draw(startX, startY, startX+width, startY+height, offx, offy, offx+width, offy+height);
  }
}