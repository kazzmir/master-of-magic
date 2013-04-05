/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick.state.BasicGameState;

abstract class InputManagerGameState extends BasicGameState {
  var mTopLevelContainer:Container = null;  
  
  def topLevelContainer = mTopLevelContainer;
  def topLevelContainer_=(v:Container) = {
    mTopLevelContainer = v;
  }
  
  // if we are in "held" state, send events to this component
  var mTargetComponent:Option[Component] = None;
  var mButtonHeld:Int = -1;
  
  var mHoldUntil:(Event => Boolean) = null;
  
  def doEvent(event:Event) = { 
    //println(event.toString)
    //println(mTargetComponent.toString)
    
    mTargetComponent match {
      case Some(x:Component) => x.notifyOf(event);
      case _ =>
    }
    
    if (!event.consumed) {
      topLevelContainer.notifyOf(event);
    }  
    event;
  }
  
  override def mouseClicked(button:Int, x:Int, y:Int, clicks:Int):Unit = {    
    doEvent(
      Event.MOUSE_CLICKED.spawn(
        topLevelContainer,
        new MouseClickedEventPayload(button, x, y, clicks)));
  }
  override def mousePressed(button:Int, x:Int, y:Int):Unit = {
    val event = doEvent(
      Event.MOUSE_PRESSED.spawn(
        topLevelContainer,
        new MouseEventPayload(button, x, y)));
    
    if (event.consumed) {
      mTargetComponent = Some(event.component);
      mButtonHeld = button;
    }
  }
  override def mouseReleased(button:Int, x:Int, y:Int):Unit = {
    doEvent(
      Event.MOUSE_RELEASED.spawn(
        topLevelContainer,
        new MouseEventPayload(button, x, y)));
    
    mTargetComponent = None;
    mButtonHeld = -1;
  }
  override def mouseMoved(oldx:Int, oldy:Int, newx:Int, newy:Int):Unit = {    
    doEvent(
      Event.MOUSE_MOVED.spawn(
        topLevelContainer, 
        new MouseMotionEventPayload(mButtonHeld, oldx, oldy, newx, newy)));
  }
  override def mouseDragged(oldx:Int, oldy:Int, newx:Int, newy:Int):Unit = {    
    doEvent(
      Event.MOUSE_DRAGGED.spawn(
        topLevelContainer, 
        new MouseMotionEventPayload(mButtonHeld, oldx, oldy, newx, newy)));
  }
}
