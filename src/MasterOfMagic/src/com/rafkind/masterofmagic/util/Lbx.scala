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

  val _metaData = readLbx();

  def metaData = _metaData;

  /* Read one byte */
  def read():Int = file.read()

  def read(byteArray:Array[Byte]) = file.readFully(byteArray);

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

  private def readLbx():Lbx = {
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

  def putpixel(imageBuffer:ImageBuffer,
               x:Int,
               y:Int,
               colorIndex:Int):Unit = {

    val color = Colors.colors(colorIndex);
    val r = color.getRed();
    val g = color.getGreen();
    val b = color.getBlue();
    val a = color.getAlpha();
    
    imageBuffer.setRGBA(x, y, r, g, b, a);
  }

  def readAnd(fileName:String, 
              withPixelDo:(Int, Int, Int, Int) => Unit
              ):Unit = {
    val lbxFile = new LbxReader(fileName)

    val lbx = lbxFile.metaData

    var position:Int = lbx.subfileStart(0) + 192; // 192 byte header
    for (index <- 0 until TILE_COUNT) {
      lbxFile.seek(position + 16); // skip 8 word header

      // wierd x/y flippage!
      for (x <- 0 until TILE_WIDTH) {
        for (y <- 0 until TILE_HEIGHT) {
          withPixelDo(index, x, y, lbxFile.read())
        }
      }
      // skip 4 word footer

      // next image
      position = position + 384;
    }

    lbxFile.close();
  }

  def read(fileName:String):Image = {
    val imageBuffer = new ImageBuffer(
      TILE_WIDTH * SPRITE_SHEET_WIDTH,
      TILE_HEIGHT * SPRITE_SHEET_HEIGHT);

    readAnd(fileName, (tile, x, y, color) => {
        val row = tile / SPRITE_SHEET_WIDTH;
        val col = tile % SPRITE_SHEET_WIDTH;
        val px = (col * TILE_WIDTH) + x;
        val py = (row * TILE_HEIGHT) + y;
        putpixel(imageBuffer, px, py, color)
    });

    return imageBuffer.getImage(Image.FILTER_NEAREST);
  }
}