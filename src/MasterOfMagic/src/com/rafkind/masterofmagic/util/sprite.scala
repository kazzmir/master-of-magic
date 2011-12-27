package com.rafkind.masterofmagic.util

import org.newdawn.slick.ImageBuffer

object SpriteReader{

  case class Header(width:Int, height:Int, unknown1:Int, bitmapCount:Int, unknown2:Int, unknown3:Int, unknown4:Int, paletteInfoOffset:Int,
  unknown5:Int)
  case class PaletteInfo(paletteOffset:Int, firstPaletteColorIndex:Int, count:Int, unknown:Int)
  case class Offset(start:Int, end:Int)

  def readHeader(lbxReader:LbxReader, lbx:Lbx, index:Int):Header = {
    lbxReader.seek(lbx.subfileStart(index))
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

  def readPaletteInfo(lbxReader: LbxReader, lbx:Lbx, index:Int, paletteInfoOffset:Int):PaletteInfo = {
    if (paletteInfoOffset > 0){
      lbxReader.seek(lbx.subfileStart(index) + paletteInfoOffset)
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

  def readPalette(lbxReader:LbxReader, lbx:Lbx, index:Int, paletteOffset:Int){
    /* TODO */
    Colors.colors
  }

  def readOffsets(lbxReader:LbxReader, bitmaps:Int) = {
    val offsets = for (index <- 0 to bitmaps + 1) yield {
      lbxReader.read4()
    }
    for (index <- 0 to bitmaps) yield {
      Offset(offsets(index), offsets(index+1))
    }
  }

  def render(lbxReader:LbxReader, buffer:ImageBuffer, rleValue:Int){
    /* TODO */
  }

  def read(lbxReader:LbxReader, index:Int){
    val lbx = lbxReader.readLbx()
    val header = readHeader(lbxReader, lbx, index)
    val offsets = readOffsets(lbxReader, header.bitmapCount)
    val paletteInfo = readPaletteInfo(lbxReader, lbx, index, header.paletteInfoOffset)
    val palette = readPalette(lbxReader, lbx, index, paletteInfo.paletteOffset)
    val rleValue = paletteInfo.firstPaletteColorIndex + paletteInfo.count

    def readSprite(offset:Offset) = {
      lbxReader.seek(lbx.subfileStart(index) + offset.start)
      val imageBuffer = new ImageBuffer(header.width, header.height)
      render(lbxReader, imageBuffer, rleValue)
      imageBuffer
    }

    val sprites = for (offset <- offsets) yield {
      readSprite(offset)
    }
  }
}
