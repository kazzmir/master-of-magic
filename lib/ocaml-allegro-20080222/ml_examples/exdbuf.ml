(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates the use of double buffering.
 *  It moves a circle across the screen, first just erasing and
 *  redrawing directly to the screen, then with a double buffer.
 *)

open Allegro ;;

let () =
  allegro_init();
  install_timer();
  install_keyboard();

  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 200 0 0;
    with _ ->
      try set_gfx_mode GFX_SAFE 320 200 0 0
      with _ ->
        set_gfx_mode GFX_TEXT 0 0 0 0;
        allegro_message("Unable to set any graphic mode\n"^ (get_allegro_error()) ^"\n");
        exit 1;
  end;

  set_palette(get_desktop_palette());
  let gfx_driver_name = get_gfx_driver_name() in

  let screen, font = get_screen(), get_font() in
  let screen_w, screen_h = get_screen_width(), get_screen_height() in

  (* allocate the memory buffer *)
  let buffer = create_bitmap screen_w screen_h in

  (* First without any buffering...
   * Note use of the global retrace_counter to control the speed. We also
   * compensate screen size (GFX_SAFE) with a virtual 320 screen width.
   *)
  clear_keybuf();
  let c = 32 + (retrace_count()) in
  let rec nobuf_loop() =
    acquire_screen();
    clear_to_color screen (makecol 255 255 255);
    circlefill screen (((retrace_count())-c)*screen_w/320) (screen_h/2) 32 (makecol 0 0 0);
    textout_ex screen font ("No buffering ("^ gfx_driver_name ^")") 0 0 (makecol 0 0 0) (color_index(-1));
    release_screen();

    if not(keypressed()) && (retrace_count())-c <= 320+32 then nobuf_loop()
  in
  nobuf_loop();

  (* and now with a double buffer... *)
  clear_keybuf();
  let c = 32 + (retrace_count()) in
  let rec buf_loop() =
    clear_to_color buffer (makecol 255 255 255);
    circlefill buffer (((retrace_count())-c)*screen_w/320) (screen_h/2) 32 (makecol 0 0 0);
    textout_ex buffer font ("Double buffered ("^ gfx_driver_name ^")") 0 0 (makecol 0 0 0) (color_index(-1));
    blit buffer screen 0 0 0 0 screen_w screen_h;

    if not(keypressed()) && (retrace_count())-c <= 320+32 then buf_loop();
  in
  buf_loop();

  destroy_bitmap buffer;
;;

