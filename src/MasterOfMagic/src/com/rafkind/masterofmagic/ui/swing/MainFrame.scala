/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.swing

import javax.swing.JFrame;

import com.rafkind.masterofmagic.state._;

class MainFrame extends JFrame("abc") {
  setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);

  val gameState = State.createGameState(4);
  
  val mapPanel = new MapPanel(gameState.overworld, new ImageLibrarian());
  getContentPane().add(mapPanel);

}