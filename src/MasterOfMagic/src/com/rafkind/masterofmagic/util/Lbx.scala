package com.rafkind.masterofmagic.util

import java.io._;
import scala.collection.mutable._;
import org.newdawn.slick._;

// format from here:
// http://www.roughseas.ca/momime/phpBB3/viewtopic.php?f=1&t=3

// Contains start and end offsets for an lbx data structure
case class LbxData(val start:Int, var end:Int)

class Lbx{
  var size:Int = 0

  var subfiles:Map[Int, LbxData] = new HashMap[Int, LbxData]

  def addSubfile(index:Int, offset:LbxData):Map[Int, LbxData] = {
    subfiles += index -> offset

    subfiles
  }

  def setSize(offset:Int):Lbx = {
    size = offset
    this
  }

  def subfile(index:Int):LbxData = subfiles(index)
  def subfileStart(index:Int):Int = subfiles(index).start
  def subfileEnd(index:Int):Int = subfiles(index).end

  def subfileCount():Int = subfiles.size
}

class LbxReader(val path:String){
  val file = new RandomAccessFile(new File(path), "r")

  /* Read one byte */
  def read():Int = file.read()

  /* Read a word (2 bytes) */
  def read2():Int = {
    var a = file.read()
    var b = file.read()

    a | (b << 8)
  }

  /* Read a dword (4 bytes) */
  def read4():Int = {
    var a = file.read()
    var b = file.read()
    var c = file.read()
    var d = file.read()

    a | (b << 8) | (c << 16) | (d << 24)
  }

  def seek(position:Int){
    file.seek(position)
  }

  def close(){
    file.close()
  }

  def read(offset:LbxData) = {
    file.seek(offset.start)
    for (index <- 0 to offset.end) yield {
      file.read()
    }
  }

  def readLbx():Lbx = {
    file.seek(0)

    val subfileCount = read2()
    val magicNumber = read4()
    val version = read2()

    val lbx = new Lbx()
    
    val offsets = (for (s <- 0 until subfileCount) yield read4()) ++ List(file.length.intValue)

    for (index <- 0 until subfileCount){
      lbx.addSubfile(index, LbxData(offsets(index), offsets(index+1)))
    }

    lbx.setSize(read4())
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
    val lbxFile = new LbxReader(fileName)

    val imageBuffer = new ImageBuffer(
      TILE_WIDTH * 2 * SPRITE_SHEET_WIDTH,
      TILE_HEIGHT * 2 * SPRITE_SHEET_HEIGHT);

    var row:Int = 0;
    var col:Int = 0;

    val lbx = lbxFile.readLbx()

    var position:Int = lbx.subfileStart(0) + 192; // 192 byte header
    for (index <- 0 until TILE_COUNT) {
      lbxFile.seek(position + 16); // skip 8 word header

      // wierd x/y flippage!
      for (y <- 0 until TILE_WIDTH) {
        for (x <- 0 until TILE_HEIGHT) {
          val c = lbxFile.read()
          val px = (col * TILE_WIDTH * 2) + (y * 2)
          val py = (row * TILE_HEIGHT * 2) + (x * 2)
          fatpixel(imageBuffer, px, py, c)
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
