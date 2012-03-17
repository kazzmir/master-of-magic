/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util.terrain

import org.newdawn.slick._;

import com.rafkind.masterofmagic.system._;
import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.util._;
import scala.xml.XML._;
import scala.xml._;
import scala.collection.mutable.HashMap;
import scala.collection.mutable.HashSet;

object EditableTerrainTileMetadata {

  private def copyBorderingArray(other:Array[Option[TerrainType]]):Array[Option[TerrainType]] =
    other map ((x) => x match {
        case Some(y) => Some(y);
        case None => None
      });

  def copy(other:EditableTerrainTileMetadata) =
    new EditableTerrainTileMetadata(other.id, other.terrainType, copyBorderingArray(other.borderingTerrainTypes), other.plane, other.parentId);
}

// mirrors TerrainTileMetadata in State.scala
class EditableTerrainTileMetadata(
  var id:Int,
  var terrainType:TerrainType,
  var borderingTerrainTypes:Array[Option[TerrainType]],
  var plane:Plane,
  var parentId:Option[TerrainTileMetadata]) {

  def getTerrain(dir:CardinalDirection):Option[TerrainType] = {
    dir match {
      case CardinalDirection.CENTER => Some(terrainType)
      case _ => borderingTerrainTypes(dir.id)
    }
  }

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

  def shorelineAdjust():Unit = {
    var grid = new Array[Option[TerrainType]](9);

    grid(4) = Some(terrainType);
    for (c <- CardinalDirection.values) {
      val x = c.dx + 1;
      val y = c.dy + 1;
      val i = x + (y*3);
      grid(i) = borderingTerrainTypes(c.id);
    }

    for (y <- 0 until 3) {
      for (x <- 0 until 3) {
        for (c <- CardinalDirection.values) {
          val newX = x + c.dx;
          val newY = y + c.dy;

          if (newX >= 0 && newX <= 2 && newY >= 0 && newY <= 2) {
            val oldId = x + (y*3);
            val newId = newX + (newY*3);
            //println(oldId + ": " + grid(oldId) + " - " + newId + ": " + grid(newId));
            (grid(oldId), grid(newId)) match {
              case (Some(t1), Some(t2)) if (t1 == TerrainType.OCEAN && !(t2 == TerrainType.OCEAN || t2 == TerrainType.SHORE)) => {
                  grid(oldId) = Some(TerrainType.SHORE);
              }
              case _ =>
            }

            /*if (grid(oldId) == TerrainType.OCEAN && !(grid(newId) == TerrainType.OCEAN || grid(newId) == TerrainType.SHORE)) {
              grid(oldId) == TerrainType.SHORE;
            }*/
          }
        }
      }
    }

    grid(4) match {
      case Some(t) => terrainType = t;
      case _ =>
    }
    for (c <- CardinalDirection.values) {
      val x = c.dx + 1;
      val y = c.dy + 1;
      val i = x + (y*3);
      borderingTerrainTypes(c.id) = grid(i);
    }
  }
}

class TerrainMetadataEditor(title:String) extends BasicGame(title) {
  import com.rafkind.masterofmagic.util.TerrainLbxReader._;
  
  var terrainTileSheet:Image = null;

  var currentTile:Int = 0;

  var currentDirection:CardinalDirection = CardinalDirection.CENTER;

  val BIG_WIDTH = 133;
  val BIG_HEIGHT = BIG_WIDTH * TILE_HEIGHT / TILE_WIDTH;

  val uiColor = Color.white;
  val guessColor = Color.yellow;

  val metadataGuess = new Array[EditableTerrainTileMetadata](TILE_COUNT);
  val metadata = new HashMap[Int, EditableTerrainTileMetadata]();

