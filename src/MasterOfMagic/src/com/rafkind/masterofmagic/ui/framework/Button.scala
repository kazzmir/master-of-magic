/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick._;

class Button(upImage:Image, downImage:Image) extends Component {
  var state:Boolean = false;
  
  override def render(graphics:Graphics):Unit = {
    if (state) {
      downImage.draw(left, top);
    } else {
      upImage.draw(left, top);
    }
  }
}
