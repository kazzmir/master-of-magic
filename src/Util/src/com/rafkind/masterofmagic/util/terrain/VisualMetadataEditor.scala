package com.rafkind.masterofmagic.util.terrain

import org.newdawn.slick.AppGameContainer
import org.newdawn.slick.BasicGame
import org.newdawn.slick.GameContainer
import org.newdawn.slick.Graphics
import org.newdawn.slick.ScalableGame

object VisualMetadataEditor {
  def main(args:Array[String]):Unit = {
    
    val game = new VisualMetadataEditor();
    
    val app = new AppGameContainer(
      new ScalableGame(game, 320, 200));
    
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(960, 600, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app.start();  
  }
}

class VisualMetadataEditor extends BasicGame("Visual Metadata Editor") {  
  override def init(container:GameContainer):Unit = {    
  }
  
  override def update(container:GameContainer, delta:Int):Unit = {    
  }
  
  override def render(container:GameContainer, graphics:Graphics):Unit = {    
  }
}