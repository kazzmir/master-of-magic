(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to load and display a bitmap
 *  file.  You have to use this example from the command line to
 *  specify as first parameter a graphic file in one of Allegro's
 *  supported formats.  If the file is loaded successfully,
 *  it will be displayed until you press a key.
 *)

open Allegro ;;

let () =
  allegro_init();

  if Array.length Sys.argv <> 2 then begin
    allegro_message "Usage: 'exbitmap filename.[bmp|lbm|pcx|tga]'\n";
    exit 1;
  end;

  install_keyboard();

  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 640 480 0 0
    with _ ->
      try set_gfx_mode GFX_SAFE 320 200 0 0
      with _ ->
        set_gfx_mode GFX_TEXT 0 0 0 0;
        allegro_message("Unable to set any graphic mode\n"^ (get_allegro_error()) ^"\n");
        exit 1;
  end;

  let the_palette = new_palette() in

  (* read in the bitmap file *)
  let the_image =
    try load_bitmap Sys.argv.(1) the_palette
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error reading bitmap file '"^ Sys.argv.(1) ^"'\n");
      exit 1;
  in

  (* select the bitmap palette *)
  set_palette the_palette;

  let screen_w, screen_h = get_screen_width(), get_screen_height() in
  let the_image_w, the_image_h = get_bitmap_width the_image, get_bitmap_height the_image in
  let screen = get_screen() in

  (* blit the image onto the screen *)
  blit the_image  screen  0  0 
      ((screen_w - the_image_w)/2)
      ((screen_h - the_image_h)/2) the_image_w the_image_h;

  (* destroy the bitmap *)
  destroy_bitmap the_image;

  ignore(readkey());
;;

