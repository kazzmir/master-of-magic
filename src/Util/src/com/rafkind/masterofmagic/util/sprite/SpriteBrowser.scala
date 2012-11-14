/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util.sprite

import com.rafkind.masterofmagic.system.Data
import com.rafkind.masterofmagic.util.LbxReader
import com.rafkind.masterofmagic.util.OriginalGameAsset
import java.awt.GraphicsEnvironment
import javax.swing.Box
import javax.swing.JFrame
import javax.swing.JList
import javax.swing.JScrollPane
import javax.swing.event.ListSelectionEvent
import javax.swing.event.ListSelectionListener

class SpriteBrowser {

}

object SpriteBrowser {
  def main(args:Array[String]):Unit = {
    var graphicsEnvironment = GraphicsEnvironment.getLocalGraphicsEnvironment();
    var graphicsDevice = graphicsEnvironment.getDefaultScreenDevice();
    var displayMode = graphicsDevice.getDisplayMode();

    var frame = new JFrame("Sprite Browser");
    frame.setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);
    val mainBox = Box.createHorizontalBox();
    val lbxes = new JList(OriginalGameAsset.values);      
    mainBox.add(new JScrollPane(lbxes));
    val subfiles = new JList(Array("A", "B", "C"));    
    mainBox.add(Box.createHorizontalStrut(5));
    mainBox.add(subfiles);
    val groupIndexes = new JList(Array("A", "B", "C"));
    mainBox.add(Box.createHorizontalStrut(5));
    mainBox.add(groupIndexes);
    val sprites = new JList(Array("A", "B", "C"));
    mainBox.add(Box.createHorizontalStrut(5));
    mainBox.add(sprites);
    
    lbxes.addListSelectionListener(new ListSelectionListener(){
      override def valueChanged(event:ListSelectionEvent):Unit = {
        if (!event.getValueIsAdjusting()) {
          val selected = lbxes.getSelectedValue();
          if (selected.fileName.endsWith(".LBX")) {
            /*val reader = new LbxReader(Data.)
            subfiles.setModel(x$1);*/
          }
        }
      }
    });
    
    frame.getContentPane().add(mainBox);
    frame.pack();
    frame.setBounds((displayMode.getWidth() - 800)/2,
                    (displayMode.getHeight() - 600) / 2,
                    800,
                    600);
    
    frame.setVisible(true);
  }
}
