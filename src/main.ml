let init graphics =
  graphics#initialize Graphics.Windowed 640 480
;;

let main () = 
  init (new Graphics.AllegroGraphics.graphics);
  1
;;

main ();
