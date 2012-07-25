/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.ui
import org.newdawn.slick.state._;
import org.newdawn.slick._;

import com.rafkind.masterofmagic.util._;
import com.rafkind.masterofmagic.state._;
import com.rafkind.masterofmagic.ui.framework._;

import com.google.inject._;

class PlayingGameState @Inject() (imageLibrarian:ImageLibrarian, mainScreen:Screen) extends BasicGameState {

  var currentScreen:Screen = null;
  
  def getID() = 1;
  
  def init(container:GameContainer, game:StateBasedGame):Unit = {

    //mainScreen = new Screen(imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 0, 0));
    mainScreen.set(
      Component.BACKGROUND_IMAGE ->
        imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 0, 0)
    ).add(
      new PackingContainer()
        .set(
          Component.LEFT -> 7,
          Component.TOP -> 4
        )
        .add(
          new Button().set(        
            Button.UP_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 1, 0),
            Button.DOWN_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 1, 1)
          ), Alignment.HORIZONTAL
        )
        .add(
          new Button().set(
            Button.UP_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 2, 0),
            Button.DOWN_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 2, 1)
          ), Alignment.HORIZONTAL
        )
        /*.add(
          new Button().set(
            Button.UP_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 3, 0),
            Button.DOWN_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 3, 1)
          ), Alignment.HORIZONTAL
        )
        .add(
          new Button().set(
            Button.UP_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 4, 0),
            Button.DOWN_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 4, 1)
          ), Alignment.HORIZONTAL
        )
        .add(
          new Button().set(
            Button.UP_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 5, 0),
            Button.DOWN_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 5, 1)
          ), Alignment.HORIZONTAL
        )
        .add(
          new Button().set(
            Button.UP_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 6, 0),
            Button.DOWN_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 6, 1)
          ), Alignment.HORIZONTAL
        )
        .add(
          new Button().set(
            Button.UP_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 7, 0),
            Button.DOWN_IMAGE ->
              imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 7, 1)
          ), Alignment.HORIZONTAL
        )*/
      );

    
    currentScreen = mainScreen;
  }

  def render(container:GameContainer, game:StateBasedGame, graphics:Graphics):Unit = {
    currentScreen.render(graphics);
  }

  def update(container:GameContainer, game:StateBasedGame, delta:Int):Unit = { }
}
