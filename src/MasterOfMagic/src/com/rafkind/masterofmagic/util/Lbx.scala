package com.rafkind.masterofmagic.util

import java.io._;
import scala.collection.mutable._;
import org.newdawn.slick._;

// format from here:
// http://www.roughseas.ca/momime/phpBB3/viewtopic.php?f=1&t=3

class Lbx(val subfileCount:Int, val magicNumber:Int, val version:Int) {
  var size:Int = 0

  // Contains start and end offsets for an lbx data structure
  case class LbxData(val start:Int, var end:Int)

  var subfiles:Map[Int, LbxData] = new HashMap[Int, LbxData]

  def addSubfile(index:Int, offset:LbxData):Map[Int, LbxData] = {
    subfiles += index -> offset

    subfiles
  }

  def setSize(offset:Int):Lbx = {
    size = offset
    this
  }

  def subfileStart(index:Int):Int = subfiles(index).start
  def subfileEnd(index:Int):Int = subfiles(index).end

  def subFiles():Int = subfiles.size
}

object LbxReader{
  def read2(file:RandomAccessFile):Int = {
    var a = file.read()
    var b = file.read()

    a | (b << 8)
  }

  def read4(file:RandomAccessFile):Int = {
    var a = file.read()
    var b = file.read()
    var c = file.read()
    var d = file.read()

    a | (b << 8) | (c << 16) | (d << 24)
  }

  def read(fileName:String):Lbx = {
    val lbxFile = new RandomAccessFile(new File(fileName), "r");
    lbxFile.seek(0)

    val subfileCount = read2(lbxFile)
    val magicNumber = read4(lbxFile)
    val version = read2(lbxFile)

    val lbx = new Lbx(subfileCount, magicNumber, version)
    
    val offsets = (for (s <- 0 until subfileCount) yield read4(lbxFile)) ++ List(lbxFile.length.intValue)

    for (index <- 0 until subfileCount){
      lbx.addSubfile(index, lbx.LbxData(offsets(index), offsets(index+1)))
    }

    lbx.setSize(read4(lbxFile))
    
    lbxFile.close()
    lbx
  }
}

object TerrainLbxReader {
  
  // there are 1761 images,
  // so i'll put them all on the same
  // sprite sheet'
  val SPRITE_SHEET_WIDTH = 40;
  val SPRITE_SHEET_HEIGHT = 45;

  val TILE_WIDTH = 20;
  val TILE_HEIGHT = 18;

  val TILE_COUNT = 1761;

  def fatpixel(imageBuffer:ImageBuffer,
               x:Int,
               y:Int,
               colorIndex:Int):Unit = {

    val color = Colors.colors(colorIndex);
    val r = color.getRed();
    val g = color.getGreen();
    val b = color.getBlue();
    val a = color.getAlpha();
    
    imageBuffer.setRGBA(x, y, r, g, b, a);
    imageBuffer.setRGBA(x+1, y, r, g, b, a);
    imageBuffer.setRGBA(x, y+1, r, g, b, a);
    imageBuffer.setRGBA(x+1, y+1, r, g, b, a);
  }

  def read(fileName:String):Image = {
    var lbx = LbxReader.read(fileName);

    var imageBuffer = new ImageBuffer(
      TILE_WIDTH * 2 * SPRITE_SHEET_WIDTH,
      TILE_HEIGHT * 2 * SPRITE_SHEET_HEIGHT);

    var row:Int = 0;
    var col:Int = 0;

    var lbxFile = new RandomAccessFile(new File(fileName), "r");

    var position:Int = lbx.subfileStart(0) + 192; // 192 byte header
    for (index <- 0 until TILE_COUNT) {
      lbxFile.seek(position + 16); // skip 8 word header

      // wierd x/y flippage!
      for (y <- 0 until TILE_WIDTH) {
        for (x <- 0 until TILE_HEIGHT) {
          val c = lbxFile.read();
          val px = (col * TILE_WIDTH * 2) + (y * 2);
          val py = (row * TILE_HEIGHT * 2) + (x * 2);
          fatpixel(imageBuffer, px, py, c);
        }
      }
      // skip 4 word footer

      // next image
      position = position + 384;
      col = col+1;
      if (col >= SPRITE_SHEET_WIDTH) {
        col = 0;
        row = row + 1;
      }
    }

    lbxFile.close();
    
    return imageBuffer.getImage();
  }
}
