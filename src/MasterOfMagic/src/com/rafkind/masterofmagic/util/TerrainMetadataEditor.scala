/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util

import org.newdawn.slick._;

import com.rafkind.masterofmagic.system._;
import com.rafkind.masterofmagic.state._;
import scala.xml.XML._;
import scala.xml._;
import scala.collection.mutable.HashMap;

// mirrors TerrainTileMetadata in State.scala
class EditableTerrainTileMetadata(
  var id:Int,
  var terrainType:TerrainType,
  var borderingTerrainTypes:Array[Option[TerrainType]],
  var plane:Plane,
  var parentId:Option[TerrainTileMetadata]) {

  def toNode() =
    <metadata 
      id={id.toString}
      terrainType={terrainType.id.toString}
      plane={plane.id.toString}>
      {(borderingTerrainTypes zip CardinalDirection.values) map {
        case (b, d) =>
          b match {
            case Some(x) => <borders direction={d.id.toString} terrain={x.id.toString} />
            case _ =>
          }
      }
    }
    </metadata>
}

class TerrainMetadataEditor(title:String) extends BasicGame(title) {
  import com.rafkind.masterofmagic.util.TerrainLbxReader._;
  
  var terrainTileSheet:Image = null;

  var currentTile:Int = 0;

  var currentDirection:CardinalDirection = CardinalDirection.CENTER;

  val BIG_WIDTH = 133;
  val BIG_HEIGHT = BIG_WIDTH * TILE_HEIGHT / TILE_WIDTH;

  val uiColor = Color.white;

  var metadata = new Array[EditableTerrainTileMetadata](TILE_COUNT);

  load(Data.path("terrainMetaData.xml")) \ "metadata" foreach { (m) =>
    var borders = new Array[Option[TerrainType]](CardinalDirection.values.length);
    m \ "borders" foreach { (b) =>
      borders(Integer.parseInt((b \ "@direction").text)) =
        Some(TerrainType.values(Integer.parseInt((b \ "@terrain").text)));
    }

    val id = Integer.parseInt((m \ "@id").text)
    val terrainType = Integer.parseInt((m \ "@terrainType").text);
    val plane = Integer.parseInt((m \ "@plane").text);
    metadata(id) = new EditableTerrainTileMetadata(id,
                                            TerrainType.values(terrainType),
                                            borders,
                                            Plane.values(plane), None);
  }

  class TerrainGuess(val terrainGuess:TerrainType, val borders: Array[Option[TerrainType]]);
  

