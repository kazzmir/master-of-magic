(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to get mouse input. The
 *  first part of the test retrieves the raw mouse input data
 *  and displays it on the screen without using any mouse
 *  cursor. When you press a key the standard arrow-like mouse
 *  cursor appears.  You are not restricted to this shape,
 *  and a second key press modifies the cursor to be several
 *  concentric colored circles. They are not joined together,
 *  so you can still see bits of what's behind when you move the
 *  cursor over the printed text message.
 *)

open Allegro ;;


let print_all_buttons() =
  let screen, font = get_screen(), get_font() in
  let fc = makecol 0 0 0 in
  let bc = makecol 255 255 255 in
  textprintf_right_ex screen font 320 50 fc bc "buttons";
  for i = 0 to pred 8 do
    let x = 320 in
    let y = 60 + i * 10 in
    if ((get_mouse_b()) land (1 lsl i)) <> 0
    then textout_right_ex screen font (string_of_int(1 + i)) x y fc bc
    else textout_right_ex screen font "  " x y fc bc;
  done;
;;


let () =
  allegro_init();
  install_keyboard();
  install_timer();

  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 200 0 0
    with _ ->
      try set_gfx_mode GFX_SAFE 320 200 0 0
      with _ ->
        set_gfx_mode GFX_TEXT 0 0 0 0;
        allegro_message("Unable to set any graphic mode\n"^ (get_allegro_error()) ^"\n");
        exit 1;
  end;

  let screen, font = get_screen(), get_font() in
  let screen_w = get_screen_width()
  and screen_h = get_screen_height() in

  set_palette(get_desktop_palette());
  clear_to_color screen (makecol 255 255 255);

  (* Detect mouse presence *)
  let _ =
    try install_mouse()
    with _ ->
      textout_centre_ex screen font "No mouse detected, but you need one!"
          (screen_w/2) (screen_h/2) (makecol 0 0 0) (makecol 255 255 255);
      ignore(readkey());
      exit 0;
  in

  textprintf_centre_ex screen font (screen_w/2) 8 (makecol 0 0 0)
                       (makecol 255 255 255)
                       "Driver: %s" (mouse_driver_name());

  let c = ref 0 in
  let mickeyx = ref 0 in
  let mickeyy = ref 0 in

  while not(keypressed()) do
    (* On most platforms (eg. DOS) things will still work correctly
     * without this call, but it is a good idea to include it in any
     * programs that you want to be portable, because on some platforms
     * you may not be able to get any mouse input without it.
     *)
    poll_mouse();

    acquire_screen();

    let black, white = (makecol 0 0 0), (makecol 255 255 255) in

    (* the mouse position is stored in the variables mouse_x and mouse_y *)
    textprintf_ex screen font 16 48 black white "mouse_x = %-5d" (get_mouse_x());
    textprintf_ex screen font 16 64 black white "mouse_y = %-5d" (get_mouse_y());

    (* or you can use this function to measure the speed of movement.
     * Note that we only call it every fourth time round the loop:
     * there's no need for that other than to slow the numbers down
     * a bit so that you will have time to read them...
     *)
    incr c;
    if ((!c land 3) = 0) then
      begin
        let _mickeyx, _mickeyy = get_mouse_mickeys() in
        mickeyx := _mickeyx;
        mickeyy := _mickeyy;
      end;

    textprintf_ex screen font 16 88 black white "mickey_x = %-7d" !mickeyx;
    textprintf_ex screen font 16 104 black white "mickey_y = %-7d" !mickeyy;

    (* the mouse button state is stored in the variable mouse_b *)
    if left_button_pressed()
    then textout_ex screen font "left button is pressed " 16 128 black white
    else textout_ex screen font "left button not pressed" 16 128 black white;

    if right_button_pressed()
    then textout_ex screen font "right button is pressed " 16 144 black white
    else textout_ex screen font "right button not pressed" 16 144 black white;

    if middle_button_pressed()
    then textout_ex screen font "middle button is pressed " 16 160 black white
    else textout_ex screen font "middle button not pressed" 16 160 black white;

    (* the wheel position is stored in the variable mouse_z *)
    textprintf_ex screen font 16 184 black white "mouse_z = %-5d" (get_mouse_z());

    print_all_buttons();
    release_screen();
    vsync();
  done;

  clear_keybuf();

  (*  To display a mouse pointer, call show_mouse(). There are several
   *  things you should be aware of before you do this, though. For one,
   *  it won't work unless you call install_timer() first. For another,
   *  you must never draw anything onto the screen while the mouse
   *  pointer is visible. So before you draw anything, be sure to turn 
   *  the mouse off with show_mouse(NULL), and turn it back on again when
   *  you are done.
   *)
  clear_to_color screen (makecol 255 255 255);
  textout_centre_ex screen font "Press a key to change cursor"
                   (screen_w/2) (screen_h/2) (makecol 0 0 0)
                   (makecol 255 255 255);
  show_mouse screen;
  ignore(readkey());
  hide_mouse();

  (* create a custom mouse cursor bitmap... *)
  let custom_cursor = create_bitmap 32 32 in
  clear_to_color custom_cursor (bitmap_mask_color screen);
  for c=0 to pred 8 do
    circle custom_cursor 16 16 (c*2) (palette_color c);
  done;

  (* select the custom cursor and set the focus point to the middle of it *)
  set_mouse_sprite custom_cursor;
  set_mouse_sprite_focus 16 16;

  clear_to_color screen (makecol 255 255 255);
  textout_centre_ex screen font "Press a key to quit" (screen_w/2)
                   (screen_h/2) (makecol 0 0 0) (makecol 255 255 255);
  show_mouse screen;
  ignore(readkey());
  hide_mouse();

  destroy_bitmap custom_cursor;
;;

