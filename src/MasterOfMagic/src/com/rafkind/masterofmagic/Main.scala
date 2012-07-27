package com.rafkind.masterofmagic

import org.newdawn.slick.AppGameContainer;
import org.newdawn.slick.ScalableGame;
import org.newdawn.slick.Game;


import com.google.inject._;

class MainModule extends AbstractModule {
  override def configure():Unit = {
    bind(classOf[Game]).to(classOf[MasterOfMagic]);
  }

  @Provides def provideAppGameContainer(game:Game) = {
    val app = new AppGameContainer(
      new ScalableGame(game, 320, 200));
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(960, 600, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app;
  }
}

object Main {

  def fontViewer(){
    import org.newdawn.slick.GameContainer;
    import org.newdawn.slick.Graphics
    import org.newdawn.slick.Color
    import com.rafkind.masterofmagic.util._;
    import com.rafkind.masterofmagic.system._;
    class FontGame extends org.newdawn.slick.BasicGame("Fonts") {

      class Glyph(val character:Int, width:Int, height:Int, data:Seq[Int]){
        def render(startX:Int, startY:Int, graphics:Graphics){
          var x = 0;
          var y = 0;
          // println("Render %d %c".format(character, character + 32))
          graphics.setColor(Color.white)

          for (value <- data){
            if (value == 0x80) {
              // println();
              x = 0;
              y += 1;
            } else if (value > 0x80) {
              val count = value - 0x80;
              /*
              for (i <- 0 until count){
                print("  ");
              }
              */
                x += count;
              } else {
                val high = (value >> 4);
                val low = (value & 0xF);
                //print("%02X".format(data));
                //x += 1;
                for (q <- 0 until high) {
                  graphics.fillOval(startX + x, startY + y, 1, 1)
                  // print("%X%X".format(low, low));
                  x += 1;
                  if (x >= width) {
                    // println();
                    y += 1;
                    x = 0;
                  }
                }
            }
            if (x >= width) {
              // println();
              y += 1;
              x = 0;
            }
           //print("%02X ".format(data))
          }
        }
      }

      def loadFont(font:Int):Seq[Glyph] = {
        val reader = new LbxReader(Data.originalDataPath("FONTS.LBX"));
        val metadata = reader.metaData;
        val data = reader.read(metadata.subfile(0));
        reader.seek(metadata.subfileStart(0) + 0x16a);
        var fontHeights = for (i <- 0 until 24) yield reader.read2()
        var fontWidths =
          for (f <- 0 until 8) yield {
            for (c <- 32 until 128) yield {
              val x = reader.read();
              x
            }
        }

        var fontOffsets =
          for (f <- 0 to 7) yield {
            for (c <- 32 until 128) yield reader.read2()
          }

        val glyphs = for (ch <- 0 until 90) yield {
          val start = fontOffsets(font)(ch);
          val end = fontOffsets(font)(ch + 1);
          reader.seek(metadata.subfileStart(0) + start);
          /* Read N bytes where N = end - start */
          var h = fontWidths(font)(ch);
          val w = fontHeights(font);
          new Glyph(ch, w, h, for (byte <- start to end) yield reader.read())
        }

        glyphs
      }

      val fonts = for (font <- 0 until 8) yield loadFont(font)
      var currentGlyph = 65
      var currentFont = 0

      override def keyPressed(key:Int, char:Char){
        val left = 203
        val right = 205
        val up = 200
        val down = 208

        /*
        key match {
          case left => currentGlyph = currentGlyph - 1
          case right => currentGlyph = currentGlyph + 1
        }
        */
        if (key == left){
          currentGlyph -= 1
        }
        if (key == right){
          currentGlyph += 1
        }
        if (key == up){
          currentFont -= 1
        }
        if (key == down){
          currentFont += 1
        }

        currentFont = (currentFont + fonts.length) % fonts.length
        currentGlyph = (currentGlyph + fonts(currentFont).length) % fonts(currentFont).length

        val glyph = fonts(currentFont)(currentGlyph)
        println("Font %d render %d %c".format(currentFont, glyph.character, glyph.character))
      }

      def update(container:GameContainer, delta:Int){
      }

      def init(container:GameContainer){
      }

      def render(container:GameContainer, graphics:Graphics){
        graphics.scale(3, 3)
        fonts(currentFont)(currentGlyph).render(10, 10, graphics)
      }
    }

    val app = new AppGameContainer(new ScalableGame(new FontGame(), 320, 200));
    org.lwjgl.input.Keyboard.enableRepeatEvents(true)
    app.setDisplayMode(960, 600, false)
    app.setSmoothDeltas(true)
    app.setTargetFrameRate(40)
    app.setShowFPS(false)
    app.start()
  }