  def guessTerrain(whichTile:Int, plane:Plane):TerrainGuess = {
    var terrainGuess = TerrainType.OCEAN;
    var borders = new Array[Option[TerrainType]](CardinalDirection.values.length);
    for (j <- 0 until borders.length) {
      borders(j) = None;
    }
    val votes = new HashMap[TerrainType, Int];

    getColorSwatchFromTile(whichTile,
                           CardinalDirection.CENTER) map {(c) =>
      val voting =
        (c.getRed(), c.getGreen(), c.getBlue()) match {
          case (0, 0, 88) => List(TerrainType.OCEAN, TerrainType.RIVER);
          case (16, 16, 92) => List(TerrainType.OCEAN, TerrainType.SWAMP, TerrainType.RIVER);
          case (76, 116, 36) => List(TerrainType.GRASSLAND, TerrainType.NATURE_NODE, TerrainType.SWAMP, TerrainType.RIVER, TerrainType.HILLS);
          case (60, 92, 60) => List(TerrainType.GRASSLAND, TerrainType.NATURE_NODE, TerrainType.FOREST, TerrainType.RIVER, TerrainType.MOUNTAIN, TerrainType.HILLS);
          case (56, 92, 16) => List(TerrainType.GRASSLAND, TerrainType.FOREST, TerrainType.NATURE_NODE, TerrainType.SWAMP, TerrainType.HILLS);
          case (140, 132, 128) => List(TerrainType.GRASSLAND, TerrainType.TUNDRA, TerrainType.MOUNTAIN);
          case (36, 116, 36) => List(TerrainType.FOREST, TerrainType.GRASSLAND, TerrainType.HILLS);
          case (4, 68, 4) => List(TerrainType.FOREST, TerrainType.NATURE_NODE, TerrainType.RIVER);
          case (16, 92, 16) => List(TerrainType.FOREST);
          case (172, 164, 160) => List(TerrainType.MOUNTAIN, TerrainType.VOLCANO, TerrainType.TUNDRA);
          case (240, 232, 228) => List(TerrainType.MOUNTAIN);
          case (156, 148, 144) => List(TerrainType.MOUNTAIN, TerrainType.GRASSLAND, TerrainType.TUNDRA);
          case (124, 116, 112) => List(TerrainType.MOUNTAIN, TerrainType.TUNDRA);
          case (88, 80, 76) => List(TerrainType.MOUNTAIN);
          case (192, 184, 180) => List(TerrainType.MOUNTAIN);
          case (72, 64, 60) => List(TerrainType.MOUNTAIN);
          case (56, 48, 44) => List(TerrainType.MOUNTAIN);
          case (52, 92, 16) => List(TerrainType.MOUNTAIN);
          case (224, 216, 212) => List(TerrainType.MOUNTAIN);
          case (188, 152, 116) => List(TerrainType.DESERT);
          case (212, 180, 152) => List(TerrainType.DESERT);
          case (16, 56, 92) => List(TerrainType.SWAMP);
          case (4, 36, 68) => List(TerrainType.SWAMP);
          case (92, 60, 60) => List(TerrainType.SWAMP);
          case (92, 120, 92) => List(TerrainType.TUNDRA);
          case (76, 104, 76) => List(TerrainType.TUNDRA);
          case (0, 0, 188) => List(TerrainType.SORCERY_NODE);
          case (0, 0, 252) => List(TerrainType.SORCERY_NODE);
          case (152, 0, 0) => List(TerrainType.CHAOS_NODE, TerrainType.VOLCANO);
          case (172, 0, 0) => List(TerrainType.CHAOS_NODE);
          case (92, 52, 16) => List(TerrainType.CHAOS_NODE);
          case (36, 68, 4) => List(TerrainType.GRASSLAND, TerrainType.RIVER, TerrainType.MOUNTAIN, TerrainType.FOREST, TerrainType.HILLS);
          case (96, 140, 56) => List(TerrainType.HILLS);
          case (44, 72, 44) => List(TerrainType.GRASSLAND, TerrainType.RIVER);
          case (4, 68, 36) => List(TerrainType.RIVER);
          case (140, 116, 116) => List(TerrainType.VOLCANO);
          case (208, 200, 196) => List(TerrainType.VOLCANO);
          case (104, 96, 92) => List(TerrainType.TUNDRA);
          case (56, 140, 56) => List(TerrainType.HILLS);
          case (140, 112, 88) => List(TerrainType.VOLCANO);
          case (116, 36, 36) => List(TerrainType.VOLCANO);
          case (92, 16, 16) => List(TerrainType.VOLCANO);
          case (144, 164, 144) => List(TerrainType.TUNDRA);
          case (68, 68, 4) => List(TerrainType.RIVER);
          case _ => List(TerrainType.OCEAN);
        }
      voting map {(t) =>
        votes.put(t, votes.getOrElse(t, 0)+1);
      }
    }

    terrainGuess = votes.foldLeft(TerrainType.OCEAN)( (best, mapEntry) =>
      if (mapEntry._2 > votes.getOrElse(best, 0)) {
        mapEntry._1
      } else {
        best
      }
    );

    return new TerrainGuess(terrainGuess, borders);
  }

  def getColorSwatchFromTile(whichTile:Int, from:CardinalDirection):List[Color] = {
    val tX = (whichTile % SPRITE_SHEET_WIDTH) * TILE_WIDTH;
    val tY = (whichTile / SPRITE_SHEET_WIDTH) * TILE_HEIGHT;

    var sX = TILE_WIDTH / 2;
    var sY = TILE_HEIGHT / 2;

    /*List(terrainTileSheet.getColor(tX + sX, tY + sY),
         terrainTileSheet.getColor(tX + sX+1, tY + sY),
         terrainTileSheet.getColor(tX + sX, tY + sY+1),
         terrainTileSheet.getColor(tX + sX+1, tY + sY+1));*/

    var answer = List[Color]();
    from match {
      case CardinalDirection.CENTER =>
            CardinalDirection.valuesAll map {
          (d) =>
            answer ::= terrainTileSheet.getColor(tX + sY + d.dx, tY + sY + d.dy);
        }
      case CardinalDirection.NORTH_WEST =>
        answer ::= terrainTileSheet.getColor(tX, tY);
      case CardinalDirection.NORTH_EAST =>
        answer ::= terrainTileSheet.getColor(tX + TILE_WIDTH - 1, tY);
      case CardinalDirection.SOUTH_WEST =>
        answer ::= terrainTileSheet.getColor(tX, tY + TILE_HEIGHT - 1);
      case CardinalDirection.SOUTH_EAST =>
        answer ::= terrainTileSheet.getColor(tX + TILE_WIDTH -1, tY + TILE_HEIGHT - 1);
    }
    

    answer;
  }

  override def init(container:GameContainer):Unit = {
    terrainTileSheet = TerrainLbxReader.read(Data.originalDataPath("TERRAIN.LBX"));
    //org.lwjgl.input.Keyboard.enableRepeatEvents(false);

    for (i <- 0 until metadata.length) {

      /*var borders = new Array[Option[TerrainType]](CardinalDirection.values.length);
      for (j <- 0 until borders.length) {
        borders(j) = None;
      }*/

      var plane = if (i < 888) {
          Plane.ARCANUS
        } else {
          Plane.MYRROR
        };
      var terrainGuess = guessTerrain(i, plane);
      metadata(i) = new EditableTerrainTileMetadata(i,
                                            terrainGuess.terrainGuess,
                                            terrainGuess.borders,
                                            plane, None);
    }
  }

