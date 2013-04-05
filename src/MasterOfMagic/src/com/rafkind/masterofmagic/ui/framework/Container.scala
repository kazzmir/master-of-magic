/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework
import scala.collection.mutable.HashSet;

import org.newdawn.slick._;

trait Container extends Component {
  var components = new HashSet[Component]();
  
  //var keyFocusedComponent:Option[Component[_]] = None;

  def add(component:Component) = {
    components += component;
    this
  }

  def remove(component:Component) = {
    components -= component;
    this
  }

  def render(graphics:Graphics) = {
    for (component <- components) {
      component.render(graphics);
    }
    this;
  }  
    
  override def notifyOf(event:Event) {
    /*// send key events to the focused component if applicable
    (event.descriptor, keyFocusedComponent) match {
      case (Event.KEY_PRESSED, Some(x)) => 
        x.notifyOf(event);        
      case _ =>
    }*/
    
    // if we still haven't done anything yet, send to other 
    // listeners too
    if (!event.consumed) {
      listeners.get(event.descriptor)
        .map(y =>
          y.foreach(
            z => if (!event.consumed) {
              z(event);
            }
          )
        );
    }
    
    // if we still haven't done anything yet, send to child components
    if (!event.consumed) {
      var cmps = event.payload match {
        case p:LocatedPayload => 
            components.filter( c => c.containsScreenPoint(p.x, p.y))
        case _ => components
      }
      
      cmps foreach { 
        c => 
          if (!event.consumed) { 
            c.notifyOf(event) 
          } 
        };              
    }
  }
  
    
  
  /*def notifyOf(whatHappened:ComponentEventDescriptor, x:Int, y:Int, eventObject:ComponentEvent) {
    // send to oursends
    listeners.get(whatHappened)
      .map(y =>
        y.foreach(
          z => if ((!eventObject.consumed)) {
            z(eventObject);
          }
        )
      );
    if (!eventObject.consumed) {
      components.map(c => {
          if ((!eventObject.consumed) && (c.containsScreenPoint(x, y))) {
            c.notifyOf(whatHappened, eventObject);
          }
        }
      );
    }
  }*/ 
}