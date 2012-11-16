/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util.sprite

import com.google.common.cache.CacheBuilder
import com.google.common.cache.CacheLoader
import com.google.common.cache.RemovalListener
import com.google.common.cache.RemovalNotification
import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.system.Data
import com.rafkind.masterofmagic.util._;
import java.awt._;
import java.awt.image._;

// http://www.roughseas.ca/momime/phpBB3/viewtopic.php?f=1&t=5

object AwtSpriteReaderHelper {
  def create(width:Int, height:Int) =
    GraphicsEnvironment
      .getLocalGraphicsEnvironment()
      .getDefaultScreenDevice()
      .getDefaultConfiguration()
      .createCompatibleImage(width, height);

  def reset(image:BufferedImage) = {
    for (y <- 0 until image.getHeight()) {
      for (x <- 0 until image.getWidth()) {
        image.setRGB(x, y, 0);
      }
    }
  }

  def copy(source:BufferedImage) = 
    new BufferedImage(source.getColorModel(), 
                      source.copyData(null), 
                      source.isAlphaPremultiplied(), null);
  

  /*var shouldLog = false;
  def turnLoggingOn:Unit = {
    shouldLog = true;
  }
  def turnLoggingOff:Unit = {
    shouldLog = false;
  }

  var colorMap = new scala.collection.mutable.HashMap[Color, Int];
  def numberColors(x:Array[Color]):Unit = {
    for (index <- 0 until 256) {
      colorMap += x(index) -> index;
    }
  }*/
  
  def withPixelDo(image:BufferedImage, x:Int, y:Int, color:Color):Unit = {
    /*if (shouldLog) {
      //println("[" + x + ", " + y + "] = " + (color.getRed(), color.getGreen(), color.getBlue()));
      println("[" + x + ", " + y + "] = " + "%02X".format(color.getRed()) + "%02X".format(color.getGreen()) + "%02X".format(color.getBlue()) + ".." + colorMap(color));
      //println("[" + x + ", " + y + "] = " + color);
    }*/
   image.setRGB(x, y, color.getRGB());
  }
  

  def finish(image:BufferedImage) =
    image;
}

object AwtSpriteReader {

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
        val c = Colors.colors(i);
        pal(i) = new Color(c.getRed(), c.getGreen(), c.getBlue());
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
      var pal = new Array[Color](Colors.colors.length);
      for (i <- 0 until pal.length) {
        val c = Colors.colors(i);
        pal(i) = new Color(c.getRed(), c.getGreen(), c.getBlue());
      }
      pal;
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
                target:BufferedImage,
                colorFilter:(Int)=>Int):Unit = {

    var index = 0;
    if (data(index) == 1 && bitmapNumber > 0) {
      AwtSpriteReaderHelper.reset(target);
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
        //SpriteReaderHelper.numberColors(palette);
        
        while (n_r < next_ctl) {
          while ((n_r < (index + long_data)) && (x < header.width)) {
            if (data(n_r) >= rle_value) {
              last_pos = n_r + 1;
              var rle_length = data(n_r) - rle_value + 1;
              var rle_counter = 0;
              while ((rle_counter < rle_length) && (y < header.height)) {
                if ((x < header.width) && (y < header.height) && (x >= 0) && (y >= 0)) {
                  AwtSpriteReaderHelper.withPixelDo(target, x, y, palette(colorFilter(data(last_pos))));
                } else {
                  throw new Exception("Overrun");

                }
                y += 1;
                rle_counter += 1;
              }
              n_r += 2;
            } else {
              if ((x < header.width) && (y < header.height) && (x >= 0) && (y >= 0)) {
                AwtSpriteReaderHelper.withPixelDo(target, x, y, palette(colorFilter(data(n_r))));
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

  def buildColorFilter(flag:Option[FlagColor]) = {
    flag match {
      case Some(flagColor) => {
          (x:Int) => if (x >= 214 && x <= 218) {
            x + flagColor.paletteIndexOffset;
          } else {
            x
          }
      }
      case None => {
          (x:Int) => x;
      }
    }
  }

  def read(lbxReader:LbxReader, groupIndex:Int, flag:Option[FlagColor]) = {

    val lbxMetaData = lbxReader.metaData
    
    val header = readHeader(lbxReader, groupIndex)
    
    val offsets = readOffsets(lbxReader, header.bitmapCount)
    val paletteInfo = readPaletteInfo(lbxReader, groupIndex, header.paletteInfoOffset)
    val palette = readPalette(lbxReader, groupIndex, header, paletteInfo)

    var canvas = AwtSpriteReaderHelper.create(header.width, header.height);
    AwtSpriteReaderHelper.reset(canvas);
    def readSprite(bitmapNumber:Int, offset:Offset) = {
      lbxReader.seek(lbxMetaData.subfileStart(groupIndex) + offset.start);
      var data = new Array[Byte](offset.end - offset.start);
      lbxReader.read(data);
      // make it unsigned
      var data2 = new Array[Int](data.length);
      for (i <- 0 until data.length) {
        data2(i) = data(i) & 0xFF;
      }

      
      render(bitmapNumber, data2, header, paletteInfo, palette, canvas, buildColorFilter(flag));
      AwtSpriteReaderHelper.copy(canvas);
    }

    val answer = new Array[Image](offsets.size);
    for (index <- 0 until offsets.size) {
      answer(index) = AwtSpriteReaderHelper.finish(
        readSprite(index, offsets(index)));
    }

    answer;
  }
}

case class SpriteGroupKey(originalGameAsset:OriginalGameAsset, group:Int, flag:Option[FlagColor]);

class AwtImageLibrarian {
  
  val spriteGroupCache = CacheBuilder
    .newBuilder()
    .maximumSize(256)
    .removalListener(new RemovalListener[SpriteGroupKey, Array[Image]]() {
      def onRemoval(removal:RemovalNotification[SpriteGroupKey, Array[Image]]):Unit = {
        /*val images = removal.getValue();
        for (image <- images) 
          image.destroy();      */
      }
    })
    .build(new CacheLoader[SpriteGroupKey, Array[Image]](){
      def load(key:SpriteGroupKey):Array[Image] = {
        val reader = new LbxReader(Data.originalDataPath(key.originalGameAsset.fileName));
        val sprites = AwtSpriteReader.read(reader, key.group, key.flag);
        reader.close();
        return sprites;
      }
    });

  def getRawSprite(originalGameAsset:OriginalGameAsset, group:Int, index:Int) = {
    val images = spriteGroupCache.get(new SpriteGroupKey(originalGameAsset, group, None));
    images(index);
  }

  def getFlaggedSprite(originalGameAsset:OriginalGameAsset, group:Int, index:Int, flag:FlagColor) = {
    val images = spriteGroupCache.get(new SpriteGroupKey(originalGameAsset, group, Some(flag)));
    images(index);
  }
}