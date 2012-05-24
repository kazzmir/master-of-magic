/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.swing

import javax.swing.JFrame;

class MainFrame extends JFrame("abc") {
  setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);
  getContentPane().add(new MapPanel());

}