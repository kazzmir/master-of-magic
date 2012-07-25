/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework
import scala.collection.mutable.HashSet;

import org.newdawn.slick._;

trait Container[T] extends Component[T] {
  var components = new HashSet[Component[_]]();

  def add(component:Component[_]):T = {
    components += component;
    this.asInstanceOf[T]
  }

  def remove(component:Component[_]):T = {
    components -= component;
    this.asInstanceOf[T]
  }

  def render(graphics:Graphics):T = {
    for (component <- components) {
      component.render(graphics);
    }
    this.asInstanceOf[T];
  }
}