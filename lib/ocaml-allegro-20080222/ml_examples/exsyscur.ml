(*  Example program for the Allegro library, by Peter Wang and
 *  Evert Glebbeek.
 *
 *  This program demonstrates the use of hardware accelerated mouse cursors.
 *)

open Allegro ;;



let do_select_cursor ~cursor =
  let screen, font = get_screen(), get_font() in
  let screen_w, screen_h = get_screen_width(), get_screen_height() in

  let black = makecol 0 0 0
  and white = makecol 255 255 255 in

  let gfx_capabilities = get_gfx_capabilities() in
  textout_centre_ex screen font
      ("Before: " ^
        (if (List.mem GFX_HW_CURSOR gfx_capabilities) then "HW_CURSOR   " else "no HW_CURSOR") ^", "^
        (if (List.mem GFX_SYSTEM_CURSOR gfx_capabilities) then "SYSTEM_CURSOR   " else "no SYSTEM_CURSOR"))
      (screen_w/2) (screen_h/3+2*(text_height font)) black white;

  select_mouse_cursor cursor;
  show_mouse screen;

  let gfx_capabilities = get_gfx_capabilities() in
  textout_centre_ex screen font
      ("After:  " ^
        (if (List.mem GFX_HW_CURSOR gfx_capabilities) then "HW_CURSOR   " else "no HW_CURSOR") ^", "^
        (if (List.mem GFX_SYSTEM_CURSOR gfx_capabilities) then "SYSTEM_CURSOR   " else "no SYSTEM_CURSOR"))
      (screen_w/2) (screen_h/3+3*(text_height font)) black white;
;;


let handle_key() =
  match readkey() with
  | '\027'
  | 'q'
  | 'Q' -> (true)
  | '1' ->
      do_select_cursor MOUSE_CURSOR_ALLEGRO;
      (false)
  | '2' ->
      do_select_cursor MOUSE_CURSOR_ARROW;
      (false)
  | '3' ->
      do_select_cursor MOUSE_CURSOR_BUSY;
      (false)
  | '4' ->
      do_select_cursor MOUSE_CURSOR_QUESTION;
      (false)
  | '5' ->
      do_select_cursor MOUSE_CURSOR_EDIT;
      (false)
  | _ ->
      (false)
;;


let () =
  (* Initialize Allegro *)
  begin try allegro_init()
  with _ ->
    allegro_message("Error initializing Allegro: "^ (get_allegro_error()) ^"\n");
    exit 1;
  end;
  
  (* Initialize mouse and keyboard *)
  install_timer();
  ignore(install_mouse());
  install_keyboard();
  
  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 640 480 0 0
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error setting video mode: "^ (get_allegro_error()) ^"\n");
      exit 1;
  end;
  let screen, font = get_screen(), get_font() in
  let screen_w, screen_h = get_screen_width(), get_screen_height() in

  let black = makecol 0 0 0
  and white = makecol 255 255 255 in

  clear_to_color screen white;
  
  textout_centre_ex screen font ("Graphics driver: "^ (get_gfx_driver_name()))
      (screen_w/2) (screen_h/3) black white;
  enable_hardware_cursor();

  let height = screen_h/3+5*(text_height font) in
  textout_centre_ex screen font "1) MOUSE_CURSOR_ALLEGRO  " (screen_w/2) height black white;
  let height = height + text_height font in
  textout_centre_ex screen font "2) MOUSE_CURSOR_ARROW    " (screen_w/2) height black white;
  let height = height + text_height font in
  textout_centre_ex screen font "3) MOUSE_CURSOR_BUSY     " (screen_w/2) height black white;
  let height = height + text_height font in
  textout_centre_ex screen font "4) MOUSE_CURSOR_QUESTION " (screen_w/2) height black white;
  let height = height + text_height font in
  textout_centre_ex screen font "5) MOUSE_CURSOR_EDIT     " (screen_w/2) height black white;
  let height = height + text_height font in
  textout_centre_ex screen font "Escape) Quit             " (screen_w/2) height black white;

  (* first cursor shown *)
  do_select_cursor MOUSE_CURSOR_ALLEGRO;

  while not(handle_key()) do () done;
;;