  /**
   * @param args the command line arguments
   */
  def mainX(args: Array[String]): Unit = {
    val injector = Guice.createInjector(new MainModule());
    val app = injector.getInstance(classOf[AppGameContainer]);
    app.start();   
  }

  import com.rafkind.masterofmagic.util._;
  import com.rafkind.masterofmagic.system._;
  // http://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=91
  def main(args:Array[String]):Unit = {
    fontViewer()

    val reader = new LbxReader(Data.originalDataPath("FONTS.LBX"));
    val metadata = reader.metaData;

    val data = reader.read(metadata.subfile(0));

    reader.seek(metadata.subfileStart(0) + 0x16a);

    var fontHeights =
    for (i <- 0 until 24) yield {
      val x = reader.read2();
      x;
    }

    var fontWidths =
    for (f <- 0 until 8) yield {
      for (c <- 32 until 128) yield {
        val x = reader.read();
        x
      }
    }

   /*
   for (f <- 0 until 4) {
      for (c <- 32 until 128) {
        val offset = reader.read2()
        println("Font %d: char %d '%c': %d".format(f, c, c, offset));
      }
    }
    */
    /*
   for (color <- 0 until 256){
     val r = reader.read()
     val g = reader.read()
     val b = reader.read()
     println("Color %d red %d green %d blue %d".format(color, r, g, b))
   }
   */

    var fontOffsets =
  for (f <- 0 to 7) yield {
    for (c <- 32 until 128) yield {
      /*
      val rle = reader.read()
      val color = reader.read();
      println("Font %d Rle %d color %d".format(f, rle, color));
      */
      val off = reader.read2()
      println("Font %d Glyph %d Offset %x".format(f, c, off))
      off
    }
  }


    /**for (i <- 0 until (fontOffsets(5).length-1)) {
      val char = i + 32;
      val start = fontOffsets(5)(i);
      val end = fontOffsets(5)(i+1);
      println("Font %d char %d '%c' from %d to %d".format(5, char, char, start, end));
      reader.seek(metadata.subfileStart(0) + start);
      for (j <- 0 until (end-start)) {
        val data = reader.read();
        print("%02X ".format(data));
      }
      println();
    }*/

    var whichfont = 5;
   for (ch <- 65 until (65+65)) {
   var x = 0;
   var y = 0;
   var h = fontWidths(whichfont)(ch-32);
   val w = fontHeights(whichfont);

     println("%d is %c: %d x %d".format(ch, ch, w, h));
   val start = fontOffsets(whichfont)(ch-32);
   val end = fontOffsets(whichfont)(ch-32 + 1);
   reader.seek(metadata.subfileStart(0) + start);
   for (j <- 0 until (end-start)) {
        val data = reader.read();
        if (data == 0x80) {
          println();
          x = 0;
          y += 1;
        } else if (data > 0x80) {
          val count = data - 0x80;
          for (i <- 0 until count)
            print("  ");
          x += count;
        } else {
          val high = (data >> 4);
          val low = (data & 0xF);
          //print("%02X".format(data));
          //x += 1;
          for (q <- 0 until high) {
            print("%X%X".format(low, low));
            x += 1;
            if (x >= w) {
              println();
              y += 1;
              x = 0;
            }
          }
        }
        if (x >= w) {
          println();
          y += 1;
          x = 0;
        }
       //print("%02X ".format(data))
   }
   }
  

    /*for (i <- 0 until data.length) {
      if ((i % 32 == 0)) {
        println();
        print("%04X | ".format(i));
      }

      print("%02X ".format(data(i)));
    }

    for (i <- 0 until data.length) {
      if ((i % 32 == 0)) {
        println();
        print("%04X | ".format(i));
      }

      if ((i % 2 == 0)) {
        val value = (data(i)) + (data(i+1) << 8);
        //if ((value >= 0) && (value <= data.length))
        if (scala.math.abs(0x19a-value) < 50)
          print("%05X ".format(value));
        else {
          print("      ");
        }
      }
    }*/
  }
}
