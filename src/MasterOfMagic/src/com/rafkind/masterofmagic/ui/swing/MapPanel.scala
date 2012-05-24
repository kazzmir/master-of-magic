/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.swing

import javax.swing.JPanel;
import java.awt.Color;
import java.awt.Graphics;

class MapPanel extends JPanel {
  setDoubleBuffered(true);

  override def paintComponent(g:Graphics):Unit = {
    g.setColor(Color.BLUE);

    g.fillRect(0, 0, getWidth(), getHeight());
  }
}
