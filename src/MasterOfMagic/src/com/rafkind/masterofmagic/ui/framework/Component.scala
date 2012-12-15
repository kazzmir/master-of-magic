/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick._;
import com.rafkind.masterofmagic.util._;

case class ComponentProperty(val name:String, val default:Any)

object Component {
  val LEFT = ComponentProperty("left", 0);
  val TOP = ComponentProperty("top", 0);
  val WIDTH = ComponentProperty("width", 0);
  val HEIGHT = ComponentProperty("height", 0);
  val BACKGROUND_IMAGE = ComponentProperty("background_image", null);
}

trait Component[T] {
  var properties = new scala.collection.mutable.HashMap[ComponentProperty, Any]();

  def set(settings:Tuple2[ComponentProperty, Any]*):T = {
    settings.foreach( (x:Tuple2[ComponentProperty, Any]) => {
        properties += x;
        
        notifyOf(
          Event.PROPERTY_CHANGED.spawn(
            this, 
            new PropertyEventPayload(x)));          
      }
    );
    this.asInstanceOf[T]
  }

  def getInt(key:ComponentProperty) = 
    properties.getOrElse(key, key.default).asInstanceOf[Int];

  def getImage(key:ComponentProperty) =
    properties.getOrElse(key, key.default).asInstanceOf[Image];

  var listeners = new CustomMultiMap[EventDescriptor, (Event => Unit)];
  
  def listen(toWhat:EventDescriptor, andThen:(Event => Unit)):T = {
    listeners.put(toWhat, andThen);
    this.asInstanceOf[T]
  }

  def notifyOf(event:Event) {
    listeners.get(event.descriptor)
      .map(y =>
        y.foreach(
          z => if (!event.consumed) {
            z(event);
          }
        )
      );
  }
  
  def render(graphics:Graphics):T;

  def containsScreenPoint(x:Int, y:Int) = {
    val left = getInt(Component.LEFT);
    val width = getInt(Component.WIDTH);
    val top = getInt(Component.TOP);
    val height = getInt(Component.HEIGHT);

    ((x >= left) && (x < left + width) && (y >= top) && (y < top + height));
  }
}