package com.rafkind.masterofmagic

//import javax.swing.SwingUtilities;
//import java.awt._
//import java.awt.image._;
//import javax.swing._;
//import java.awt.geom._;
import org.newdawn.slick.AppGameContainer;
import org.newdawn.slick.ScalableGame;
//import com.rafkind.masterofmagic.util._;


//import com.rafkind.masterofmagic.ui.swing.MainFrame;

object Main {

  /**
   * @param args the command line arguments
   */
  def main(args: Array[String]): Unit = {
    val app = new AppGameContainer(
      new ScalableGame(
        new MasterOfMagic("Master of Magic"), 320, 200));
    org.lwjgl.input.Keyboard.enableRepeatEvents(true);
    app.setDisplayMode(960, 600, false);
    app.setSmoothDeltas(true);
    app.setTargetFrameRate(40);
    app.setShowFPS(false);
    app.start();   
  }

  /*
  def main(args: Array[String]):Unit = {
    SwingUtilities.invokeLater(new Runnable {
      override def run():Unit = {
        val mainFrame = new MainFrame();
        val screenSize = Toolkit.getDefaultToolkit().getScreenSize();
        val WIDTH = 1024;
        val HEIGHT = 768;

        mainFrame.setBounds((screenSize.width - WIDTH)/2, (screenSize.height - HEIGHT)/2, WIDTH, HEIGHT);

        mainFrame.setVisible(true);
      }
    });
  }
  */
 
  /*def main(args:Array[String]):Unit = {
    import java.io._;
    
    val folder = new File("C:/apps/Master of Magic");

    /*for (x <- folder.listFiles(new FilenameFilter {
        override def accept(dir:File, path:String):Boolean = {
          return path.toUpperCase().endsWith(".LBX");
        }
      })) {*/
      val reader = new LbxReader("C:/apps/Master of Magic/DIPLOMAC.LBX");
      val lbx = reader.metaData;

      //println(x + " has " + lbx.subfileCount() + " subs:");
      try {
        for (count <- 0 until lbx.subfileCount() ) {
          val sprites = SpriteReader.read(reader, count,
                                     (w:Int, h:Int) => {
                                      val i = GraphicsEnvironment
                                      .getLocalGraphicsEnvironment()
                                      .getDefaultScreenDevice()
                                      .getDefaultConfiguration()
                                      .createCompatibleImage(w, h);
                                      i;
                                     }, (i:BufferedImage) => {
                                      val g = i.createGraphics();
                                      g.setColor(new Color(255, 255, 255, 0));
                                      g.fill(new Rectangle(0, 0, i.getWidth(), i.getHeight()));
                                     },
                                     (i:BufferedImage) => {
                                      val i2 = GraphicsEnvironment
                                      .getLocalGraphicsEnvironment()
                                      .getDefaultScreenDevice()
                                      .getDefaultConfiguration()
                                      .createCompatibleImage(i.getWidth(), i.getHeight());

                                      i2.createGraphics().drawImage(i, 0, 0, null);
                                      i2
                                     },
                                     (image:BufferedImage, x:Int, y:Int, c:Color) =>
                                       image.setRGB(x, y, c.getRGB()));
          
        for (sprite <- sprites) {
          JOptionPane.showMessageDialog(null, "Hello", "Title", JOptionPane.INFORMATION_MESSAGE, new ImageIcon(sprite));
        }
        
          //println("  " + y + ": " + SpriteReader.read(reader, y).size);
        }
      } catch {
        case e:IOException =>
          println("  " + e);
      } finally {
      }
    //}
  }*/
}
