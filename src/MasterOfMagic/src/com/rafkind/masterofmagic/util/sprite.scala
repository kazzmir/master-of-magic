package com.rafkind.masterofmagic.util

object SpriteReader{
  def read(lbxReader:LbxReader, index:Int){
    val lbx = lbxReader.readLbx()
    val data = lbxReader.read(lbx.subfile(index))
  }
}
