(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to use hardware scrolling.
 *  The scrolling should work on anything that supports virtual
 *  screens larger than the physical screen.
 *)
open Allegro ;;

let () =
  Random.self_init();
  allegro_init();
  install_keyboard();

  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 240 640 240;
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Unable to set a 320x240 mode with 640x240 " ^
       	       "virtual dimensions\n");
      exit 1;
  end;

  let screen = get_screen() in
  let screen_w, screen_h = get_screen_width(), get_screen_height() in

  (* the scrolling area is twice the width of the screen (640x240) *)
  let scroller = create_sub_bitmap screen 0 0 (screen_w*2) screen_h in

  set_palette (get_desktop_palette());
  set_color 0  0 0 0 0;

  rectfill scroller 0 0 screen_w 100 (color_index 6);
  rectfill scroller 0 100 screen_w screen_h (color_index 2);

  let rec loop x h =

    (* advance the scroller, wrapping every 320 pixels *)
    let next_x = x + 1 in
    let next_x =
      if (next_x >= 320) then 0 else next_x
    in

    (* draw another column of the landscape *)
    acquire_bitmap scroller;
    vline scroller (next_x+screen_w-1) 0 h (color_index 6);
    vline scroller (next_x+screen_w-1) (h+1) screen_h (color_index 2);
    release_bitmap scroller;

    (* scroll the screen *)
    scroll_screen next_x 0;

    (* duplicate the landscape column so we can wrap the scroller *)
    if next_x > 0 then begin
      acquire_bitmap scroller;
      vline scroller x 0 h (color_index 6);
      vline scroller x (h+1) screen_h (color_index 2);
      release_bitmap scroller;
    end;

    (* randomly alter the landscape position *)
    let h =
      if Random.bool()
      then (if (h > 5) then h - 1 else h)
      else (if (h < 195) then h + 1 else h)
    in

    if not(keypressed()) then loop next_x h
  in
  loop 0 100;

  destroy_bitmap scroller;
  clear_keybuf();
;;

