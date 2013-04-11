package com.rafkind.masterofmagic.util.terrain

import org.newdawn.slick.AppGameContainer;
import org.newdawn.slick.state.StateBasedGame;
import org.newdawn.slick.GameContainer;
import org.newdawn.slick.Graphics;
import org.newdawn.slick.ScalableGame;
import org.newdawn.slick.Image;

import com.rafkind.masterofmagic.ui.framework._;
import com.rafkind.masterofmagic.util._;
import com.rafkind.masterofmagic.system._;
import com.rafkind.masterofmagic.state._;

import scala.collection.immutable.TreeSet;

import com.google.common.base.Objects

class TerrainMetadata {
  
}

trait Rectangular {
  def x:Int;
  def y:Int;
  def width:Int;
  def height:Int;
  
  def contains(px:Int, py:Int) =
    (px >= x && py >= y && px < x+width && py < y + height);
  
  def contains(r:Rectangular) =
    (r.x < x + width) && (r.y < y + height) && (r.x + r.width > x) && (r.y + r.height > y);
  
  def contains(x1:Int, y1:Int, x2:Int, y2:Int) = {
    val answer = (x1 < x + width) && (y1 < y + height) && (x2 > x) && (y2 > y);
    //println("is [" + x1 + ", " + y1 + ", " + x2 + ", " + y2 + "] in [" + x + ", " + y + "|" + width + ", " + height + "] " + answer);
    answer;
  }
  
  override def toString() =
    Objects.toStringHelper(this).add("x", x).add("y", y).add("width", width).add("height", height).toString()
}

class Tile(var x:Int, var y:Int, var index:Int, var highlighted:Boolean, var lastDrawnX:Int, var lastDrawnY:Int) extends Rectangular with Ordered[Tile] {  
  def height = TerrainLbxReader.TILE_HEIGHT;
  def width = TerrainLbxReader.TILE_WIDTH;
  
  override def compare(other:Tile) = 
    index - other.index;
}

class Bin(var x:Int, var y:Int, var width:Int, var height:Int) extends Rectangular {
  private var tiles = List[Tile]();
  
  def clear() = {
    tiles = List[Tile]();  
  }  
  def add(t:Tile) = {
    tiles = t :: tiles;  
  }  
  def getTiles() = tiles;
}

object TileTable {
  val BACKGROUND_COLOR = ComponentProperty("background_color", 0);
  val LINE_COLOR = ComponentProperty("line_color", 0);
  
  val MARGIN = 1;
  
  val BIN_WIDTH = 120;
  val BIN_HEIGHT = 110;
  
}

class TileTable(baseTileImage:Image, initialTileCount:Int) extends Component {  
  private var bins = List[Bin]();
  private var tiles = TreeSet[Tile]();
  
  private var offsetX:Int = 0;
  private var offsetY:Int = 0;
  private var startX:Int = 0;
  private var startY:Int = 0;
  private var endX:Int = 0;
  private var endY:Int = 0;  
  
  private var draggingMouseButton:Int = -1;
  
  for (tileIndex <- 0 until initialTileCount) {
    var tile = new Tile(0, 0, tileIndex, false, 0, 0);
    tiles = tiles + tile;
  }
  
  layout();
  
  listen(Event.MOUSE_DRAGGED, (event:Event) => {
    val mouseEvent = event.payload.asInstanceOf[MouseMotionEventPayload];  
    //println("dragged " + mouseEvent);    
    endX = mouseEvent.x;
    endY = mouseEvent.y;   
    Some(this);      
  });

  listen(Event.MOUSE_PRESSED, (event:Event) => {
    val mouseEvent = event.payload.asInstanceOf[MouseEventPayload];
    //println("pressed " + mouseEvent);
    draggingMouseButton = mouseEvent.button;
    startX = mouseEvent.x;
    startY = mouseEvent.y;
    endX = startX;
    endY = startY;    
    
    if (draggingMouseButton == 0) {
      for (tile <- tiles) {
        tile.highlighted = false;
      }
    }
    
    Some(this);
    
  });

