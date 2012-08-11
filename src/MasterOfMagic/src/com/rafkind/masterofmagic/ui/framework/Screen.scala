/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick._;
import org.newdawn.slick.state._;

class Screen extends Container[Screen] {

  override def render(graphics:Graphics):Screen = {
    getImage(Component.BACKGROUND_IMAGE).draw(0, 0);

    super.render(graphics);
    this;
  }
}