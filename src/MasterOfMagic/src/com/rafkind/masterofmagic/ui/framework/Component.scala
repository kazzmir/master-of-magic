/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick._;
import com.rafkind.masterofmagic.util._;

case class ComponentProperty(val name:String, val default:Any)
case class ComponentEventDescriptor(val name:String)

object Component {
  val LEFT = ComponentProperty("left", 0);
  val TOP = ComponentProperty("top", 0);
  val WIDTH = ComponentProperty("width", 0);
  val HEIGHT = ComponentProperty("height", 0);
  val BACKGROUND_IMAGE = ComponentProperty("background_image", null);

  val PROPERTY_CHANGED = ComponentEventDescriptor("property_changed");
  val MOUSE_CLICKED = ComponentEventDescriptor("mouse_clicked");
  val KEY_PRESSED = ComponentEventDescriptor("key_pressed");
}

case class ComponentEvent(val component:Component[_], var consumed:Boolean)

case class PropertyChangedEvent(
  override val component:Component[_],
  val whatChanged:Tuple2[ComponentProperty, Any])
    extends ComponentEvent(component, false)

case class MouseClickedEvent(
  override val component:Component[_],
  val button:Int,
  val x:Int,
  val y:Int,
  val clickCount:Int)
    extends ComponentEvent(component, false)

case class KeyPressedEvent(
  override val component:Component[_],
  val key:Int,
  val ch:Char)
    extends ComponentEvent(component, false)


trait Component[T] {
  var properties = new scala.collection.mutable.HashMap[ComponentProperty, Any]();

  def set(settings:Tuple2[ComponentProperty, Any]*):T = {
    settings.foreach( (x:Tuple2[ComponentProperty, Any]) => {
        properties += x;
        
        notifyOf(Component.PROPERTY_CHANGED, new PropertyChangedEvent(this, x));
      }
    );
    this.asInstanceOf[T]
  }

  def getInt(key:ComponentProperty) = 
    properties.getOrElse(key, key.default).asInstanceOf[Int];

  def getImage(key:ComponentProperty) =
    properties.getOrElse(key, key.default).asInstanceOf[Image];

  var listeners = new CustomMultiMap[ComponentEventDescriptor, ComponentEvent => Unit];
  
  def listen(toWhat:ComponentEventDescriptor, andThen:(ComponentEvent => Unit)):T = {
    listeners.put(toWhat, andThen);
    this.asInstanceOf[T]
  }

  def notifyOf(whatHappened:ComponentEventDescriptor, eventObject:ComponentEvent) {
    listeners.get(whatHappened)
      .map(y =>
        y.foreach(
          z => if (!eventObject.consumed) {
            z(eventObject);
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