  listen(Event.MOUSE_RELEASED, (event:Event) => {
    val mouseEvent = event.payload.asInstanceOf[MouseEventPayload];
    //println("released " + mouseEvent);
    
    mouseEvent.button match {
        case 1 => {
          endX = mouseEvent.x;
          endY = mouseEvent.y;
          offsetX += startX - endX;
          offsetY += startY - endY;
          startX = endX;
          startY = endY;
        } 
        case 0 => {
          endX = mouseEvent.x;
          endY = mouseEvent.y;
          
          var x1 = scala.math.min(startX, endX) + offsetX;
          var x2 = scala.math.max(startX, endX) + offsetX;
          var y1 = scala.math.min(startY, endY) + offsetY;
          var y2 = scala.math.max(startY, endY) + offsetY;
          
          var fakeBin = new Bin(x1, y1, x2-x1, y2-y1);
          for (tile <- tiles) {
            if (fakeBin.contains(tile)) {
              tile.highlighted = true;
            }
          }
          
          startX = endX;
          startY = endY;
        }
        case _ => {
            
        }
    }
    draggingMouseButton = -1;
    
    Some(this);    
  });
  
  def layout() = {
    var minX = 0;
    var minY = 0;
    var maxX = 0;
    var maxY = 0;
    
    var row = 0;
    var column = 0;
    
    for (tile <- tiles) {            
      tile.x = TileTable.MARGIN + (column * (TerrainLbxReader.TILE_WIDTH + TileTable.MARGIN));
      tile.y = TileTable.MARGIN + (row * (TerrainLbxReader.TILE_HEIGHT + TileTable.MARGIN));
  
       
      minX = scala.math.min(minX, tile.x);
      minY = scala.math.min(minY, tile.y);
      maxX = scala.math.max(maxX, tile.x + tile.width);
      maxY = scala.math.max(maxY, tile.y + tile.height);
      
      column = column + 1;
      if (column >= TerrainLbxReader.SPRITE_SHEET_WIDTH) {
        column = 0;
        row = row + 1;
      }
    }
    
    var screenWidth = maxX - minX;
    var screenHeight = maxY - minY;
    var horizBinCount = screenWidth / TileTable.BIN_WIDTH;
    var vertBinCount = screenHeight / TileTable.BIN_HEIGHT;
    for (j <- 0 until vertBinCount) {
      for (i <- 0 until horizBinCount) {
        
        var b = new Bin(i * TileTable.BIN_WIDTH, 
                        j * TileTable.BIN_HEIGHT, 
                        TileTable.BIN_WIDTH, 
                        TileTable.BIN_HEIGHT);    
        
        bins = b :: bins;
      }
    }
    
    for (tile <- tiles) {
      for (bin <- bins) {
        if (bin.contains(tile)) {
          bin.add(tile);
        }
      }
    }
  }
  
  override def render(graphics:Graphics) = {    
    var oldClip = graphics.getClip();
    
    graphics.setColor(
      Colors.colors(getInt(TileTable.BACKGROUND_COLOR)));
    var left = getInt(Component.LEFT);
    var top = getInt(Component.TOP);
    var width = getInt(Component.WIDTH);
    var height = getInt(Component.HEIGHT);    
    
    graphics.setClip(left * VisualMetadataEditor.SCALE_FACTOR, 
                     top * VisualMetadataEditor.SCALE_FACTOR, 
                     width * VisualMetadataEditor.SCALE_FACTOR, 
                     height * VisualMetadataEditor.SCALE_FACTOR);
    
    graphics.fillRect(left, top, width, height); 
    
    var dragx = 0;
    var dragy = 0;
    
    if (draggingMouseButton == 1) {
      dragx = startX - endX;
      dragy = startY - endY;
    }
    
    var x2 = offsetX + width + dragx;
    var y2 = offsetY + height + dragy;

    graphics.setColor(
      Colors.colors(getInt(TileTable.LINE_COLOR)));
    
    baseTileImage.startUse();
    
    for (bin <- bins) {      
      if (bin.contains(offsetX + dragx, offsetY + dragy, x2, y2)) {
        //println("Bin " + bin);
        for (t <- bin.getTiles()) {
          val whichTile:Int = t.index;

          val tX = (whichTile % TerrainLbxReader.SPRITE_SHEET_WIDTH) * TerrainLbxReader.TILE_WIDTH;
          val tY = (whichTile / TerrainLbxReader.SPRITE_SHEET_WIDTH) * TerrainLbxReader.TILE_HEIGHT;
          val dX = t.x - (offsetX + dragx);
          val dY = t.y - (offsetY + dragy);

          baseTileImage.drawEmbedded(
            dX, dY, dX + TerrainLbxReader.TILE_WIDTH, dY + TerrainLbxReader.TILE_HEIGHT,
            tX, tY, tX + TerrainLbxReader.TILE_WIDTH, tY + TerrainLbxReader.TILE_HEIGHT
          );
          t.lastDrawnX = dX;
          t.lastDrawnY = dY;
        }
      }
    }
    
    baseTileImage.endUse();
    
    for (bin <- bins) {
      if (bin.contains(offsetX + dragx, offsetY + dragy, x2, y2)) {
        for (t <- bin.getTiles()) {
          if (t.highlighted) {
            graphics.drawRect(t.lastDrawnX-1, t.lastDrawnY-1, TerrainLbxReader.TILE_WIDTH+1, TerrainLbxReader.TILE_HEIGHT+1);
          }
        }
      }
    }
    graphics.setClip(oldClip);
    
    
    if (draggingMouseButton == 0) {
      graphics.drawRect(startX, startY, endX - startX, endY - startY);
    }
    
    this;
  }
}

