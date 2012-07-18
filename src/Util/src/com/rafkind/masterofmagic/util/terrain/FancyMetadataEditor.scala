/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util.terrain
import java.awt._;
import javax.swing._;

object TileLibrary {
  val TILE_WIDTH = 20;
  val TILE_HEIGHT = 18;
}

class TileLibrary {
  
  def size = 2000;

  val im = createDummy();

  def getTile(index:Int) = im;
  
  def createDummy():Image = {
    var i = GraphicsEnvironment
      .getLocalGraphicsEnvironment()
      .getDefaultScreenDevice()
      .getDefaultConfiguration()
      .createCompatibleImage(TileLibrary.TILE_WIDTH, TileLibrary.TILE_HEIGHT);

    var g2d = i.createGraphics();
    g2d.setColor(Color.GREEN);
    g2d.fill(new Rectangle(0, 0, TileLibrary.TILE_WIDTH, TileLibrary.TILE_HEIGHT));

    return i;
  }
}


class Palette(tileLibrary:TileLibrary) extends JPanel with Scrollable {
  val COLUMNS = 4;
  val ROWS = tileLibrary.size / COLUMNS;

  override def getPreferredScrollableViewportSize()=
    new Dimension(COLUMNS * (TileLibrary.TILE_WIDTH*2+2)+2, ROWS * (TileLibrary.TILE_HEIGHT*2+2)+2);


  override def getScrollableBlockIncrement(visibleRect:Rectangle, orientation:Int, direction:Int) =
    orientation match {
      case SwingConstants.VERTICAL => TileLibrary.TILE_HEIGHT*2 + 2;
      case SwingConstants.HORIZONTAL => TileLibrary.TILE_WIDTH*2 + 2;
    }

  override def getScrollableTracksViewportHeight() = false;
  override def getScrollableTracksViewportWidth() = false;

  override def getScrollableUnitIncrement(visibleRect:Rectangle, orientation:Int, direction:Int) = 1;

  override def paintComponent(graphics:Graphics):Unit = {
    super.paintComponent(graphics);

    val clip = graphics.getClipBounds();
    println(clip);
    val tw = TileLibrary.TILE_WIDTH*2 + 2;
    val th = TileLibrary.TILE_HEIGHT*2 + 2;
    println(tw + ", " + th);
    val x1 = clip.x - (clip.x % tw);
    val y1 = clip.y - (clip.y % th);

    var x = x1;
    var y = y1;
    while (y < clip.y + clip.height) {
      while (x < clip.x + clip.width) {
        println(x + ", " + y);
        val tx = x / tw;
        val ty = y / th;
        val t = x + y * COLUMNS;
        println(t);
        if ((t >= 0) && (t < tileLibrary.size)) {
          graphics.drawImage(tileLibrary.getTile(t),
                             x+1,
                             y+1,
                             TileLibrary.TILE_WIDTH * 2,
                             TileLibrary.TILE_HEIGHT * 2,
                             null);
        }

        x += tw;
      }
      y += th;
    }
  }
}

class SandboxMap extends JPanel {
  
}

object FancyMetadataEditor {
  def main(args: Array[String]): Unit = {
    var graphicsEnvironment = GraphicsEnvironment.getLocalGraphicsEnvironment();
    var graphicsDevice = graphicsEnvironment.getDefaultScreenDevice();
    var displayMode = graphicsDevice.getDisplayMode();

    val library = new TileLibrary();
    val map = new SandboxMap();
    val pal = new Palette(library);
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