/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick._;
import com.google.common.base.Objects
import com.rafkind.masterofmagic.util._;

case class ComponentProperty(val name:String, val default:Any)

object Component {
  val LEFT = ComponentProperty("left", 0);
  val TOP = ComponentProperty("top", 0);
  val WIDTH = ComponentProperty("width", 0);
  val HEIGHT = ComponentProperty("height", 0);
  val BACKGROUND_IMAGE = ComponentProperty("background_image", null);
}

trait Component {
  var properties = new scala.collection.mutable.HashMap[ComponentProperty, Any]();

  def set(settings:Tuple2[ComponentProperty, Any]*):this.type = {
    settings.foreach( (x:Tuple2[ComponentProperty, Any]) => {
        properties += x;
        
        notifyOf(
          Event.PROPERTY_CHANGED.spawn(
            this, 
            new PropertyEventPayload(x)));          
      }
    );
    this
  }

  def getInt(key:ComponentProperty) = 
    properties.getOrElse(key, key.default).asInstanceOf[Int];

  def getImage(key:ComponentProperty) =
    properties.getOrElse(key, key.default).asInstanceOf[Image];

  var listeners = new CustomMultiMap[EventDescriptor, (Event => Option[Component])];
  
  def listen(toWhat:EventDescriptor, andThen:(Event => Option[Component])):this.type = {
    listeners.put(toWhat, andThen);
    this;
  }

  def notifyOf(event:Event) {
    listeners.get(event.descriptor)
      .map(y =>
        y.foreach(
          z => if (!event.consumed) {
            event.consumedBy(z(event));              
          }
        )
      );
  }
  
  def render(graphics:Graphics):this.type;

  def containsScreenPoint(x:Int, y:Int) = {
    val left = getInt(Component.LEFT);
    val width = getInt(Component.WIDTH);
    val top = getInt(Component.TOP);
    val height = getInt(Component.HEIGHT);

    val answer = ((x >= left) && (x < left + width) && (y >= top) && (y < top + height));
    //println("Checking " + x + ", " + y + " in " + "[" + left + ", " + top + "|" + width + ", " + height + "] " + this + " " + answer);
    answer;
  }
  
  override def toString() =
    Objects.toStringHelper(this).add("properties", properties).toString();
}