  try {
    load(Data.path("terrainMetaData.xml")) \ "metadata" foreach { (m) =>
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

  def representativeDirection(d:CardinalDirection):CardinalDirection = {
    d match {
      case CardinalDirection.CENTER => CardinalDirection.CENTER;
      case CardinalDirection.NORTH => CardinalDirection.NORTH;
      case CardinalDirection.SOUTH => CardinalDirection.NORTH;
      case CardinalDirection.EAST => CardinalDirection.NORTH;
      case CardinalDirection.WEST => CardinalDirection.NORTH;
      case CardinalDirection.NORTH_EAST => CardinalDirection.NORTH_EAST;
      case CardinalDirection.NORTH_WEST => CardinalDirection.NORTH_EAST;
      case CardinalDirection.SOUTH_EAST => CardinalDirection.NORTH_EAST;
      case CardinalDirection.SOUTH_WEST => CardinalDirection.NORTH_EAST;
    }
  }

  def guessTerrain():Unit = {
    val directionalModels = new HashMap[CardinalDirection,
                                        HashMap[Tuple3[Int, Int, Int], HashSet[TerrainType]]]();
    for (c <- List(CardinalDirection.CENTER, CardinalDirection.NORTH, CardinalDirection.NORTH_EAST)) {
      val d = representativeDirection(c);
      directionalModels += d -> new HashMap[Tuple3[Int, Int, Int], HashSet[TerrainType]]();
    }
    
    for (k <- metadata.keys) {
      val tile = metadata(k);
      for (c <- CardinalDirection.valuesAll) {
        val colors = getColorSwatchFromTile(k, c);
        val model = directionalModels(representativeDirection(c));
        for (color <- colors) {
          val colorTuple = (color.getRed(), color.getGreen(), color.getBlue());
          var terrains = model.getOrElseUpdate(colorTuple, new HashSet[TerrainType]());
          c match {
            case CardinalDirection.CENTER => terrains += tile.terrainType;
            case dir:CardinalDirection => tile.borderingTerrainTypes(dir.id) match {
                case Some(terrainType) => terrains += terrainType;
                case _ =>
            }
          }
        }
      }
    }

    for (index <- 0 until TILE_COUNT) {
      var newTerrain = newBlankTerrain(index);
      for (direction <- CardinalDirection.valuesAll) {
        val votes = new HashMap[TerrainType, Int];
        val model = directionalModels(representativeDirection(direction));
        for (color <- getColorSwatchFromTile(index, direction)) {
          val colorTuple = (color.getRed(), color.getGreen(), color.getBlue());
          val terrains = model.getOrElse(colorTuple, new HashSet());
          for (terrain <- terrains) {
            votes.put(terrain, votes.getOrElse(terrain, 0) + 1);
          }
        }

        val terrainGuess = votes.foldLeft(TerrainType.OCEAN)( (best, mapEntry) =>
          if (mapEntry._2 > votes.getOrElse(best, 0)) {
            mapEntry._1
          } else {
            best
          }
        );

        direction match {
          case CardinalDirection.CENTER => newTerrain.terrainType = terrainGuess;
          case d:CardinalDirection => newTerrain.borderingTerrainTypes(d.id) = Some(terrainGuess);
        }
      }
      newTerrain.shorelineAdjust();
      metadataGuess(index) = newTerrain;
    }
  }

  def getColorSwatchFromTile(whichTile:Int, from:CardinalDirection):List[Color] = {
    val tX = (whichTile % SPRITE_SHEET_WIDTH) * TILE_WIDTH;
    val tY = (whichTile / SPRITE_SHEET_WIDTH) * TILE_HEIGHT;

    val sX = TILE_WIDTH / 2;
    val sY = TILE_HEIGHT / 2;

    var answer = List[Color]();
    from match {
      case CardinalDirection.CENTER =>
        for (d <- CardinalDirection.valuesAll) {
          answer ::= terrainTileSheet.getColor(tX + sY + d.dx*2, tY + sY + d.dy*2);
        }
      case CardinalDirection.NORTH_WEST =>
        for (d <- CardinalDirection.valuesAll) {
          answer ::= terrainTileSheet.getColor(tX + 1, tY + 1);
        }
      case CardinalDirection.NORTH_EAST =>
        for (d <- CardinalDirection.valuesAll) {
          answer ::= terrainTileSheet.getColor(tX + TILE_WIDTH - 2, tY + 1);
        }
      case CardinalDirection.SOUTH_WEST =>
        for (d <- CardinalDirection.valuesAll) {
          answer ::= terrainTileSheet.getColor(tX + 1, tY + TILE_HEIGHT - 2);
        }
      case CardinalDirection.SOUTH_EAST =>
        for (d <- CardinalDirection.valuesAll) {
          answer ::= terrainTileSheet.getColor(tX + TILE_WIDTH - 2, tY + TILE_HEIGHT - 2);
        }
      case CardinalDirection.NORTH =>
        for (d <- 0 until sX) {
          answer ::= terrainTileSheet.getColor(tX + d + sX / 2, tY);
        }
      case CardinalDirection.SOUTH =>
        for (d <- 0 until sX) {
          answer ::= terrainTileSheet.getColor(tX + d + sX / 2, tY + TILE_HEIGHT - 1);
        }
      case CardinalDirection.WEST =>
        for (d <- 0 until sY) {
          answer ::= terrainTileSheet.getColor(tX, tY + d + sY / 2);
        }
      case CardinalDirection.EAST =>
        for (d <- 0 until sY) {
          answer ::= terrainTileSheet.getColor(tX + TILE_WIDTH - 1, tY + d + sY/2);
        }
      case _ =>
        
    }
    

    answer;
  }

  /* Arcanus or Myrror depending on the tile */
  def whichPlane(tile:Int):Plane = if (tile < 888) Plane.ARCANUS else Plane.MYRROR

  override def init(container:GameContainer):Unit = {
    terrainTileSheet = TerrainLbxReader.read(Data.originalDataPath("TERRAIN.LBX"));
    //org.lwjgl.input.Keyboard.enableRepeatEvents(false);

    /*for (i <- 0 until metadata.length) {

      /*var borders = new Array[Option[TerrainType]](CardinalDirection.values.length);
      for (j <- 0 until borders.length) {
        borders(j) = None;
      }*/

      var plane = whichPlane(i)
      var terrainGuess = guessTerrain(i, plane);
      metadata(i) = new EditableTerrainTileMetadata(i,
                                            terrainGuess.terrainGuess,
                                            terrainGuess.borders,
                                            plane, None);
    }*/
  }

  def newBlankTerrain(id:Int):EditableTerrainTileMetadata = {
    val borders = new Array[Option[TerrainType]](CardinalDirection.values.length);
    for (j <- 0 until borders.length) {
      borders(j) = None;
    }

    return new EditableTerrainTileMetadata(id,
                                           TerrainType.OCEAN,
                                           borders,
                                           whichPlane(id),
                                           None);
  }

  def copyGuess():Unit = {
    for (guess <- metadataGuess) {
      metadata.get(guess.id) match {
        case Some(m) =>
          for (i <- 0 until m.borderingTerrainTypes.length) {
            m.borderingTerrainTypes(i) = guess.borderingTerrainTypes(i);
          }
        case None =>
          metadata += guess.id -> EditableTerrainTileMetadata.copy(guess);
      }
    }
  }

  /* The order maps to Terraintypes */
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

  override def update(container:GameContainer, delta:Int):Unit = {
    val input = container.getInput();

    val keys = Array(     
      Input.KEY_I,
      Input.KEY_O,
      Input.KEY_L,
      Input.KEY_PERIOD,
      Input.KEY_COMMA,
      Input.KEY_M,
      Input.KEY_J,
      Input.KEY_U,
      Input.KEY_K);
    
    (keys zip CardinalDirection.valuesAll) map {
      case (k, d) =>
        if (input.isKeyPressed(k)) {
          currentDirection = d
          input.clearKeyPressedRecord()
        }
    }

    def directionToCoordinate(direction:CardinalDirection) = direction match {
      case CardinalDirection.CENTER => (0, 0)
      case CardinalDirection.NORTH => (0, -1)
      case CardinalDirection.SOUTH => (0, 1)
      case CardinalDirection.EAST => (1, 0)
      case CardinalDirection.WEST => (-1, 0)
      case CardinalDirection.NORTH_EAST => (1, -1)
      case CardinalDirection.NORTH_WEST => (-1, -1)
      case CardinalDirection.SOUTH_EAST => (1, 1)
      case CardinalDirection.SOUTH_WEST => (-1, 1)
    }

    def coordinateToDirection(x:Int, y:Int) = (x, y) match {
      case (0, 0) => CardinalDirection.CENTER
      case (0, -1) => CardinalDirection.NORTH
      case (0, 1) => CardinalDirection.SOUTH
      case (1, 0) => CardinalDirection.EAST
      case (-1, 0) => CardinalDirection.WEST
      case (1, -1) => CardinalDirection.NORTH_EAST
      case (-1, -1) => CardinalDirection.NORTH_WEST
      case (1, 1) => CardinalDirection.SOUTH_EAST
      case (-1, 1) => CardinalDirection.SOUTH_WEST
    }

    def between(low:Int, here:Int, high:Int) = Math.max(Math.min(here, high), low)

    var coordinate = directionToCoordinate(currentDirection)
    val directionKeys = Map(Input.KEY_DOWN -> (0, 1),
                            Input.KEY_LEFT -> (-1, 0),
                            Input.KEY_RIGHT -> (1, 0),
                            Input.KEY_UP -> (0, -1))

    for ((key, move) <- directionKeys){
      if (input.isKeyPressed(key)){
        coordinate = (coordinate, move) match {
          case ((coordinateX, coordinateY),
                (moveX, moveY)) => (between(-1, coordinateX + moveX, 1),
                                    between(-1, coordinateY + moveY, 1))
        }
        // input.clearKeyPressedRecord();
      }
    }

    currentDirection = coordinate match {
      case (x, y) => coordinateToDirection(x, y)
    }

    val step = input.isKeyDown(Input.KEY_LSHIFT) match {
       case true => 10
       case false => 1
    }

    if (input.isKeyPressed(Input.KEY_RBRACKET)) {
      currentTile = (currentTile + step) % TILE_COUNT;
        /*println(
          getColorSwatchFromTile(currentTile, CardinalDirection.CENTER) map {
            (c) =>
              "(" + c.getRed() + " " + c.getGreen() + " " + c.getBlue() + ")"
          });*/
      
      input.clearKeyPressedRecord();
    }

    if (input.isKeyPressed(Input.KEY_LBRACKET)) {
      currentTile = (currentTile + TILE_COUNT - step) % TILE_COUNT;
      input.clearKeyPressedRecord();
    }

    (terrainKeys zip TerrainType.values) map {      
      case (k, t) =>
        if (input.isKeyPressed(k)) {
          val tile = metadata.getOrElseUpdate(currentTile, newBlankTerrain(currentTile));
          if (currentDirection == CardinalDirection.CENTER) {
            tile.terrainType = t
          } else {
            tile.borderingTerrainTypes(currentDirection.id) = Some(t);
          }
          input.clearKeyPressedRecord();
        }
    }

    if (input.isKeyPressed(Input.KEY_P)) {
      val tile = metadata.getOrElseUpdate(currentTile, newBlankTerrain(currentTile));
      tile.plane match {
        case Plane.ARCANUS => tile.plane = Plane.MYRROR;
        case Plane.MYRROR => tile.plane = Plane.ARCANUS;
      }

      input.clearKeyPressedRecord();
    }

    if (input.isKeyPressed(Input.KEY_ESCAPE)) {
      println("Saving..")
      writeOut();
      println("Done")
      System.exit(0);
    }

    if (input.isKeyPressed(Input.KEY_SPACE)) {
      //guessTerrain();
      NeuralNetworkGuesser.guessTerrain(metadataGuess);
      input.clearKeyPressedRecord();
    }

    /*if (input.isKeyPressed(Input.KEY_ENTER)) {
      copyGuess();
      input.clearKeyPressedRecord();
    }*/
  }

  def writeOut() {
    
    val doc =
      <terrain>
        {metadata.map((kv) =>{kv._2.toNode()})}
      </terrain>;

    save(Data.path("terrainMetaData.xml"), doc, "utf-8", true);
  }

  def canonicalKey(key:Int):String = key match {
    case Input.KEY_Q => "Q"
    case Input.KEY_W => "W"
    case Input.KEY_E => "E"
    case Input.KEY_R => "R"
    case Input.KEY_T => "T"
    case Input.KEY_A => "A"
    case Input.KEY_S => "S"
    case Input.KEY_D => "D"
    case Input.KEY_F => "F"
    case Input.KEY_G => "G"
    case Input.KEY_Z => "Z"
    case Input.KEY_X => "X"
    case Input.KEY_C => "C"
    case Input.KEY_V => "V"
    case Input.KEY_B => "B"
  }

  def canonicalTerrain(terrain:TerrainType):String = terrain.name

  def showHelp(container:GameContainer, graphics:Graphics):Unit = {
    graphics.setColor(uiColor);
    var x = 10
    val stepY = 15
    val startY = container.getHeight() - (stepY * 5)
    var y = startY
    (terrainKeys zip TerrainType.values) map {
      case (key, terrain) => {
        graphics.drawString(canonicalKey(key) + ": " + canonicalTerrain(terrain), x, y)
        y += stepY
        if (y > container.getHeight() - stepY){
          y = startY
          x += 150
        }
      }
    }
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
            "Tile #" + currentTile,
            dX,
            dY + 32
            );

    metadata.get(currentTile) match {
      case Some(tile) => {
          graphics.drawString(
            tile.plane,
            dX,
            dY);

          graphics.drawString(
            tile.terrainType,
            dX,
            dY + 16
            );

          (tile.borderingTerrainTypes zip CardinalDirection.values) map {
            case (optionalTerrain, direction) =>
              graphics.drawString(optionalTerrain match {
                  case Some(terrain) => terrain
                  case _ => ""
                },
                cX + BIG_WIDTH * direction.dx - BIG_WIDTH/2,
                cY + BIG_HEIGHT * direction.dy - BIG_HEIGHT / 2)
          }
      }
      case _ =>
    }

    graphics.setColor(guessColor);

    metadataGuess(currentTile) match {
      case tile:EditableTerrainTileMetadata => {

          graphics.drawString(
            tile.terrainType,
            dX,
            dY + 48
            );

          (tile.borderingTerrainTypes zip CardinalDirection.values) map {
            case (optionalTerrain, direction) =>
              graphics.drawString(optionalTerrain match {
                  case Some(terrain) => terrain
                  case _ => ""
                },
                cX + BIG_WIDTH * direction.dx - BIG_WIDTH/2,
                cY + BIG_HEIGHT * direction.dy - BIG_HEIGHT / 2 + 48)
          }
      }
      case null =>
    }

    showHelp(container, graphics)
  }
}

object TerrainMetadataEditor {
  def main(args: Array[String]): Unit = {
    val app = new AppGameContainer(new TerrainMetadataEditor("Master of Magic: Terrain Metadata Editor"));
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(640, 400, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app.start();
  }
}
