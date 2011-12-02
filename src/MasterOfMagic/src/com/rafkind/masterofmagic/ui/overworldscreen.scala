/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui

import org.newdawn.slick._;
import org.newdawn.slick.state._;

import de.lessvoid.nifty._;
import de.lessvoid.nifty.slick._;
import de.lessvoid.nifty.screen._;

class OverworldMapScreenController extends ScreenController {
  override def onEndScreen():Unit = {}
  override def onStartScreen():Unit = {}
  override def bind(nifty:Nifty, screen:Screen):Unit = {}
}

class OverworldMapState extends NiftyGameState(0) {

  override def init(container:GameContainer, game:StateBasedGame):Unit = {
    super.init(container, game);

    fromXml("com/rafkind/masterofmagic/ui/overworld-screen.xml");
  }
}