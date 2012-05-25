/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui.swing

import java.awt._;
import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.util._;

class ImageLibrarian {

  val terrainTiles = Map(
    TerrainType.OCEAN -> createTerrainTile(Color.BLUE),
    TerrainType.SHORE -> createTerrainTile(Color.MAGENTA),
    TerrainType.RIVER -> createTerrainTile(Color.BLUE),
    TerrainType.SWAMP -> createTerrainTile(Color.GREEN),
    TerrainType.TUNDRA -> createTerrainTile(Color.WHITE),
    TerrainType.DEEP_TUNDRA -> createTerrainTile(Color.WHITE),
    TerrainType.MOUNTAIN -> createTerrainTile(Color.DARK_GRAY),
    TerrainType.VOLCANO -> createTerrainTile(Color.DARK_GRAY),
    TerrainType.CHAOS_NODE -> createTerrainTile(Color.PINK),
    TerrainType.HILLS -> createTerrainTile(Color.ORANGE),
    TerrainType.GRASSLAND -> createTerrainTile(Color.GREEN),
    TerrainType.SORCERY_NODE -> createTerrainTile(Color.CYAN),
    TerrainType.DESERT -> createTerrainTile(Color.YELLOW),
    TerrainType.FOREST -> createTerrainTile(Color.GREEN),
    TerrainType.NATURE_NODE -> createTerrainTile(Color.LIGHT_GRAY));

  def getTerrainTileImage(terrainSquare:TerrainSquare):Image = {
    terrainTiles(terrainSquare.terrainType);
  }

  def createTerrainTile(c:Color):Image = {
    val bi = GraphicsEnvironment
      .getLocalGraphicsEnvironment()
      .getDefaultScreenDevice()
      .getDefaultConfiguration()
      .createCompatibleImage(TerrainLbxReader.TILE_WIDTH,
                             TerrainLbxReader.TILE_HEIGHT);

    val graphics = bi.createGraphics();

    graphics.setColor(c);
    graphics.fill(
      new Rectangle(
        0,
        0,
        TerrainLbxReader.TILE_WIDTH,
        TerrainLbxReader.TILE_HEIGHT));
    bi;
  }
}
