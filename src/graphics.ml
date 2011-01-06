type windowModes = Windowed | FullScreen;;
type mouseButton = MouseButtonLeft | MouseButtonRight

let mouse_button_name button =
  match button with
  | MouseButtonLeft -> "left button"
  | MouseButtonRight -> "right button"
;;

class virtual eventHandler = object
  method virtual mouse_down: float -> mouseButton -> unit
  method virtual mouse_up: unit
  method virtual key_down: unit
  method virtual key_up: unit
  method virtual mouse_click: float -> mouseButton -> unit
  method virtual mouse_hover: unit
  method virtual mouse_move: unit
  method virtual keypress: unit
end;;

module type GraphicsSignature = sig
  class graphics: object
    method initialize: windowModes -> int -> int -> unit
    method event_loop: eventHandler -> unit
  end
end;;

type mouseButtonState = {held : int; last_pressed_time : float;}
type eventLoopState = {mouse_left_button : mouseButtonState;
                       mouse_right_button : mouseButtonState;}

module AllegroGraphics: GraphicsSignature = struct
  class graphics = object(self)
    method initialize (mode : windowModes) width height =
      Allegro.allegro_init ();
      Allegro.set_color_depth 16;
      Allegro.install_timer ();
      Allegro.install_keyboard ();
      ignore(Allegro.install_mouse ());
      match mode with
      | Windowed -> Allegro.set_gfx_mode Allegro.GFX_AUTODETECT_WINDOWED width
      height 0 0;
      | FullScreen -> Allegro.set_gfx_mode Allegro.GFX_AUTODETECT_FULLSCREEN
      width height 0 0
      (*
      Printf.printf "initialize! %d %d\n"
      width height
      *)

    method private mouse_left_down () =
      (Allegro.left_button_pressed ())
    
    method private mouse_right_down () =
      (Allegro.right_button_pressed ())

    method event_loop (handler : eventHandler) =
      let mouse_delta = 100.0 /. 1000.0 in
      let check_mouse_button checker button state time =
        match checker () with
        | true -> begin match state.held with
                        | 0 -> (handler#mouse_down time button); {state with held 
                        = 1; last_pressed_time = time}
                        | n -> state
                  end
        | false -> begin match state.held with
                         | 0 -> {state with held = 0; last_pressed_time = 0.0}
                         | n -> (handler#mouse_up);
                                if time -. state.last_pressed_time < mouse_delta then
                                  (handler#mouse_click time button)
                                else
                                  ();
                                  {state with held = 0; last_pressed_time = 0.0}
                         end
      in
      let check_right state time = 
        {state with mouse_right_button = check_mouse_button (fun () -> self#mouse_right_down ()) MouseButtonRight state.mouse_right_button time}
      in
      let check_left state time =
        {state with mouse_left_button = check_mouse_button (fun () -> self#mouse_left_down ()) MouseButtonLeft state.mouse_left_button time}
      in
      let rec loop state : unit =
        Allegro.rest 10;
        let new_time = Unix.gettimeofday () in
        let checks = [check_left; check_right] in
        loop (List.fold_left (fun old_state check -> check old_state new_time) state 
        checks)
      in
      (loop {mouse_left_button = {held = 0; last_pressed_time = 0.0};
             mouse_right_button = {held = 0; last_pressed_time = 0.0}});
  end
end;;
