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

import scala.xml.XML._;
import scala.xml._;
import scala.collection.mutable.HashMap;
import scala.collection.mutable.HashSet;

import com.rafkind.masterofmagic.util._;
import com.rafkind.masterofmagic.system._;
import com.rafkind.masterofmagic.state._;

class Palette(metadataManager:MetadataManager, imageLibrarian:ImageLibrarian) extends JPanel with Scrollable {
  val COLUMNS = 4;
  val ROWS = TerrainLbxReader.TILE_COUNT / COLUMNS;

  val tileFont = new Font("Arial", Font.PLAIN, 9);
  var selectedTile:Int = 0;

  def getSelectedTile = selectedTile;
  def setSelectedTile(x:Int):Unit = {
    selectedTile = x;
  }

  var seen = new HashSet[Int];
  def haveSeen(i:Int) =
    seen += i;

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
    graphics.setFont(tileFont);
    val metrics = graphics.getFontMetrics();
    val height1 = metrics.getHeight();

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
          /*graphics.drawImage(imageLibrarian.getTerrainTileImage(t),
                             x+1,
                             y+1,
                             TerrainLbxReader.TILE_WIDTH * 2,
                             TerrainLbxReader.TILE_HEIGHT * 2,
                             null);*/
          graphics.setColor(Color.WHITE);
          graphics.drawString(t.toString(), x+1, y+height1);
          metadataManager.metadata.get(t) match {
            case Some(tm:EditableTerrainTileMetadata) =>
              graphics.drawString(tm.terrainType, x+1, y+height1*2);
            case _ =>
          }
          if (seen.contains(t)) {
            graphics.setColor(Color.YELLOW);
            graphics.fillRect(x + tw-12, y+1, 10, 10);
          }
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

class MetadataManager(path:String) {
  val metadata = new HashMap[Int, EditableTerrainTileMetadata]();

  try {
    load(path) \ "metadata" foreach { (m) =>
      val borders = new Array[Option[TerrainType]](CardinalDirection.values.length);
      m \ "borders" foreach { (b) =>
        borders(Integer.parseInt((b \ "@direction").text)) =
          Some(TerrainType.values(Integer.parseInt((b \ "@terrain").text)));
      }

      val id = Integer.parseInt((m \ "@id").text)
      val terrainType = Integer.parseInt((m \ "@terrainType").text);
      val plane = Integer.parseInt((m \ "@plane").text);
      metadata += id -> new EditableTerrainTileMetadata(id,
                                              TerrainType.values(terrainType),
                                              borders,
                                              Plane.values(plane), None);
    }
  } catch {
    case x => println(x);
  }  
}

class SandboxMap(metadataManager:MetadataManager, imageLibrarian:ImageLibrarian, palette:Palette) extends JPanel {
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

  def rememberSeen():Unit = {
    for (y <- 0 until TILES_DOWN) {
      for (x <- 0 until TILES_ACROSS) {
        palette.haveSeen(terrain(x+(y*TILES_ACROSS)));
      }
    }
    palette.repaint();
  }

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