class PlaneSelectionState(val imageLibrarian:ImageLibrarian,
                          val metadata:TerrainMetadata,
                          val baseTileImage:Image)
  extends InputManagerGameState
     with Container {
    
  topLevelContainer = this;

  override def getID() = 1;
  
  override def init(container:GameContainer, game:StateBasedGame):Unit = {

    val gameButton = new Button();
    gameButton.set(
      Button.UP_IMAGE ->
        imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 1, 0),
      Button.DOWN_IMAGE ->
        imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 1, 1),
      Component.LEFT -> 0,
      Component.TOP -> 0);

    val arcanusTerrainTiles = new TileTable(baseTileImage, TerrainLbxReader.TILE_COUNT);
    arcanusTerrainTiles.set(
      TileTable.LINE_COLOR -> 206,
      Component.WIDTH -> VisualMetadataEditor.WIDTH,
      Component.HEIGHT -> VisualMetadataEditor.HEIGHT/2);

    val myrrorTerrainTiles = new TileTable(baseTileImage, 0);
    myrrorTerrainTiles.set(
      TileTable.BACKGROUND_COLOR -> 133,
      TileTable.LINE_COLOR -> 206,
      Component.WIDTH -> VisualMetadataEditor.WIDTH,
      Component.HEIGHT -> VisualMetadataEditor.HEIGHT/2);

    val container = new PackingContainer();
    container   
      .add(arcanusTerrainTiles, Alignment.VERTICAL)
      .add(myrrorTerrainTiles, Alignment.VERTICAL);
    
    add(container);
    add(gameButton);
  }

  override def update(container:GameContainer, game:StateBasedGame, delta:Int):Unit = {
  }

  override def render(container:GameContainer, game:StateBasedGame, graphics:Graphics):Unit = {
    this.render(graphics);
  }
}

class TerrainBorderState(val metadata:TerrainMetadata) 
  extends InputManagerGameState
     with Container {
    
  topLevelContainer = this;

  override def getID() = 2;

  override def init(container:GameContainer, game:StateBasedGame):Unit = {
  }

  override def update(container:GameContainer, game:StateBasedGame, delta:Int):Unit = {
  }

  override def render(container:GameContainer, game:StateBasedGame, graphics:Graphics):Unit = {
  }
}

object VisualMetadataEditor {
  val WIDTH = 320;
  val HEIGHT = 200;
  
  val SCALE_FACTOR = 3;

  def main(args:Array[String]):Unit = {

    val fontManager =
      new FontManager(
        Data.originalDataPath(
          OriginalGameAsset.FONTS.fileName));
    
    val imageLibrarian = new ImageLibrarian(fontManager);
    val metadata = new TerrainMetadata();
    val game = new VisualMetadataEditor(imageLibrarian, metadata);
    
    val app = new AppGameContainer(
      new ScalableGame(game, WIDTH, HEIGHT));
    
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(WIDTH * SCALE_FACTOR, HEIGHT * SCALE_FACTOR, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app.start();  
  }
}

class VisualMetadataEditor(val imageLibrarian:ImageLibrarian,
                           val metadata:TerrainMetadata)
  extends StateBasedGame("Visual Metadata Editor") {
    
  

  override def initStatesList(container:GameContainer):Unit = {
    
    val baseTileImage = TerrainLbxReader.read(
        Data.originalDataPath(
          OriginalGameAsset.TERRAIN.fileName));
      
    addState(new PlaneSelectionState(imageLibrarian, metadata, baseTileImage));
    addState(new TerrainBorderState(metadata));
  }
}