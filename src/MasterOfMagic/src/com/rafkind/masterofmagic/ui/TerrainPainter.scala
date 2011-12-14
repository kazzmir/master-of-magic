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

  val VIEW_WIDTH = 11;
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
  var baseTileSpriteSheet = new SpriteSheet(
    baseTileImage,
    TerrainPainter.TILE_WIDTH,
    TerrainPainter.TILE_HEIGHT);

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
}
