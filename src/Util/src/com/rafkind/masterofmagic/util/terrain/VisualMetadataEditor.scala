package com.rafkind.masterofmagic.util.terrain

import org.newdawn.slick.AppGameContainer
import org.newdawn.slick.state.StateBasedGame
import org.newdawn.slick.state.BasicGameState
import org.newdawn.slick.GameContainer
import org.newdawn.slick.Graphics
import org.newdawn.slick.ScalableGame

import com.rafkind.masterofmagic.ui.framework._;
import com.rafkind.masterofmagic.util._;
import com.rafkind.masterofmagic.system._;
import com.rafkind.masterofmagic.state._;

class TerrainMetadata {
  
}

object TileTable {
  val BACKGROUND_COLOR = ComponentProperty("background_color", 0);
}

class TileTable extends Component[TileTable] {  
  override def render(graphics:Graphics):TileTable = {
    graphics.setColor(
      Colors.colors(getInt(TileTable.BACKGROUND_COLOR)));
    graphics.fillRect(
      getInt(Component.LEFT),
      getInt(Component.TOP),
      getInt(Component.WIDTH),
      getInt(Component.HEIGHT));
    this;
  }
}

class PlaneSelectionState(val imageLibrarian:ImageLibrarian,
                          val metadata:TerrainMetadata)
  extends BasicGameState
     with Container[PlaneSelectionState] {

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

    val arcanusTerrainTiles = new TileTable();
    arcanusTerrainTiles.set(
      Component.WIDTH -> VisualMetadataEditor.WIDTH,
      Component.HEIGHT -> VisualMetadataEditor.HEIGHT/2);

    val myrrorTerrainTiles = new TileTable();
    myrrorTerrainTiles.set(
      TileTable.BACKGROUND_COLOR -> 133,
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
  override def mouseClicked(button:Int, x:Int, y:Int, clicks:Int):Unit = {
    currentScreen.notifyOf(
      Event.MOUSE_CLICKED.spawn(
        currentScreen,
        new MouseClickedEventPayload(button, x, y, count)));
  }
  override def mousePressed(button:Int, x:Int, y:Int):Unit = {
    //println("B: " + button + " x: " + x + " y: " + y);
  }
  override def mouseReleased(button:Int, x:Int, y:Int):Unit = {
    //println("B: " + button + " x: " + x + " y: " + y);
  }
  override def mouseMoved(oldx:Int, oldy:Int, newx:Int, newy:Int):Unit = {
    //println("B: " + button + " x: " + x + " y: " + y);
  }
  override def mouseDragged(oldx:Int, oldy:Int, newx:Int, newy:Int):Unit = {
    //println("B: " + button + " x: " + x + " y: " + y);
  }
}

class TerrainBorderState(val metadata:TerrainMetadata) extends BasicGameState {

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
    app.setDisplayMode(960, 600, false);
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
    addState(new PlaneSelectionState(imageLibrarian, metadata));
    addState(new TerrainBorderState(metadata));
  }
}