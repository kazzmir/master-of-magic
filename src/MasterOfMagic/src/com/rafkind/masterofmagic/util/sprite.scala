package com.rafkind.masterofmagic.util

import org.newdawn.slick._;

// http://www.roughseas.ca/momime/phpBB3/viewtopic.php?f=1&t=5

object SpriteReaderHelper {
  def create(width:Int, height:Int) =
    new ImageBuffer(width, height);

  def reset(image:ImageBuffer) = {
    for (y <- 0 until image.getHeight()) {
      for (x <- 0 until image.getWidth()) {
        image.setRGBA(x, y, 0, 0, 0, 0);
      }
    }
  }

  def copy(source:ImageBuffer) = {
    import java.nio.ByteOrder;
    
    val sourceRgba = source.getRGBA();
    var dest = new ImageBuffer(source.getWidth(), source.getHeight());

    if (ByteOrder.nativeOrder() == ByteOrder.BIG_ENDIAN) {
      for (y <- 0 until source.getHeight()) {
        for (x <- 0 until source.getWidth()) {
          val offset = (x + (y * source.getTexWidth())) * 4;
          dest.setRGBA(x, y, sourceRgba(offset+2), sourceRgba(offset+1), sourceRgba(offset), sourceRgba(offset+3));
        }
      }
    } else {
      for (y <- 0 until source.getHeight()) {
        for (x <- 0 until source.getWidth()) {
          val offset = (x + (y * source.getTexWidth())) * 4;
          dest.setRGBA(x, y, sourceRgba(offset), sourceRgba(offset+1), sourceRgba(offset+2), sourceRgba(offset+3));
        }
      }
    }

    dest;
  }

  def withPixelDo(image:ImageBuffer, x:Int, y:Int, color:Color) =
    image.setRGBA(x, y, color.getRed(), color.getGreen(), color.getBlue(), color.getAlpha());
  

  def finish(image:ImageBuffer) =
    image.getImage(Image.FILTER_NEAREST);
}

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

  def readPalette(lbxReader:LbxReader, index:Int, header:Header, paletteInfo:PaletteInfo):Array[Color] = {
    if (header.paletteInfoOffset > 0) {
      var pal = new Array[Color](Colors.colors.length);
      for (i <- 0 until pal.length) {
        pal(i) = Colors.colors(i);
      }

      lbxReader.seek(lbxReader.metaData.subfileStart(index) + paletteInfo.paletteOffset);
      for (c <- 0 until paletteInfo.count) {
        val r:Int = lbxReader.read() & 0xFF;
        val g:Int = lbxReader.read() & 0xFF;
        val b:Int = lbxReader.read() & 0xFF;
        pal(c + paletteInfo.firstPaletteColorIndex) = new Color(r*4, g*4, b*4);
      }
      pal;
    } else {
      Colors.colors
    }
  }

  def readOffsets(lbxReader:LbxReader, bitmaps:Int) = {
    val offsets = for (index <- 0 to bitmaps) yield {
      lbxReader.read4()
    }
    for (index <- 0 until bitmaps) yield {
      Offset(offsets(index), offsets(index+1))
    }
  }

  def render(bitmapNumber:Int,
                data:Array[Int],
                header:Header,
                paletteInfo:PaletteInfo,
                palette:Array[Color],
                target:ImageBuffer):Unit = {

    var index = 0;
    if (data(index) == 1 && bitmapNumber > 0) {
      SpriteReaderHelper.reset(target);
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
                  SpriteReaderHelper.withPixelDo(target, x, y, palette(data(last_pos)));
                } else {
                  throw new Exception("Overrun");

                }
                y += 1;
                rle_counter += 1;
              }
              n_r += 2;
            } else {
              if ((x < header.width) && (y < header.height) && (x >= 0) && (y >= 0)) {
                SpriteReaderHelper.withPixelDo(target, x, y, palette(data(n_r)));
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

  def read(lbxReader:LbxReader, groupIndex:Int) = {

    val lbxMetaData = lbxReader.metaData
    
    val header = readHeader(lbxReader, groupIndex)
    
    val offsets = readOffsets(lbxReader, header.bitmapCount)
    val paletteInfo = readPaletteInfo(lbxReader, groupIndex, header.paletteInfoOffset)
    val palette = readPalette(lbxReader, groupIndex, header, paletteInfo)

    var canvas = SpriteReaderHelper.create(header.width, header.height);
    SpriteReaderHelper.reset(canvas);
    def readSprite(bitmapNumber:Int, offset:Offset) = {
      lbxReader.seek(lbxMetaData.subfileStart(groupIndex) + offset.start);
      var data = new Array[Byte](offset.end - offset.start);
      lbxReader.read(data);
      // make it unsigned
      var data2 = new Array[Int](data.length);
      for (i <- 0 until data.length) {
        data2(i) = data(i) & 0xFF;
      }
      render(bitmapNumber, data2, header, paletteInfo, palette, canvas);
      SpriteReaderHelper.copy(canvas);
    }

    val answer = new Array[Image](offsets.size);
    for (index <- 0 until offsets.size) {
      answer(index) = SpriteReaderHelper.finish(
        readSprite(index, offsets(index)));
    }

    answer;
  }
}