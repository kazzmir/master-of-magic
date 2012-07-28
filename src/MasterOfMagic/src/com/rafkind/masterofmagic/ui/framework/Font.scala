/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.framework

import com.google.inject._;
import org.newdawn.slick._;
import com.rafkind.masterofmagic.util._;

case class FontIdentifier(val id:Int, val name:String);
object FontIdentifier {
  val TINY = FontIdentifier(0, "Tiny");
  val SMALL = FontIdentifier(1, "Small");
  val MEDIUM = FontIdentifier(2, "Medium");
  val LARGE = FontIdentifier(3, "Large");
  val HUGE = FontIdentifier(4, "Huge");
  val FANCY = FontIdentifier(5, "Fancy");
  val ALIEN_SMALL = FontIdentifier(6, "Alien Small");
  val ALIEN_LARGE = FontIdentifier(7, "Alien Large");

  val values = Array(TINY, SMALL, MEDIUM, LARGE, HUGE, FANCY, ALIEN_SMALL, ALIEN_LARGE);
}

object GlyphTemplate {
  def build(fontIdentifier:FontIdentifier,
            glyphIndex:Int,
            width:Int,
            height:Int,
            rleData:Seq[Int]):GlyphTemplate = {
    val data = new Array[Int](width * height);

    var x = 0;
    var y = 0;
    for (value <- rleData) {
      if (value == 0x80) {
        x = 0;
        y += 1;
      } else if (value > 0x80) {
        var count = value - 0x80;
        x += count;
      } else {
        val high = (value >> 4);
        val low = (value & 0xF);
        for (q <- 0 until high) {
          // put flipped pixel
          val px = y;
          val py = x;
          // need this guard due to bad data
          if ((px < width) && (py < height))
            data((py * width) + px) = low; 
          
          x + 1;
          if (x >= height) { // flippage
            y += 1;
            x = 0;
          }
        }
      }
      if (x >= height) { // flippage
        y += 1;
        x = 0;
      }
    }

    new GlyphTemplate(fontIdentifier,
                      glyphIndex,
                      (32 + glyphIndex).toChar,
                      width,
                      height,
                      data);
  }
}

case class GlyphTemplate(fontIdentifier:FontIdentifier,
                    glyphIndex:Int,
                    char:Character,
                    width:Int,
                    height:Int,
                    data:Array[Int]) {

  def render(target:ImageBuffer,
             offsetX:Int,
             offsetY:Int,
             palette:Array[Color]):Unit = {
    // here, color 0 is transparent
    for (y <- 0 until height) {
      for (x <- 0 until width) {
        val pixelValue = data(x + (y * width));
        if (pixelValue > 0) {
          val color = palette(pixelValue-1);
          target.setRGBA(x + offsetX,
                         y + offsetY,
                         color.getRed(),
                         color.getGreen(),
                         color.getBlue(),
                         color.getAlpha());
        }
       //target.setRGBA(x + offsetX, y + offsetY, 255, 255, 255, 255);
      }
    }
  }
}

case class Font(fontIdentifier:FontIdentifier, height:Int, glyphTemplates:Array[GlyphTemplate]) {
  def getWidthOf(text:String):Int = {
    var running = 0;
    for (c <- text) {
      val glyph = glyphTemplates(c - 32);
      running += glyph.width + 1;
    }
    running;
  }

  def walkLayout(width:Int,
                 height:Int,
                 text:String,
                 callback:(Int, Int, GlyphTemplate) => Unit):Unit = {
    var running = 0;
    for (c <- text) {
      val glyph = glyphTemplates(c - 32);
      callback(running, 0, glyph);
      running += glyph.width + 1;
    }
  }

  def render(imageBuffer:ImageBuffer, 
             offsetX:Int,
             offsetY:Int,
             palette:Array[Color],
             text:String):Unit = {
    walkLayout(
      imageBuffer.getWidth(),
      imageBuffer.getHeight(),
      text,
      (x:Int, y:Int, glyph:GlyphTemplate) => {
        glyph.render(imageBuffer, offsetX + x, offsetY + y, palette);
      });
  }
}

@Singleton
class FontManager(path:String) {

  val fonts = loadAll(path);

  // from FONTS.LBX subfile 0
  def loadAll(path:String):Array[Font] = {
    val reader = new LbxReader(path);
    val metadata = reader.metaData;

    reader.seek(metadata.subfileStart(0) + 0x16A); // magic constant

    var fontHeights =
      for (i <- 0 until 24) yield reader.read2();
    
    var fontWidths =
      for (i <- 0 until 8) yield {
        for (c <- 32 until 128) yield reader.read();
      }

    val fontOffsets =
      for (f <- 0 until 8) yield {
        for (c <- 32 until 128) yield reader.read2();
      }

    val fonts = new Array[Font](8);
    for (fontIdentifier <- FontIdentifier.values) {

      val glyphs = new Array[GlyphTemplate](96);
      val height = fontHeights(fontIdentifier.id);
      for (c <- 0 until 95) {
        val start = fontOffsets(fontIdentifier.id)(c);
        val end = fontOffsets(fontIdentifier.id)(c+1);

        reader.seek(metadata.subfileStart(0) + start);
        val width = fontWidths(fontIdentifier.id)(c);
        glyphs(c) = GlyphTemplate.build(
          fontIdentifier,
          c,
          width,
          height,
          for (b <- start until end) yield reader.read());
      }

      fonts(fontIdentifier.id) = new Font(fontIdentifier, height, glyphs);
    }

    reader.close();
    fonts;
  }
}