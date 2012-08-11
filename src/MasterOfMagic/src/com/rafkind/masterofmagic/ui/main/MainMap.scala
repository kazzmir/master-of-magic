/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.main

import com.rafkind.masterofmagic.ui.framework._;
import org.newdawn.slick._;
import com.rafkind.masterofmagic.util.TerrainLbxReader._;
import com.rafkind.masterofmagic.state._;

object MainMap {
  val TILESIZE_ACROSS = 12;
  val TILESIZE_DOWN = 10;
}

class MainMap(terrainPainter:TerrainPainter) extends Component[MainMap] {
  set(Component.WIDTH -> MainMap.TILESIZE_ACROSS * TILE_WIDTH);
  set(Component.HEIGHT -> MainMap.TILESIZE_DOWN * TILE_HEIGHT);

  var overworld:Overworld = null;
  
  def setOverworld(overworld:Overworld):MainMap = {
    this.overworld = overworld;
    this
  }

  override def render(graphics:Graphics):MainMap = {
    val top = getInt(Component.TOP);
    val left = getInt(Component.LEFT);
    
    terrainPainter.render(graphics,
                          left, top,
                          0, 0,
                          MainMap.TILESIZE_ACROSS, MainMap.TILESIZE_DOWN,
                          Plane.ARCANUS,
                          overworld);

    this;
  }
}
