/* Contains some global information about the state of the system
 * we are running on.
 */

package com.rafkind.masterofmagic.system

import java.util.Properties
import java.io._

/* maybe try to hide this class? */
class Data{
  val properties:Properties = Data.loadProperties("../../game.properties")

  /* Get the 'data' property. Defaults to the 'data' directory */
  def getDataPath() = properties.getProperty("data", "data")

  /* FIXME: probably use some Path class to join paths together */
  def getPath(user:String) = getDataPath() + "/" + user
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

  /* Load a Properties object from a file given as a path on the filesystem
   * FIXME: this is just a utility method. move it elsewhere?
   */ 
  def loadProperties(path:String):Properties = {
     val properties = new Properties()
     properties.load(new FileInputStream(new File(path)))
     properties
  }
}
