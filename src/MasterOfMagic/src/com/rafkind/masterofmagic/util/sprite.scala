package com.rafkind.masterofmagic.util

// http://www.roughseas.ca/momime/phpBB3/viewtopic.php?f=1&t=5

object SpriteReader {

  case class Header(width:Int,
                    height:Int,
                    unknown1:Int,
                    bitmapCount:Int,
                    unknown2:Int,
                    unknown3:Int,
                    unknown4:Int,
                    paletteInfoOffset:Int,
                    unknown5:Int) {
    def debug() {
      println("  width= "
              + width
              + ", height= "
              + height
              + ", bitmapCount= "
              + bitmapCount
              + ", paletteInfoOffset="
              + paletteInfoOffset)
    }
  }

  case class PaletteInfo(paletteOffset:Int, firstPaletteColorIndex:Int, count:Int, unknown:Int)
  case class Offset(start:Int, end:Int)

  def readHeader(lbxReader:LbxReader, index:Int):Header = {
    lbxReader.seek(lbxReader.metaData.subfileStart(index))
    val width = lbxReader.read2()
    val height = lbxReader.read2()
    val unknown1 = lbxReader.read2()
    val bitmapCount = lbxReader.read2()
    val unknown2 = lbxReader.read2()
    val unknown3 = lbxReader.read2()
    val unknown4 = lbxReader.read2()
    val paletteInfoOffset = lbxReader.read2()
    val unknown5 = lbxReader.read2()
    Header(width, height, unknown1, bitmapCount, unknown2, unknown3, unknown4, paletteInfoOffset, unknown5)
  }

  def readPaletteInfo(lbxReader: LbxReader, index:Int, paletteInfoOffset:Int):PaletteInfo = {
    if (paletteInfoOffset > 0){
      lbxReader.seek(lbxReader.metaData.subfileStart(index) + paletteInfoOffset)
      val paletteOffset = lbxReader.read2()
      val firstPaletteColorIndex = lbxReader.read2()
      val count = lbxReader.read2()
      val unknown = lbxReader.read2()
      PaletteInfo(paletteOffset, firstPaletteColorIndex, count, unknown)
    } else {
      /* Default palette info */
      PaletteInfo(0, 0, 255, 0)
    }
  }

  def readPalette(lbxReader:LbxReader, index:Int, paletteOffset:Int):Array[java.awt.Color] = {
    /* TODO */
    Colors.colors
  }

  def readOffsets(lbxReader:LbxReader, bitmaps:Int) = {
    val offsets = for (index <- 0 to bitmaps) yield {
      lbxReader.read4()
    }
    for (index <- 0 until bitmaps) yield {
      Offset(offsets(index), offsets(index+1))
    }
  }

  def render[T](bitmapNumber:Int, 
                data:Array[Int],
                header:Header,
                paletteInfo:PaletteInfo,
                palette:Array[java.awt.Color],
                target:T,
                resetter:(T) => Unit,
                withPixelDo:(T, Int, Int, java.awt.Color) => Unit):Unit = {

    var index = 0;
    if (data(index) == 1 && bitmapNumber > 0) {
      resetter(target);
    }
    index = 1;

    var x = 0;
    var rle_value = 0;
    var last_pos = 0;
    while ((x < header.width) && (index < data.length)) {
      var y = 0;
      if (data(index) == 0xFF) {
        index += 1;
        rle_value = paletteInfo.firstPaletteColorIndex + paletteInfo.count;
      } else {
        var long_data = data(index + 2);
        var next_ctl = index + data(index + 1) + 2;

        data(index) match {
          case 0 => rle_value = paletteInfo.firstPaletteColorIndex + paletteInfo.count;
          case 0x80 => rle_value = 0xE0;
          case _ => throw new Exception("Bad RLE Value");
        }

        y = data(index+3);
        index += 4;

        var n_r = index;
        while (n_r < next_ctl) {
          while ((n_r < (index + long_data)) && (x < header.width)) {
            if (data(n_r) >= rle_value) {
              last_pos = n_r + 1;
              var rle_length = data(n_r) - rle_value + 1;
              var rle_counter = 0;
              while ((rle_counter < rle_length) && (y < header.height)) {
                if ((x < header.width) && (y < header.height) && (x >= 0) && (y >= 0)) {
                  var color = palette(data(last_pos));
                  withPixelDo(target, x, y, color);
                } else {
                  throw new Exception("Overrun");

                }
                y += 1;
                rle_counter += 1;
              }
              n_r += 2;
            } else {
              if ((x < header.width) && (y < header.height) && (x >= 0) && (y >= 0)) {
                var color = palette(data(n_r));
                withPixelDo(target, x, y, color);
              }
              n_r += 1;
              y += 1;
            }
          }

          if (n_r < next_ctl) {
            y += data(n_r + 1);
            index = n_r + 2;
            long_data = data(n_r);
            n_r += 2;
          }
        }
        index = next_ctl;
      }
      x += 1;
    }
  }

  def read[T](lbxReader:LbxReader, index:Int, creator:(Int, Int) => T, resetter:(T) => Unit, copier:(T) => T, withPixelDo:(T, Int, Int, java.awt.Color) => Unit) = {
    val lbxMetaData = lbxReader.metaData
    println(index + ": from " + lbxMetaData.subfile(index).start + " to " + lbxMetaData.subfile(index).end);
    
    val header = readHeader(lbxReader, index)
    header.debug();
    
    val offsets = readOffsets(lbxReader, header.bitmapCount)
    for (o <- offsets) {
      println("    bitmap from " + o.start + " to " + o.end)
    }
    val paletteInfo = readPaletteInfo(lbxReader, index, header.paletteInfoOffset)
    val palette = readPalette(lbxReader, index, paletteInfo.paletteOffset)

    var canvas = creator(header.width, header.height);
    resetter(canvas);
    def readSprite(bitmapNumber:Int, offset:Offset) = {
      lbxReader.seek(lbxMetaData.subfileStart(index) + offset.start);
      var data = new Array[Byte](offset.end - offset.start);
      lbxReader.read(data);
      // make it unsigned
      var data2 = new Array[Int](data.length);
      for (i <- 0 until data.length) {
        data2(i) = data(i) & 0xFF;
      }
      render(bitmapNumber, data2, header, paletteInfo, palette, canvas, resetter, withPixelDo);
      copier(canvas);
    }

    var c:Int = 0;
    for (offset <- offsets) yield {
      val s = readSprite(c, offset)
      c += 1;
      s;
    }
  }
}