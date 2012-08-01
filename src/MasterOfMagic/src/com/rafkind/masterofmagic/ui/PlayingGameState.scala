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

  var text:Image = null;

  def init(container:GameContainer, game:StateBasedGame):Unit = {

    val font = imageLibrarian.getFont(FontIdentifier.HUGE);
    val textib = new ImageBuffer(320, 50);
    val c = new Color(255, 255, 255, 255);
    val colors = Array(c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c, c);

    font.render(textib, 0, 0, colors, "THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG. PACK MY BOX WITH FIVE DOZEN LIQUOR JUGS.");
    text = textib.getImage();

    mainScreen.set(
      Component.BACKGROUND_IMAGE ->
        imageLibrarian.getRawSprite(OriginalGameAsset.MAIN, 0, 0)
    ).add(
      new PackingContainer()
        .set(
          Component.LEFT -> 7,
          Component.TOP -> 4,
          PackingContainer.SPACING -> 1
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
        .add(
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
        )
      );

    
    currentScreen = mainScreen;
  }

  def render(container:GameContainer, game:StateBasedGame, graphics:Graphics):Unit = {
    currentScreen.render(graphics);
    text.draw(0, 0);
  }

  def update(container:GameContainer, game:StateBasedGame, delta:Int):Unit = { }
}
