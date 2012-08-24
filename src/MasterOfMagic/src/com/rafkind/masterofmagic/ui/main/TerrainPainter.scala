/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.main

import java.awt.Color;
import org.newdawn.slick._;

import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.util._;
import com.google.inject._;

object TerrainPainter {

  import com.rafkind.masterofmagic.util.TerrainLbxReader._;

  // size of the big viewport, in tiles
  val VIEW_WIDTH = 12;
  val VIEW_HEIGHT = 10;

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

@Singleton
class TerrainPainter(baseTileImage:Image, librarian:ImageLibrarian) {
  import com.rafkind.masterofmagic.util.TerrainLbxReader._;

  def render(
    graphics:Graphics,
    startX:Int,
    startY:Int,
    startTileX:Int,
    startTileY:Int,
    tilesAcross:Int,
    tilesDown:Int,
    plane:Plane,
    overworld:Overworld):Unit = {

    baseTileImage.startUse();

    for (tileY <- 0 until tilesDown) {
      for (tileX <- 0 until tilesAcross) {
        val terrainSquare:TerrainSquare = overworld.get(
          plane,
          tileX + startTileX,
          tileY + startTileY);

        val whichTile:Int = terrainSquare.spriteNumber;

        val tX = (whichTile % TerrainLbxReader.SPRITE_SHEET_WIDTH) * TILE_WIDTH;
        val tY = (whichTile / TerrainLbxReader.SPRITE_SHEET_WIDTH) * TILE_HEIGHT;
        val dX = startX + tileX * TILE_WIDTH;
        val dY = startY + tileY * TILE_HEIGHT;

        baseTileImage.drawEmbedded(
          dX, dY, dX + TILE_WIDTH, dY + TILE_HEIGHT,
          tX, tY, tX + TILE_WIDTH, tY + TILE_HEIGHT
        );
      }
    }
    baseTileImage.endUse();

    for (tileY <- 0 until tilesDown) {
      for (tileX <- 0 until tilesAcross) {
        val terrainSquare:TerrainSquare = overworld.get(
          plane,
          tileX + startTileX,
          tileY + startTileY);

        val dX = startX + tileX * TILE_WIDTH;
        val dY = startY + tileY * TILE_HEIGHT;

        terrainSquare.place match {
          case Some(city:City) => {
            // println("Draw city from %d, %d at %d, %d".format(city.x, city.y, dX, dY))
            val citySprite = city.getSprite(librarian);
            
            citySprite.draw(dX + (TILE_WIDTH-citySprite.getWidth())/2, dY + (TILE_HEIGHT-citySprite.getHeight())/2)
          }
          case Some(node:Node) => {
          }
          case Some(lair:Lair) => {
            lair.getSprite(librarian).draw(dX, dY)
          }
          case None => {
          }
        }
      }
    }
  }
}
