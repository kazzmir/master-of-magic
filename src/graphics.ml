type windowModes = Windowed | FullScreen;;

module type GraphicsSignature = sig
  class graphics: object
    method initialize: windowModes -> int -> int -> unit
  end
end;;

module AllegroGraphics: GraphicsSignature = struct
  class graphics = object(self)
    method initialize (mode : windowModes) width height =
      Allegro.allegro_init ();
      Allegro.set_color_depth 16;
      match mode with
      | Windowed -> Allegro.set_gfx_mode Allegro.GFX_AUTODETECT_WINDOWED width
      height 0 0;
      | FullScreen -> Allegro.set_gfx_mode Allegro.GFX_AUTODETECT_FULLSCREEN
      width height 0 0;
      Allegro.install_timer ();
      Allegro.install_keyboard ()
      (*
      Printf.printf "initialize! %d %d\n"
      width height
      *)
  end
end;;
