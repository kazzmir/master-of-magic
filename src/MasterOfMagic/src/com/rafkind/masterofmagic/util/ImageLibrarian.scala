/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util
import org.newdawn.slick._;
import com.google.common.cache._;
import com.rafkind.masterofmagic.system._;

import com.google.inject._;

case class SpriteGroupKey(originalGameAsset:OriginalGameAsset, group:Int);

@Singleton
class ImageLibrarian {
  val spriteGroupCache = CacheBuilder
    .newBuilder()
    .maximumSize(256)
    .removalListener(new RemovalListener[SpriteGroupKey, Array[Image]]() {
      def onRemoval(removal:RemovalNotification[SpriteGroupKey, Array[Image]]):Unit = {
        val images = removal.getValue();
        for (image <- images) 
          image.destroy();        
      }
    })
    .build(new CacheLoader[SpriteGroupKey, Array[Image]](){
      def load(key:SpriteGroupKey):Array[Image] = {
        val reader = new LbxReader(Data.originalDataPath(key.originalGameAsset.fileName));
        val sprites = SpriteReader.read(reader, key.group);
        reader.close();
        return sprites;
      }
    });

  def getRawSprite(originalGameAsset:OriginalGameAsset, group:Int, index:Int) = {
    val images = spriteGroupCache.get(new SpriteGroupKey(originalGameAsset, group));
    images(index);
  }
}