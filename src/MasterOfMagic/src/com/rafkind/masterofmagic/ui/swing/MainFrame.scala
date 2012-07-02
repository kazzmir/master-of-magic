/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.swing

import javax.swing._;
import java.awt._;

import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.system._;

class OverworldPanel(gameState:State) extends JPanel {
  val mapPanel = new MapPanel(
    gameState.overworld,
    new ImageLibrarian(Data.originalDataPath("TERRAIN.LBX")));
  setLayout(new BorderLayout());
  add(mapPanel, BorderLayout.CENTER);

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

  val resourcesUnitsOrSurveyPanel = new JPanel();
  resourcesUnitsOrSurveyPanel.setLayout(new CardLayout());

  val resourcesPanel = new JPanel();
  resourcesPanel.add(new JLabel("Resources"));

  val unitsPanel = new JPanel();
  unitsPanel.add(new JLabel("Units"));

  val surveyPanel = new JPanel();
  surveyPanel.add(new JLabel("Survey"));

  resourcesUnitsOrSurveyPanel.add(resourcesPanel, "RESOURCES");
  resourcesUnitsOrSurveyPanel.add(unitsPanel, "UNITS");
  resourcesUnitsOrSurveyPanel.add(surveyPanel, "SURVEY");

  sidePanel.add(resourcesUnitsOrSurveyPanel);

  add(menuPanel, BorderLayout.NORTH);
  add(sidePanel, BorderLayout.EAST);
}

class CityPanel extends JPanel {
  setLayout(new GridLayout(2, 2));

  val upperLeft = new Box(BoxLayout.Y_AXIS);
  upperLeft.add(new JLabel("Town of Ass"));
  val ulTextLine1 = new JPanel();
  ulTextLine1.setLayout(new BorderLayout());
  ulTextLine1.add(new JLabel("High Elf"), BorderLayout.WEST);
  ulTextLine1.add(new JLabel("Population 5,000 (+100)"), BorderLayout.EAST);
  upperLeft.add(ulTextLine1);


  val upperRight = new JPanel();
  upperRight.setBackground(Color.RED);

  val lowerLeft = new JPanel();
  lowerLeft.setBackground(Color.GREEN);

  val lowerRight = new JPanel();
  lowerRight.setBackground(Color.WHITE);

  add(upperLeft);
  add(upperRight);
  add(lowerLeft);
  add(lowerRight);
}

class MainFrame extends JFrame("abc") {
  setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);

  val gameState = State.createGameState(4);

  val OVERWORLD_PANEL = "OVERWORLD";  
  val overworldPanel = new OverworldPanel(gameState);

  val CITY_PANEL = "CITY";
  val cityPanel = new CityPanel();
  
  getContentPane().setLayout(new CardLayout());
  //getContentPane().add(overworldPanel, OVERWORLD_PANEL);
  getContentPane().add(cityPanel, CITY_PANEL);
}