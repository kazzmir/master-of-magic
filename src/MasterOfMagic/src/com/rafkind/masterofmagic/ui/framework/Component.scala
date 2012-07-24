/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import org.newdawn.slick._;

trait Component {
  
  var left:Int = 0;
  var top:Int = 0;
  var width:Int = 0;
  var height:Int = 0;
  
  def render(graphics:Graphics):Unit
}