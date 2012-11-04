package com.rafkind.masterofmagic

import org.newdawn.slick.AppGameContainer;
import org.newdawn.slick.ScalableGame;
import org.newdawn.slick.Game;
import com.rafkind.masterofmagic.util.OriginalGameAsset;
import com.rafkind.masterofmagic.system.Data;
import com.rafkind.masterofmagic.util.ImageLibrarian

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

  @Provides def provideFontManager() = {
    import com.rafkind.masterofmagic.ui.framework.FontManager;
    

    val fontManager =
      new FontManager(
        Data.originalDataPath(
          OriginalGameAsset.FONTS.fileName));

    fontManager;
  }

  @Provides def provideTerrainPainter() = {
    import com.rafkind.masterofmagic.ui.main.TerrainPainter;
    import com.rafkind.masterofmagic.util.TerrainLbxReader;
    
    new TerrainPainter(
      TerrainLbxReader.read(
        Data.originalDataPath(
          OriginalGameAsset.TERRAIN.fileName)),
        Main.appInjector.getInstance(classOf[ImageLibrarian]))
  }
}

object Main {
  val appInjector = Guice.createInjector(new MainModule());
  
  /**
   * @param args the command line arguments
   */
  def main(args: Array[String]): Unit = {

    /*for (i <- 0 until 256) {
      val color = com.rafkind.masterofmagic.util.Colors.colors(i);
      val code = "%02X".format(color.getRed()) + "%02X".format(color.getGreen()) + "%02X".format(color.getBlue())

      println("<tr><td>" + i + "</td><td style='background-color: #" + code + "'>" + i + "</td></tr>");
    }*/

    val app = appInjector.getInstance(classOf[AppGameContainer]);
    app.start();   
  }
}
