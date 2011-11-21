/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui

import java.awt.Color;
import org.newdawn.slick._;

object TerrainTilePainter {

  // double sized
  val WIDTH = 40;
  val HEIGHT = 36;

  def createDummySpriteSheetImage():Image = {
    // todo(drafkind): fix this to use awt image creation instead of
    // opengl image creation
    var imageBuffer = new ImageBuffer(WIDTH * 3, HEIGHT * 3);

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
        for (y <- 0 until HEIGHT) {
          for (x <- 0 until WIDTH) {
            imageBuffer.setRGBA(
              tx * WIDTH + x,
              ty * HEIGHT + y,
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

class TerrainTilePainter(baseTileImage:Image) {
  var baseTileSpriteSheet = new SpriteSheet(
    baseTileImage,
    TerrainTilePainter.WIDTH,
    TerrainTilePainter.HEIGHT);
}
