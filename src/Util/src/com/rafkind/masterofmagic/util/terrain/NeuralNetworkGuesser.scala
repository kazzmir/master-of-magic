package com.rafkind.masterofmagic.util.terrain

import com.rafkind.masterofmagic.util.TerrainLbxReader;

import com.rafkind.masterofmagic.system._;

object NeuralNetworkGuesser {
  def guessTerrain(metadataGuess:Array[EditableTerrainTileMetadata]):Unit = {
    val tileData = Array.ofDim[Int](
      TerrainLbxReader.TILE_COUNT,
      TerrainLbxReader.TILE_HEIGHT,
      TerrainLbxReader.TILE_WIDTH);

    TerrainLbxReader.readAnd(Data.originalDataPath("TERRAIN.LBX"),
      (tile, x, y, color) => {
        tileData(tile)(y)(x) = color;
      });
  }
}
