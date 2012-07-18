/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util.terrain
import java.awt._;
import java.awt.geom._;
import java.awt.event._;
import javax.swing._;
import javax.swing.event._;

import com.rafkind.masterofmagic.util._;
import com.rafkind.masterofmagic.ui.swing._;
import com.rafkind.masterofmagic.system._;

class Palette(imageLibrarian:ImageLibrarian) extends JPanel with Scrollable {
  val COLUMNS = 4;
  val ROWS = TerrainLbxReader.TILE_COUNT / COLUMNS;

  var selectedTile:Int = 0;

  def getSelectedTile = selectedTile;
  def setSelectedTile(x:Int):Unit = {
    selectedTile = x;
  }

  addMouseListener(new MouseAdapter() {
      override def mouseClicked(e:MouseEvent):Unit = {
        val tw = TerrainLbxReader.TILE_WIDTH*2 + 2;
        val th = TerrainLbxReader.TILE_HEIGHT*2 + 2;
        val tx = e.getX() / tw;
        val ty = e.getY() / th;

        val t = tx + (ty * COLUMNS);
        e.getButton() match {
          case MouseEvent.BUTTON1 => 
            selectedTile = t;
            repaint();
          case MouseEvent.BUTTON3 =>
          case _ =>
        }
      }
  });

  override def getPreferredScrollableViewportSize()=
    new Dimension(COLUMNS * (TerrainLbxReader.TILE_WIDTH*2+2)+2 + 12, ROWS * (TerrainLbxReader.TILE_HEIGHT*2+2)+2);
  override def getPreferredSize() = getPreferredScrollableViewportSize();
  override def getMinimumSize() = getPreferredScrollableViewportSize();
  
  override def getScrollableBlockIncrement(visibleRect:Rectangle, orientation:Int, direction:Int) =
    orientation match {
      case SwingConstants.VERTICAL => TerrainLbxReader.TILE_HEIGHT*2 + 2;
      case SwingConstants.HORIZONTAL => TerrainLbxReader.TILE_WIDTH*2 + 2;
    }

  override def getScrollableTracksViewportHeight() = false;
  override def getScrollableTracksViewportWidth() = true;

  override def getScrollableUnitIncrement(visibleRect:Rectangle, orientation:Int, direction:Int) =
    getScrollableBlockIncrement(visibleRect, orientation, direction);
  
  override def paintComponent(graphics:Graphics):Unit = {
    super.paintComponent(graphics);

    val clip = graphics.getClipBounds();
    val tw = TerrainLbxReader.TILE_WIDTH*2 + 2;
    val th = TerrainLbxReader.TILE_HEIGHT*2 + 2;
    
    val x1 = clip.x - (clip.x % tw);
    val y1 = clip.y - (clip.y % th);

    
    var y = y1;
    while (y < clip.y + clip.height) {
      var x = x1;
      while (x < clip.x + clip.width) {
        val tx = x / tw;
        val ty = y / th;
        val t = tx + (ty * COLUMNS);
        if ((t >= 0) && (t < TerrainLbxReader.TILE_COUNT)) {
          graphics.drawImage(imageLibrarian.getTerrainTileImage(t),
                             x+1,
                             y+1,
                             TerrainLbxReader.TILE_WIDTH * 2,
                             TerrainLbxReader.TILE_HEIGHT * 2,
                             null);
          graphics.setColor(Color.WHITE);
          graphics.drawString(t.toString(), x, y+TerrainLbxReader.TILE_WIDTH);
          if (t == selectedTile) {
            graphics.setColor(Color.RED);
            graphics.drawRect(x, y, tw-1, th-1);
          }
        }

        x += tw;
      }
      y += th;
    }
  }
}

class SandboxMap(imageLibrarian:ImageLibrarian, palette:Palette) extends JPanel {
  val TILES_ACROSS = 11;
  val TILES_DOWN = 10;

  val terrain = new Array[Int](TILES_ACROSS * TILES_DOWN);
  val zoomTransform = new AffineTransform();
  val unzoomTransform = new AffineTransform();
  addComponentListener(new ComponentAdapter(){
      override def componentResized(e:ComponentEvent):Unit = {
        val w:Double = TILES_ACROSS * TerrainLbxReader.TILE_WIDTH;
        val h:Double = TILES_DOWN * TerrainLbxReader.TILE_HEIGHT;

        zoomTransform.setToScale(getWidth()/w, getHeight()/h);
        unzoomTransform.setTransform(zoomTransform.createInverse());
        
      }
  });

  def place(e:MouseEvent):Unit = {
    val point = unzoomTransform.transform(e.getPoint(), null);
    val x = scala.math.floor(point.getX() / TerrainLbxReader.TILE_WIDTH).toInt;
    val y = scala.math.floor(point.getY() / TerrainLbxReader.TILE_HEIGHT).toInt;
    val t = x + (y * TILES_ACROSS);

    if ((e.getButton() == MouseEvent.BUTTON1)
        || ((e.getModifiers() & InputEvent.BUTTON1_MASK) == InputEvent.BUTTON1_MASK)) {
      terrain(t) = palette.getSelectedTile
      repaint();
    } else if (e.getButton() == MouseEvent.BUTTON3) {
      palette.setSelectedTile(terrain(t));
      palette.repaint();
    }
    
  }

  addMouseListener(new MouseAdapter(){
      override def mousePressed(e:MouseEvent):Unit = {
        place(e);
      }
  });
  addMouseMotionListener(new MouseMotionAdapter() {
      override def mouseDragged(e:MouseEvent):Unit = {
        place(e);
      }
  });
  
  override def paintComponent(graphics:Graphics):Unit = {
    val g2d = graphics.asInstanceOf[Graphics2D];
    val oldTransform = g2d.getTransform();

    g2d.transform(zoomTransform);

    for (y <- 0 until TILES_DOWN) {
      for (x <- 0 until TILES_ACROSS) {
        val image = imageLibrarian.getTerrainTileImage(terrain(x+y*TILES_ACROSS));
        graphics.drawImage(image, x * TerrainLbxReader.TILE_WIDTH, y * TerrainLbxReader.TILE_HEIGHT, null);
      }
    }

    g2d.setTransform(oldTransform);
  }
}

object FancyMetadataEditor {
  def main(args: Array[String]): Unit = {
    var graphicsEnvironment = GraphicsEnvironment.getLocalGraphicsEnvironment();
    var graphicsDevice = graphicsEnvironment.getDefaultScreenDevice();
    var displayMode = graphicsDevice.getDisplayMode();

    val library = new ImageLibrarian(Data.originalDataPath("TERRAIN.LBX"));    
    val pal = new Palette(library);
    val map = new SandboxMap(library, pal);
    val scrollPal = new JScrollPane(pal);


    var frame = new JFrame("Metadata Editor");
    frame.setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);
    frame.setLayout(new BorderLayout());
    frame.getContentPane().add(map, BorderLayout.CENTER);
    frame.getContentPane().add(scrollPal, BorderLayout.EAST);
    frame.pack();
    frame.setBounds((displayMode.getWidth() - 800)/2,
                    (displayMode.getHeight() - 600) / 2,
                    800,
                    600);


    frame.setVisible(true);
  }
}