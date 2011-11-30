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


object Main {

  /**
   * @param args the command line arguments
   */
  def main(args: Array[String]): Unit = {
    var app = new AppGameContainer(new MasterOfMagic("Master of Magic"));
    app.setDisplayMode(640, 480, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(true);
    app.start();
  }

}
