/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.main

import com.rafkind.masterofmagic.ui.framework._;
import org.newdawn.slick._;
import com.rafkind.masterofmagic.util.TerrainLbxReader._;
import com.rafkind.masterofmagic.state._;
import com.google.inject._;

object MainMap {
  val TILESIZE_ACROSS = 12;
  val TILESIZE_DOWN = 10;
}

class MainMap @Inject() (terrainPainter:TerrainPainter) extends Component {
  set(Component.WIDTH -> MainMap.TILESIZE_ACROSS * TILE_WIDTH);
  set(Component.HEIGHT -> MainMap.TILESIZE_DOWN * TILE_HEIGHT);

  var overworld:Overworld = null;
  var plane = Plane.ARCANUS;
  var cx = 0;
  var cy = 0;
  
  listen(Event.KEY_PRESSED, (event:Event) => {
    event.payload.asInstanceOf[KeyPressedEventPayload].key match {
      case Input.KEY_UP => {
        moveMap(CardinalDirection.NORTH, 1);
        Some(this);
      }    
      case Input.KEY_DOWN => {
        moveMap(CardinalDirection.SOUTH, 1);
        Some(this);
      }    
      case Input.KEY_LEFT => {
        moveMap(CardinalDirection.WEST, 1);
        Some(this);
      }
      case Input.KEY_RIGHT => {
        moveMap(CardinalDirection.EAST, 1);
        Some(this);
      }
      case _ => 
        None;
    }
  });

  def setOverworld(overworld:Overworld):MainMap = {
    this.overworld = overworld;
    this
  }

  def moveMap(direction:CardinalDirection, distance:Int):Unit = {
    val dx = direction.dx * distance;
    val dy = direction.dy * distance;

    cx += dx;
    if (cx < 0) {
      cx += overworld.width;
    } else if (cx >= overworld.width) {
      cx -= overworld.width;
    }
    cy += dy;
    if (cy < 0) {
      cy = 0;
    } else if (cy >= (overworld.height - MainMap.TILESIZE_DOWN)) {
      cy = overworld.height - MainMap.TILESIZE_DOWN;
    }
  }

  override def render(graphics:Graphics) = {
    val top = getInt(Component.TOP);
    val left = getInt(Component.LEFT);
    
    terrainPainter.render(graphics,
                          left, top,
                          cx, cy,
                          MainMap.TILESIZE_ACROSS, MainMap.TILESIZE_DOWN,
                          plane,
                          overworld);

    this;
  }
}