/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util.sprite

import com.rafkind.masterofmagic.system.Data
import com.rafkind.masterofmagic.util.LbxReader
import com.rafkind.masterofmagic.util.OriginalGameAsset
import com.rafkind.masterofmagic.util.SpriteReader
import java.awt.GraphicsEnvironment
import javax.swing.AbstractListModel
import javax.swing.Box
import javax.swing.ImageIcon
import javax.swing.JFrame
import javax.swing.JList
import javax.swing.JScrollPane
import javax.swing.JLabel
import javax.swing.event.ListSelectionEvent
import javax.swing.event.ListSelectionListener
import javax.swing.ListCellRenderer;
import java.awt.Image;

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
    val lbxModel = new AbstractListModel() {
      override def getElementAt(index:Int) =
        OriginalGameAsset.values(index);
      override def getSize() =
        OriginalGameAsset.values.size;
    };
    val lbxes = new JList(lbxModel);      
    mainBox.add(new JScrollPane(lbxes));
    val subFileModel = 
      new AbstractListModel() {
      var currentSize:Int = 0;
      
      val self = this;
      
      lbxes.addListSelectionListener(new ListSelectionListener(){
        override def valueChanged(event:ListSelectionEvent):Unit = {
          var newSize:Int = currentSize;
          
          if (!event.getValueIsAdjusting()) {
            val selected = lbxes.getSelectedValue().asInstanceOf[OriginalGameAsset];
            if (selected.fileName.endsWith(".LBX")) {
              val reader = new LbxReader(Data.originalDataPath(selected.fileName));
              val metaData = reader.metaData;
              newSize = metaData.subfileCount
            } else {
              newSize = 0;
            }
            if (newSize != currentSize) {
              if (newSize > currentSize) {
                fireIntervalAdded(self, currentSize, newSize);
              } else {
                fireIntervalRemoved(self, newSize, currentSize);
              }
              currentSize = newSize;
            } else {
              fireContentsChanged(self, 0, newSize);
            }
          }
        }
      });      
      override def getElementAt(index:Int) = new Integer(index);
      override def getSize() = currentSize;
    }
    val subfiles = new JList(subFileModel);    
    mainBox.add(Box.createHorizontalStrut(5));
    mainBox.add(new JScrollPane(subfiles));
    
    
    val groupIndexModel = new AbstractListModel(){
      var currentSize:Int = 0;
      
      val self = this;
      
      subfiles.addListSelectionListener(new ListSelectionListener() {
        override def valueChanged(event:ListSelectionEvent):Unit = {
          var newSize:Int = currentSize;
          
          if (!event.getValueIsAdjusting()) {
            val selected = lbxes.getSelectedValue().asInstanceOf[OriginalGameAsset];
            if (selected.fileName.endsWith(".LBX")) {              
              var subSelected = subfiles.getSelectedIndex();
              if (subSelected >= 0) {
                val reader = new LbxReader(Data.originalDataPath(selected.fileName));
                val header = SpriteReader.readHeader(reader, subfiles.getSelectedIndex());
                newSize = header.bitmapCount;
              } else {
                newSize = 0;
              }
            } else {
              newSize = 0;
            }
            if (newSize != currentSize) {
              if (newSize > currentSize) {
                fireIntervalAdded(self, currentSize, newSize);
              } else {
                fireIntervalRemoved(self, newSize, currentSize);
              }
              currentSize = newSize;
            } else {
              fireContentsChanged(self, 0, newSize);
            } 
          }
        }
      });      
      override def getElementAt(index:Int) = new Integer(index);
      override def getSize() = currentSize;
    };
    
    val groupIndexes = new JList(groupIndexModel);

    val spriteLabel = new JLabel();
    val librarian = new AwtImageLibrarian();
    
    val renderer = new JLabel() with ListCellRenderer {      
      override def getListCellRendererComponent(list:JList, value:Any, index:Int, isSelected:Boolean, cellHasFocus:Boolean):java.awt.Component = {
        //println(list, value, index, isSelected, cellHasFocus);
        val image = librarian.getRawSprite(
        lbxes.getSelectedValue().asInstanceOf[OriginalGameAsset],
        subfiles.getSelectedValue().asInstanceOf[Integer],
        index);
        this.setIcon(new ImageIcon(image));

        return this;
      }
    };
    groupIndexes.setCellRenderer(renderer);
    
    mainBox.add(Box.createHorizontalStrut(5));
    mainBox.add(new JScrollPane(groupIndexes));
    
    groupIndexes.addListSelectionListener(new ListSelectionListener() {
        override def valueChanged(event:ListSelectionEvent):Unit = {
          if (lbxes.getSelectedIndex() >= 0 
              && subfiles.getSelectedIndex() >= 0
              && groupIndexes.getSelectedIndex() >= 0) {
            val image = librarian.getRawSprite(
              lbxes.getSelectedValue().asInstanceOf[OriginalGameAsset],
              subfiles.getSelectedValue().asInstanceOf[Integer],
              groupIndexes.getSelectedValue().asInstanceOf[Integer]);
            spriteLabel.setIcon(new ImageIcon(image));
          }    
        }
    });
    
    mainBox.add(Box.createHorizontalStrut(5));
    mainBox.add(new JScrollPane(spriteLabel));
    
    lbxes.addListSelectionListener(new ListSelectionListener(){
        override def valueChanged(event:ListSelectionEvent):Unit = {
          subfiles.clearSelection();
          groupIndexes.clearSelection();
          spriteLabel.setIcon(null);
        }
    });
  
    subfiles.addListSelectionListener(new ListSelectionListener(){
        override def valueChanged(event:ListSelectionEvent):Unit = {
          groupIndexes.clearSelection();
          spriteLabel.setIcon(null);
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
