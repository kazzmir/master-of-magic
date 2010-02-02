(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program shows how to specify colors in the various different
 *  truecolor pixel formats. The example shows the same screen (a few
 *  text lines and three coloured gradients) in all the color depth
 *  modes supported by your video card. The more color depth you have,
 *  the less banding you will see in the gradients.
 *)

open Allegro ;;


let test ~colordepth =
  try
    (* set the screen mode *)
    set_color_depth colordepth;

    begin
      try set_gfx_mode GFX_AUTODETECT_WINDOWED 640 480 0 0
      with x -> raise x
    end;

    (* in case this is a 256 color mode, we'd better make sure that the
     * palette is set to something sensible. This function generates a
     * standard palette with a nice range of different colors...
     *)
    let pal = new_palette() in
    generate_332_palette pal;
    set_palette pal;

    acquire_screen();

    let screen_w, screen_h = get_screen_width(), get_screen_height() in
    let screen, font = get_screen(), get_font() in

    clear_to_color screen (makecol 0 0 0);

    let transp = color_index(-1) in

    textout_ex screen font 
          ((string_of_int colordepth)^" bit color...")
          0 0 (makecol 255 255 255) transp;

    (* use the makecol() function to specify RGB values... *)
    textout_ex screen font "Red"     32 80  (makecol 255 0   0  ) transp;
    textout_ex screen font "Green"   32 100 (makecol 0   255 0  ) transp;
    textout_ex screen font "Blue"    32 120 (makecol 0   0   255) transp;
    textout_ex screen font "Yellow"  32 140 (makecol 255 255 0  ) transp;
    textout_ex screen font "Cyan"    32 160 (makecol 0   255 255) transp;
    textout_ex screen font "Magenta" 32 180 (makecol 255 0   255) transp;
    textout_ex screen font "Grey"    32 200 (makecol 128 128 128) transp;

    (* or we could draw some nice smooth color gradients... *)
    for x=0 to pred 256 do
      vline screen (192+x) 112 176 (makecol x 0 0);
      vline screen (192+x) 208 272 (makecol 0 x 0);
      vline screen (192+x) 304 368 (makecol 0 0 x);
    done;

    textout_centre_ex screen font "<press a key>" (screen_w / 2) (screen_h - 16)
          (makecol 255 255 255) transp;

    release_screen();

    ignore(readkey());

  with Failure "set_gfx_mode" -> ()
;;


let () =
   allegro_init();
   install_keyboard(); 

   (* try each of the possible possible color depths... *)
   test 8;
   test 15;
   test 16;
   test 24;
   test 32;
;;

