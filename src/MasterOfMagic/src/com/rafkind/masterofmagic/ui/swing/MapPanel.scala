/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.swing

import javax.swing.JPanel;
import java.awt._;
import java.awt.event._;
import java.awt.geom._;

import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.util._;

class MapPanel(overworld:Overworld, imageLibrarian:ImageLibrarian) extends JPanel{

  // size of the big viewport, in tiles
  val VIEW_WIDTH = 12;
  val VIEW_HEIGHT = 10;

  var transform = new AffineTransform();
  setDoubleBuffered(true);

  addComponentListener(new ComponentAdapter {
      override def componentResized(e:ComponentEvent):Unit = {
        val w1:Double = VIEW_WIDTH * TerrainLbxReader.TILE_WIDTH;
        val h1:Double = VIEW_HEIGHT * TerrainLbxReader.TILE_HEIGHT;

        val w2:Double = e.getComponent().getWidth();
        val h2:Double = e.getComponent().getHeight();

        transform.setToScale(w2/w1, h2/h1);
      }
  });

  var windowStartX = 0;
  var windowStartY = 0;
  var currentPlane = Plane.ARCANUS;
  
  override def paintComponent(g:Graphics):Unit = {
    val g2d = g.asInstanceOf[Graphics2D];

    val saveTransform = g2d.getTransform();
    g2d.transform(transform);
    
    for (j <- 0 until VIEW_HEIGHT) {
      for (i <- 0 until VIEW_WIDTH) {
        val t = overworld.get(currentPlane, i + windowStartX, j + windowStartY);
        val image = imageLibrarian.getTerrainTileImage(t);
        g.drawImage(
          image,
          i * TerrainLbxReader.TILE_WIDTH,
          j * TerrainLbxReader.TILE_HEIGHT,
          null);
      }
    }

    g2d.setTransform(saveTransform);
  }
}
