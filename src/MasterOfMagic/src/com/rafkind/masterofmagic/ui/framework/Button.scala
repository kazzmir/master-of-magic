/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick._;
object Button {
  val UP_IMAGE = ComponentProperty("up_image", null);
  val DOWN_IMAGE = ComponentProperty("down_image", null);
}

class Button extends Component[Button] {

  listen(Component.PROPERTY_CHANGED, (event:ComponentEvent) => {
    event.asInstanceOf[PropertyChangedEvent].whatChanged match {
      case (Button.UP_IMAGE, image:Image) =>
        set(Component.WIDTH -> scala.math.max(getInt(Component.WIDTH), image.getWidth()));
        set(Component.HEIGHT -> scala.math.max(getInt(Component.HEIGHT), image.getHeight()));
      case (Button.DOWN_IMAGE, image:Image) =>
        set(Component.WIDTH -> scala.math.max(getInt(Component.WIDTH), image.getWidth()));
        set(Component.HEIGHT -> scala.math.max(getInt(Component.HEIGHT), image.getHeight()));
      case _ =>
    }
  });
  
  var state:Boolean = false;
  
  override def render(graphics:Graphics):Button = {
    if (state) {      
      getImage(Button.DOWN_IMAGE).draw(
        getInt(Component.LEFT),
        getInt(Component.TOP));
    } else {      
      getImage(Button.UP_IMAGE).draw(
        getInt(Component.LEFT),
        getInt(Component.TOP));
    }
    this;
  }
}