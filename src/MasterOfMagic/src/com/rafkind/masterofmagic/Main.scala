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

    var whichfont = 0;
   for (ch <- 65 until (65+26)) {
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
