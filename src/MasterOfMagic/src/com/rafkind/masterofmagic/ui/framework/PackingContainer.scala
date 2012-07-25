/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import com.rafkind.masterofmagic.state._;

class PackingContainer extends Container[PackingContainer] {

 def add(component:Component[_], alignment:Alignment):PackingContainer = {
   super.add(component);

   alignment match {
     case Alignment.HORIZONTAL =>
       component.set(Component.TOP -> getInt(Component.TOP));
       component.set(Component.LEFT -> (getInt(Component.LEFT) + getInt(Component.WIDTH)));
       set(Component.WIDTH ->
           (component.getInt(Component.WIDTH) + getInt(Component.WIDTH)));
     case Alignment.VERTICAL =>
       component.set(Component.LEFT -> getInt(Component.LEFT));
       component.set(Component.TOP -> (getInt(Component.TOP) + getInt(Component.HEIGHT)));
       set(Component.HEIGHT ->
           (component.getInt(Component.HEIGHT) + getInt(Component.HEIGHT)));
     case _ =>
   }

   this
 }
}
