/**
 * TODO:
 *
 * Game State
 * Game Logic
 * UI Screens
 * Map UI
 * Battle UI
 * Assets
 *
 */

package com.rafkind.masterofmagic

import java.util.logging.Level;
import java.util.logging.Logger;
import org.newdawn.slick.AppGameContainer;
import org.newdawn.slick.ScalableGame;

import com.rafkind.masterofmagic.util._;

object Main {

  /**
   * @param args the command line arguments
   */
  def main(args: Array[String]): Unit = {
    var app = new AppGameContainer(new MasterOfMagic("Master of Magic"));
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(640, 400, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app.start();
   //TerrainMetadataEditor.main(args);
  }

}
