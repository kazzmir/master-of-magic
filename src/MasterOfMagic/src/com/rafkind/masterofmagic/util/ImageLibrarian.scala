/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util
import org.newdawn.slick._;
import com.google.common.cache._;
import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.system._;
import com.rafkind.masterofmagic.ui.framework._;

import com.google.inject._;

case class SpriteGroupKey(originalGameAsset:OriginalGameAsset, group:Int, flag:Option[FlagColor]);
case class SpriteKey(originalGameAsset:OriginalGameAsset, group:Int, index:Int);

@Singleton
class ImageLibrarian @Inject() (fontManager:FontManager) {
  
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
        val sprites = SpriteReader.read(reader, key.group, key.flag);
        reader.close();
        return sprites;
      }
    });

  def getRawSprite(originalGameAsset:OriginalGameAsset, group:Int, index:Int) = {
    val images = spriteGroupCache.get(new SpriteGroupKey(originalGameAsset, group, None));
    images(index);
  }
    
  def getFlaggedSprite(originalGameAsset:OriginalGameAsset, group:Int, index:Int, flag:FlagColor) = {
    val images = spriteGroupCache.get(new SpriteGroupKey(originalGameAsset, group, Some(flag)));
    images(index);
  }
  
  def getFlaggedSprite(key:SpriteKey, flag:FlagColor):Image =
    getFlaggedSprite(key.originalGameAsset, key.group, key.index, flag);

  def getFont(fi:FontIdentifier) =
    fontManager.fonts(fi.id);
}