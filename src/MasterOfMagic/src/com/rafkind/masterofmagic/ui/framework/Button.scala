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

  listen(Event.PROPERTY_CHANGED, (event:Event) => {
      event.payload.asInstanceOf[PropertyEventPayload].whatChanged match {
      case (Button.UP_IMAGE, image:Image) =>
        set(Component.WIDTH -> scala.math.max(getInt(Component.WIDTH), image.getWidth()));
        set(Component.HEIGHT -> scala.math.max(getInt(Component.HEIGHT), image.getHeight()));
      case (Button.DOWN_IMAGE, image:Image) =>
        set(Component.WIDTH -> scala.math.max(getInt(Component.WIDTH), image.getWidth()));
        set(Component.HEIGHT -> scala.math.max(getInt(Component.HEIGHT), image.getHeight()));
      case _ =>
    }
  });

  listen(Event.MOUSE_CLICKED, (event:Event) => {
    val mouseEvent = event.payload.asInstanceOf[MouseClickedEventPayload];
    if (containsScreenPoint(mouseEvent.x, mouseEvent.y)) {
      println("clicked " + mouseEvent);
      event.consume();
    }  
  });

  listen(Event.MOUSE_PRESSED, (event:Event) => {
    val mouseEvent = event.payload.asInstanceOf[MouseEventPayload];
    if (containsScreenPoint(mouseEvent.x, mouseEvent.y)) {
      state = true;
      event.consume();
    }  
  });

  listen(Event.MOUSE_RELEASED, (event:Event) => {
    val mouseEvent = event.payload.asInstanceOf[MouseEventPayload];
    if (containsScreenPoint(mouseEvent.x, mouseEvent.y)) {
      state = false;
      event.consume();
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