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

import javax.swing.SwingUtilities;
import java.awt.Toolkit;
/*import org.newdawn.slick.AppGameContainer;
import org.newdawn.slick.ScalableGame;
*/
import com.rafkind.masterofmagic.util._;


import com.rafkind.masterofmagic.ui.swing.MainFrame;

object Main {

  /**
   * @param args the command line arguments
   */
  /*def mainX(args: Array[String]): Unit = {
    var app = new AppGameContainer(new MasterOfMagic("Master of Magic"));
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(640, 400, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app.start();   
  }*/

  def main(args: Array[String]):Unit = {
    SwingUtilities.invokeLater(new Runnable {
      override def run():Unit = {
        val mainFrame = new MainFrame();
        val screenSize = Toolkit.getDefaultToolkit().getScreenSize();
        val WIDTH = 1024;
        val HEIGHT = 768;

        mainFrame.setBounds((screenSize.width - WIDTH)/2, (screenSize.height - HEIGHT)/2, WIDTH, HEIGHT);

        mainFrame.setVisible(true);
      }
    });
  }
}