  override def update(container:GameContainer, delta:Int):Unit = {
    val input = container.getInput();

    val keys = Array(
      Input.KEY_K,
      Input.KEY_I,
      Input.KEY_O,
      Input.KEY_L,
      Input.KEY_PERIOD,
      Input.KEY_COMMA,
      Input.KEY_M,
      Input.KEY_J,
      Input.KEY_U);
    
    (keys zip CardinalDirection.valuesAll) map {
      case (k, d) =>
        if (input.isKeyPressed(k)) {
          currentDirection = d;
          input.clearKeyPressedRecord();
        }
    }

    if (input.isKeyPressed(Input.KEY_RBRACKET)) {
      if (currentTile < TILE_COUNT-1) {
        currentTile += 1;
        println(
          getColorSwatchFromTile(currentTile, CardinalDirection.CENTER) map {
            (c) =>
              "(" + c.getRed() + " " + c.getGreen() + " " + c.getBlue() + ")"
          });
      }
      input.clearKeyPressedRecord();
    }

    if (input.isKeyPressed(Input.KEY_LBRACKET)) {
      if (currentTile > 0) {
        currentTile -= 1;
      }
      input.clearKeyPressedRecord();
    }

    val terrainKeys = Array(
      Input.KEY_Q,
      Input.KEY_W,
      Input.KEY_E,
      Input.KEY_R,
      Input.KEY_T,
      Input.KEY_A,
      Input.KEY_S,
      Input.KEY_D,
      Input.KEY_F,
      Input.KEY_G,
      Input.KEY_Z,
      Input.KEY_X,
      Input.KEY_C,
      Input.KEY_V,
      Input.KEY_B
    );

    (terrainKeys zip TerrainType.values) map {
      case (k, t) =>
        if (input.isKeyPressed(k)) {
          if (currentDirection == CardinalDirection.CENTER) {
            metadata(currentTile).terrainType = t
          } else {
            metadata(currentTile).borderingTerrainTypes(currentDirection.id) = Some(t);
          }
          input.clearKeyPressedRecord();
        }
    }

    if (input.isKeyPressed(Input.KEY_P)) {
          metadata(currentTile).plane match {
            case Plane.ARCANUS => metadata(currentTile).plane = Plane.MYRROR;
            case Plane.MYRROR => metadata(currentTile).plane = Plane.ARCANUS;
          }

          input.clearKeyPressedRecord();
        }

    if (input.isKeyPressed(Input.KEY_ESCAPE)) {
      writeOut();
      System.exit(0);
    }
  }

  def writeOut() {
    
    val doc =
      <terrain>
        {metadata.map(m =>{m.toNode()})}
      </terrain>;

    save(Data.path("terrainMetaData.xml"), doc, "utf-8", true);
  }

  override def render(container:GameContainer, graphics:Graphics):Unit = {

    val tX = (currentTile % SPRITE_SHEET_WIDTH) * TILE_WIDTH;
    val tY = (currentTile / SPRITE_SHEET_WIDTH) * TILE_HEIGHT;

    val dX = (container.getWidth()-BIG_WIDTH)/2;
    val dY = (container.getHeight()-BIG_HEIGHT)/2;

    terrainTileSheet.draw(
      dX, dY, dX + BIG_WIDTH, dY + BIG_HEIGHT,
      tX, tY, tX + TILE_WIDTH, tY + TILE_HEIGHT
    );

    // draw direction box
    val cX = container.getWidth()/2;
    val cY = container.getHeight()/2;
    
    val x1 = cX + BIG_WIDTH * currentDirection.dx - (10 + BIG_WIDTH/2);
    val y1 = cY + BIG_HEIGHT * currentDirection.dy - (10 + BIG_HEIGHT / 2);
    
    val x2 = cX + BIG_WIDTH * currentDirection.dx + (10 + BIG_WIDTH/2);
    val y2 = cY + BIG_HEIGHT * currentDirection.dy + (10 + BIG_HEIGHT / 2);

    graphics.setColor(uiColor);
    graphics.drawRect(x1, y1, x2-x1, y2-y1);

    graphics.drawString(
      metadata(currentTile).plane,
      dX,
      dY);

    graphics.drawString(
      metadata(currentTile).terrainType,
      dX,
      dY + 16
      );

    graphics.drawString(
      "Tile #" + currentTile,
      dX,
      dY + 32
      );

    (metadata(currentTile).borderingTerrainTypes zip CardinalDirection.values) map {
      case (optionalTerrain, direction) =>
        graphics.drawString(optionalTerrain match {
            case Some(terrain) => terrain
            case _ => ""
          },
                            cX + BIG_WIDTH * direction.dx - BIG_WIDTH/2,
                            cY + BIG_HEIGHT * direction.dy - BIG_HEIGHT / 2)
    }    
  }
}

object TerrainMetadataEditor {
  def main(args: Array[String]): Unit = {
    var app = new AppGameContainer(new TerrainMetadataEditor("Master of Magic: Terrain Metadata Editor"));
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(640, 400, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app.start();
  }
}
