package com.rafkind.masterofmagic.util

import java.io._;
import scala.collection.mutable._;

// format from here:
// http://www.roughseas.ca/momime/phpBB3/viewtopic.php?f=1&t=3

class Lbx(val subfileCount:Int, val magicNumber:Int, val version:Int) {
  var size:Int = 0;
  var subfiles:Map[Int, Int] = new HashMap[Int, Int];

  def addSubfile(index:Int, offset:Int):Map[Int, Int] = {
    subfiles += index -> offset;

    subfiles;
  }

  def setSize(offset:Int):Lbx = {
    size = offset;
    this;
  }
}

object LbxReader {
  def read2(f:RandomAccessFile):Int = {
    var a = f.read();
    var b = f.read();
    return a | (b << 8);
  }

  def read4(f:RandomAccessFile):Int = {
    var a = f.read();
    var b = f.read();
    var c = f.read();
    var d = f.read();
    return a | (b << 8) | (c << 16) | (d << 24);
  }

  def read(fileName:String):Lbx = {
    var lbxFile = new RandomAccessFile(new File(fileName), "r");
    lbxFile.seek(0);

    var subfileCount = read2(lbxFile);
    var magicNumber = read4(lbxFile);
    var version = read2(lbxFile);

    var lbx = new Lbx(subfileCount, magicNumber, version);
    
    for (s <- 0 until subfileCount) {
      lbx.addSubfile(s, read4(lbxFile));
    }
    lbx.setSize(read4(lbxFile));
    
    lbxFile.close();
    return lbx;
  }
}
