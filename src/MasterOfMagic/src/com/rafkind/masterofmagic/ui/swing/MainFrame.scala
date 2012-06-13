/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.swing

import javax.swing._;
import java.awt._;

import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.system._;

class MainFrame extends JFrame("abc") {
  setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);

  val gameState = State.createGameState(4);
  
  val mapPanel = new MapPanel(
    gameState.overworld,
    new ImageLibrarian(Data.originalDataPath("TERRAIN.LBX")));
  getContentPane().setLayout(new BorderLayout());
  getContentPane().add(mapPanel, BorderLayout.CENTER);

  val menuPanel = new JPanel();
  menuPanel.setLayout(new GridLayout(1, 7))
  menuPanel.add(new JButton("Game"));
  menuPanel.add(new JButton("Spells"));
  menuPanel.add(new JButton("Armies"));
  menuPanel.add(new JButton("Cities"));
  menuPanel.add(new JButton("Magic"));
  menuPanel.add(new JButton("Info"));
  menuPanel.add(new JButton("Plane"));
  val sidePanel = new Box(BoxLayout.Y_AXIS);
  sidePanel.add(new JButton("Minimap"));
  sidePanel.add(new JButton("Mana/Money"));
  sidePanel.add(new JButton("etc"));

  getContentPane().add(menuPanel, BorderLayout.NORTH);
  getContentPane().add(sidePanel, BorderLayout.EAST);

}