  def okMatch(sourceTile:Int, destTile:Int, direction:CardinalDirection) = {    
    (metadataManager.metadata.get(sourceTile), metadataManager.metadata.get(destTile)) match {
      case (Some(source:EditableTerrainTileMetadata), Some(dest:EditableTerrainTileMetadata)) =>
        (source.terrainType,
         source.borderingTerrainTypes(direction.id),
         dest.terrainType,
         dest.borderingTerrainTypes(CardinalDirection.opposite(direction).id)) match {
          case (TerrainType.OCEAN, Some(TerrainType.OCEAN), TerrainType.OCEAN, _) => true
          case (TerrainType.OCEAN, Some(TerrainType.OCEAN), TerrainType.SHORE, _) => true
          case (TerrainType.SHORE, Some(TerrainType.OCEAN), TerrainType.OCEAN, _) => true
          case (TerrainType.SHORE, Some(TerrainType.OCEAN), TerrainType.SHORE, _) => true
          case (TerrainType.SHORE, Some(TerrainType.SHORE), TerrainType.SHORE, Some(TerrainType.SHORE)) => true
          // land adjacent to shore must be non-river, unless there is a river outlet
          case (TerrainType.SHORE, Some(TerrainType.RIVER), TerrainType.RIVER, Some(TerrainType.RIVER)) => true
          case (TerrainType.SHORE, Some(TerrainType.GRASSLAND), x:TerrainType, _)
            if ((x != TerrainType.RIVER) && x.isLand) => true

          case (TerrainType.RIVER, Some(TerrainType.RIVER), TerrainType.RIVER, Some(TerrainType.RIVER)) => true
          case (TerrainType.RIVER, Some(TerrainType.RIVER), TerrainType.SHORE, Some(TerrainType.RIVER)) => true
          case (TerrainType.TUNDRA, Some(TerrainType.TUNDRA), TerrainType.TUNDRA, Some(TerrainType.TUNDRA)) => true
          case (TerrainType.MOUNTAIN, Some(TerrainType.MOUNTAIN), TerrainType.MOUNTAIN, Some(TerrainType.MOUNTAIN)) => true
          case (TerrainType.HILLS, Some(TerrainType.HILLS), TerrainType.HILLS, Some(TerrainType.HILLS)) => true

          // any land can be adjacent to any other land
          case (y:TerrainType, Some(TerrainType.GRASSLAND), x:TerrainType, _)
            if (x.isLand) => true
          // shore can be adjacent to any land, on the land side
          case (TerrainType.SHORE, Some(y:TerrainType), TerrainType.SHORE, Some(z:TerrainType))
            if (y.isLand && z.isLand) => true
          case (y:TerrainType, _, TerrainType.SHORE, Some(z:TerrainType))
            if (y.isLand && z.isLand) => true
          case _ =>
            false
        }
      case _ =>
        false
    }
  }
  def drawArrow(graphics:Graphics, cornerX:Int, cornerY:Int, direction:CardinalDirection):Unit = {
    val halfX = TerrainLbxReader.TILE_WIDTH / 2;
    val halfY = TerrainLbxReader.TILE_HEIGHT / 2;
    val length = halfY;
    val dLength = scala.math.sqrt(direction.dx * direction.dx + direction.dy * direction.dy);

    val endX = scala.math.round(halfX + direction.dx * length / dLength).toInt;
    val endY = scala.math.round(halfY + direction.dy * length / dLength).toInt;

    graphics.drawLine(cornerX + halfX, cornerY + halfY, cornerX + endX, cornerY + endY);
  }
  
  override def paintComponent(graphics:Graphics):Unit = {
    val g2d = graphics.asInstanceOf[Graphics2D];
    val oldTransform = g2d.getTransform();

    g2d.transform(zoomTransform);

    graphics.setColor(Color.MAGENTA);

    for (y <- 0 until TILES_DOWN) {
      for (x <- 0 until TILES_ACROSS) {
        val sourceTileIndex = x+(y*TILES_ACROSS);
        val sourceTile = terrain(sourceTileIndex);
        val image:Image = null; // imageLibrarian.getTerrainTileImage(sourceTile);

        graphics.drawImage(image, x * TerrainLbxReader.TILE_WIDTH, y * TerrainLbxReader.TILE_HEIGHT, null);

        for (dir <- CardinalDirection.values) {
          val newX = x + dir.dx;
          val newY = y + dir.dy;
          if ((newX >= 0) && (newY >= 0) && (newX < TILES_ACROSS) && (newY < TILES_DOWN)) {
            val newTileIndex = newX+(newY*TILES_ACROSS);
            val newTile = terrain(newTileIndex);

            if (!okMatch(sourceTile, newTile, dir)) {
              drawArrow(graphics, x * TerrainLbxReader.TILE_WIDTH, y * TerrainLbxReader.TILE_HEIGHT, dir);
            }
          }
        }
      }
    }

    g2d.setTransform(oldTransform);
  }
}

object FancyMetadataEditor {
  def main(args: Array[String]): Unit = {
    /*var graphicsEnvironment = GraphicsEnvironment.getLocalGraphicsEnvironment();
    var graphicsDevice = graphicsEnvironment.getDefaultScreenDevice();
    var displayMode = graphicsDevice.getDisplayMode();

    var mm = new MetadataManager(Data.path("terrainMetaData.xml"));
    val library = new ImageLibrarian(Data.originalDataPath("TERRAIN.LBX"));    
    val pal = new Palette(mm, library);
    val map = new SandboxMap(mm, library, pal);
    val scrollPal = new JScrollPane(pal);


    var frame = new JFrame("Metadata Editor");
    frame.setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);
    frame.setLayout(new BorderLayout());
    frame.getContentPane().add(map, BorderLayout.CENTER);
    frame.getContentPane().add(scrollPal, BorderLayout.EAST);
    frame.addKeyListener(new KeyAdapter(){
        override def keyPressed(e:KeyEvent):Unit = {
          if (e.getKeyCode() == KeyEvent.VK_SPACE) {
            map.rememberSeen();
          }
        }
    });
    frame.pack();
    frame.setBounds((displayMode.getWidth() - 800)/2,
                    (displayMode.getHeight() - 600) / 2,
                    800,
                    600);


    frame.setVisible(true);*/
  }
}
