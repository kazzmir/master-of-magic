/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.swing

import javax.swing.JFrame;

import com.rafkind.masterofmagic.state._;

class MainFrame extends JFrame("abc") {
  setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);

  val overworld = Overworld.create();
  
  val mapPanel = new MapPanel(overworld, new ImageLibrarian());
  getContentPane().add(mapPanel);

}