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

    for (i <- 0 until 24) {
      val x = reader.read2();
      println("%02d: %04X".format(i, x));
    }

    for (f <- 0 until 8) {
      for (c <- 32 until 128) {
        val x = reader.read();
        println("Font %d: char %d '%c': %d".format(f, c, c, x));
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

  for (f <- 0 to 7){
    for (c <- 32 until 128) {
      /*
      val rle = reader.read()
      val color = reader.read();
      println("Font %d Rle %d color %d".format(f, rle, color));
      */
      val offset = reader.read2()
      println("Font %d Glyph %d Offset %x".format(f, c, offset))
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
