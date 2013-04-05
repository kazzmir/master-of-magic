/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import com.rafkind.masterofmagic.state._;

object PackingContainer {
  val SPACING = new ComponentProperty("spacing", 0);
}
class PackingContainer extends Container {

 def add(component:Component, alignment:Alignment) = {
   super.add(component);
   alignment match {
     case Alignment.HORIZONTAL =>
       component.set(Component.TOP -> getInt(Component.TOP));
       component.set(Component.LEFT -> (getInt(Component.LEFT) + getInt(Component.WIDTH)));
       set(Component.WIDTH ->
           (component.getInt(Component.WIDTH) + getInt(Component.WIDTH) + getInt(PackingContainer.SPACING)));
     case Alignment.VERTICAL =>
       component.set(Component.LEFT -> getInt(Component.LEFT));
       component.set(Component.TOP -> (getInt(Component.TOP) + getInt(Component.HEIGHT)));
       set(Component.HEIGHT ->
           (component.getInt(Component.HEIGHT) + getInt(Component.HEIGHT) + getInt(PackingContainer.SPACING)));
     case _ =>
   }

   this
 }
}
