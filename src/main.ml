let init graphics =
  graphics#initialize Graphics.Windowed 640 480
;;

class handler = object(self)
  inherit Graphics.eventHandler
  method mouse_down time button =
    let name = (Graphics.mouse_button_name button) in
    Printf.printf "Mouse down %f %s\n%!" time name
  method mouse_up = Printf.printf "Mouse up\n%!"
  method key_down = ignore()
  method key_up = ignore()
  method mouse_click time button = 
    let name = (Graphics.mouse_button_name button) in
    Printf.printf "Mouse click %f %s\n%!" time name
  method mouse_hover = ignore()
  method mouse_move = ignore()
  method keypress = ignore()
end;;

let create_event_handler graphics = 
  (new handler)
;;

let main () : unit = 
  let graphics = (new Graphics.AllegroGraphics.graphics) in
  let event = create_event_handler () in
  Printf.printf "before graphics\n";
  init graphics;
  Printf.printf "before\n";
  graphics#event_loop event;
  Printf.printf "after\n";
;;

Printf.printf "before main\n";
main ();
