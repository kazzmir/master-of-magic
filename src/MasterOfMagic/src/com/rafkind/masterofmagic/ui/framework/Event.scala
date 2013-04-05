/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import com.google.common.base.Objects

case class EventDescriptor(val name:String) {
  def spawn(component:Component, payload:Any) = 
    new Event(this, component, false, payload);
}

class Event(val descriptor:EventDescriptor, 
            var component:Component, 
            var consumed:Boolean,
            val payload:Any) {

  def consumedBy(who:Option[Component]) = 
    who match {
      case Some(x:Component) => 
        consumed = true;
        component = x;
      case _ =>
        consumed = false;
    }
  
  override def toString() =
    Objects.toStringHelper(this)
      .add("descriptor", descriptor)
      .add("component", component)
      .add("consumed", consumed)
      .add("payload", payload).toString();
}

case class PropertyEventPayload(  
  val whatChanged:Tuple2[ComponentProperty, Any])

abstract class LocatedPayload {
  def x:Int;
  def y:Int;
}
    
case class MouseClickedEventPayload(
  val button:Int,
  val x:Int,
  val y:Int,
  val clickCount:Int) extends LocatedPayload

case class MouseEventPayload(
  val button:Int,
  val x:Int,
  val y:Int) extends LocatedPayload

case class MouseMotionEventPayload(
  val button:Int,
  val oldX:Int,
  val oldY:Int,
  val x:Int,
  val y:Int) extends LocatedPayload

case class KeyPressedEventPayload(
  val key:Int,
  val ch:Char)
    
object Event {  
  val PROPERTY_CHANGED = EventDescriptor("property_changed");
  val MOUSE_CLICKED = EventDescriptor("mouse_clicked");
  val MOUSE_PRESSED = EventDescriptor("mouse_pressed");
  val MOUSE_RELEASED = EventDescriptor("mouse_released");
  val MOUSE_MOVED = EventDescriptor("mouse_moved");
  val MOUSE_DRAGGED = EventDescriptor("mouse_dragged");
  val KEY_PRESSED = EventDescriptor("key_pressed");
}
