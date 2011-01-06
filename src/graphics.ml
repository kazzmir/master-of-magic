type windowModes = Windowed | FullScreen;;

class virtual eventHandler = object
  method virtual mouse_down: unit
  method virtual mouse_up: unit
  method virtual key_down: unit
  method virtual key_up: unit
  method virtual mouse_click: unit
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

    method private mouse_click () =
      (Allegro.left_button_pressed ())

    method event_loop (handler : eventHandler) =
      let rec loop () : unit =
        if (self#mouse_click ()) then begin
          (handler#mouse_click)
        end else
          ();
          (*
        if Allegro.keypressed () then
          Printf.printf "keypress\n%!"
        else
          ();
          *)
        Allegro.rest 10;
        (loop ())
      in
      (loop ())
  end
end;;
