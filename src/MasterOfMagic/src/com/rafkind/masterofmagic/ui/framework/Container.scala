/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework
import scala.collection.mutable.HashSet;

import org.newdawn.slick._;

trait Container extends Component {
  var components = new HashSet[Component]();

  def add(component:Component):Container = {
    components += component;
    this
  }

  def remove(component:Component):Container = {
    components -= component;
    this
  }

  def render(graphics:Graphics):Unit = {
    for (component <- components) {
      component.render(graphics);
    }
  }
}