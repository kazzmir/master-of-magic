(*  Example program for the Allegro library, by Grzegorz Ludorowski.
 *
 *  This example demonstrates how to use PCX files, palettes and stretch
 *  blits. It loads a PCX file, sets its palette and does some random
 *  stretch_blits. Don't worry - it's VERY slowed down using vsync().
 *)

open Allegro ;;

let () =
  Random.self_init();
  allegro_init();
  install_keyboard();

  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 200 0 0;
    with _ ->
      try set_gfx_mode GFX_SAFE 320 200 0 0;
      with _ ->
        set_gfx_mode GFX_TEXT 0 0 0 0;
        allegro_message("Unable to set any graphic mode\n"^ (get_allegro_error()) ^"\n");
        exit 1;
  end;

  let my_palette = new_palette() in
  let pcx_name = replace_filename Sys.argv.(0) "mysha.pcx" in
  let scr_buffer =
    try load_pcx pcx_name my_palette;
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error loading "^ pcx_name);
      exit 1;
  in

  set_palette my_palette;

  let screen = get_screen() in
  let scr_buffer_w, scr_buffer_h = (get_bitmap_width scr_buffer), (get_bitmap_height scr_buffer) in
  let screen_w, screen_h = get_screen_width(), get_screen_height() in

  blit scr_buffer screen 0 0 0 0 scr_buffer_w scr_buffer_h;

  let rand = Random.int in

  while not(keypressed()) do
     stretch_blit scr_buffer  screen  0 0
           (rand scr_buffer_w) (rand scr_buffer_h)
           (rand screen_w)     (rand screen_h)
       	   (rand screen_w)     (rand screen_h);
     vsync();
  done;

  destroy_bitmap scr_buffer;
;;

