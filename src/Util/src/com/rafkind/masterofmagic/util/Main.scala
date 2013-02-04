package com.rafkind.masterofmagic.util;

import com.rafkind.masterofmagic.util.terrain.FancyMetadataEditor;
import com.rafkind.masterofmagic.util.terrain.TerrainMetadataEditor;
import com.rafkind.masterofmagic.util.terrain.VisualMetadataEditor;
import com.rafkind.masterofmagic.util.sprite.SpriteBrowser;

object Main {

  /**
   * @param args the command line arguments
   */
  def main(args: Array[String]): Unit = {
    //TerrainMetadataEditor.main(args);
    //FancyMetadataEditor.main(args);
    VisualMetadataEditor.main(args);
    //SpriteBrowser.main(args);
  }
}