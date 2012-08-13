/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick._;
import org.newdawn.slick.state._;

class Screen extends Container[Screen] {

  var focusedComponent:Option[Component[_]] = None;

  override def render(graphics:Graphics):Screen = {
    getImage(Component.BACKGROUND_IMAGE).draw(0, 0);

    super.render(graphics);
    this;
  }

  def notifyOf(whatHappened:ComponentEventDescriptor, x:Int, y:Int, eventObject:ComponentEvent) {
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
          if ((!eventObject.consumed) /* && (c.containsScreenPoint(x, y))*/ ) {
            c.notifyOf(whatHappened, eventObject);
          }
        }
      );
    }
  }
}