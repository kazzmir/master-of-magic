/* Contains some global information about the state of the system
 * we are running on.
 */

package com.rafkind.masterofmagic.system

import java.util.Properties
import java.util.zip._
import java.io._

/* maybe try to hide this class? */
class Data{
  val properties:Properties = Data.loadProperties("game.properties")

  /* Get the 'data' property. Defaults to the 'data' directory */
  def getDataPath() = properties.getProperty("data", "data")

  /* Get the 'originaldata' property. Defaults to the 'data' directory */
  def getOriginalDataPath() = properties.getProperty("originaldata", "data");

  def originalPathIsZip() = getOriginalDataPath().toUpperCase().endsWith(".ZIP");

  /* FIXME: probably use some Path class to join paths together */
  def getPath(user:String) = getDataPath() + java.io.File.separator + user

  /* FIXME: probably use some Path class to join paths together */
  def getOriginalPath(newPath:String):String = {
    var answer = getOriginalDataPath() + java.io.File.separator + newPath;

    if (originalPathIsZip()) {
      val tmpDirectory = new File(System.getProperty("java.io.tmpdir"), "master-of-magic")
      if (tmpDirectory.exists()){
        /* Possibly dangerous, maybe that was a file the user wanted? */
        if (!tmpDirectory.isDirectory()){
          tmpDirectory.delete()
          tmpDirectory.mkdirs()
        } else {
          /* all good */
        }
      } else {
        tmpDirectory.mkdirs()
      }
      val targetFile = new File(tmpDirectory, newPath);

      if (!targetFile.exists()) {
        val zfile = new ZipFile(getOriginalDataPath());
        val e = zfile.entries
        var entry = e.nextElement
        val sb = new StringBuilder();
        var targetEntry:ZipEntry = null;
        while (e.hasMoreElements()) {
          if (entry.getName().toUpperCase().endsWith(newPath.toUpperCase())) {
            targetEntry = entry;
          }
          entry = e.nextElement
        }

        if (targetEntry != null) {
          val out = new FileOutputStream(targetFile);
          val in = zfile.getInputStream(targetEntry);

          val buf = new Array[Byte](0x1000);
          var count = in.read(buf);
          while (count >= 0) {
            out.write(buf, 0, count);
            count = in.read(buf);
          }
          in.close();
          out.close();
        }
        zfile.close();
      }

      answer = targetFile.getCanonicalPath();
    }
    return answer;
  }
}

object Data{

  var dataObject: Data = null

  /* Lazily instantiates a Data object which contains various information */
  def getData():Data = {
    if (dataObject == null){
      dataObject = new Data()
    }
    dataObject
  }

  /* Converts a relative path to an absolute one using the data path
   * that is currently configured.
   */
  def path(local:String) = getData().getPath(local)

  def originalDataPath(local:String) = getData().getOriginalPath(local);

  /* Load a Properties object from a file given as a path on the filesystem
   * FIXME: this is just a utility method. move it elsewhere?
   */ 
  def loadProperties(path:String):Properties = {
     val properties = new Properties()
     properties.load(new FileInputStream(new File(path)))
     properties
  }
}
