let init graphics =
  graphics#initialize Graphics.Windowed 640 480
;;

class handler = object(self)
  inherit Graphics.eventHandler
  method mouse_down time button x y =
    let name = (Graphics.mouse_button_name button) in
    Printf.printf "Mouse down %f %s x %f y %f\n%!" time name x y
  method mouse_up time button x y =
    let name = (Graphics.mouse_button_name button) in
    Printf.printf "Mouse up %f %s x %f y %f\n%!" time name x y
  method key_down key = 
    Printf.printf "Key down %s\n%!" (Graphics.key_name key)
  method key_up key = 
    Printf.printf "Key up %s\n%!" (Graphics.key_name key)
  method mouse_click time button x y =
    let name = (Graphics.mouse_button_name button) in
    Printf.printf "Mouse click %f %s x %f y %f\n%!" time name x y
  method mouse_hover time x y = Printf.printf "Mouse hover %f x %f y %f\n%!" time x y
  method mouse_move time x y = Printf.printf "Mouse move %f x %f y %f\n%!" time x y
  method keypress = ignore()
end;;

let create_event_handler graphics = 
  (new handler)
;;

let main () : unit = 
  let graphics = (new Graphics.AllegroGraphics.graphics) in
  graphics#addTimer 100.0 (fun time -> Printf.printf "Timer called %f\n%!" time);
  let event = create_event_handler () in
  init graphics;
  graphics#event_loop event;
;;

main ();
