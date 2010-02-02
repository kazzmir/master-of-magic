(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to play samples. You have to
 *  use this example from the command line to specify as first
 *  parameter a WAV or VOC sound file to play. If the file is
 *  loaded successfully, the sound will be played in an infinite
 *  loop. While it is being played, you can use the left and right
 *  arrow keys to modify the panning of the sound. You can also
 *  use the up and down arrow keys to modify the pitch.
 *)

open Allegro ;;

let () =
  let pan = ref 128 in
  let pitch = ref 1000 in

  allegro_init();

  if (Array.length Sys.argv) <> 2 then begin
     allegro_message "Usage: 'exsample filename.[wav|voc]'\n";
     exit 1;
  end;

  install_keyboard(); 
  install_timer();

  (* install a digital sound driver *)
  begin
    try install_sound DIGI_AUTODETECT  MIDI_NONE
    with _ ->
      allegro_message("Error initialising sound system\n"^ (get_allegro_error()) ^"\n");
      exit 1;
  end;

  (* read in the WAV file *)
  let the_sample =
    try load_sample Sys.argv.(1)
    with _ -> 
      allegro_message("Error reading WAV file "^ Sys.argv.(1) ^"\n");
      exit 1;
  in

  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED  320 200  0 0
    with _ ->
      try set_gfx_mode GFX_SAFE  320 200  0 0
      with _ ->
        set_gfx_mode GFX_TEXT 0 0 0 0;
        allegro_message("Unable to set any graphic mode\n"^ (get_allegro_error()) ^"\n");
        exit 1;
  end;

  set_palette(get_desktop_palette());
  let screen = get_screen() in
  clear_to_color screen (makecol 255 255 255);

  let (screen_w, screen_h) = get_screen_width(), get_screen_height() in

  let font = get_font() in
  let black, transp = (makecol 0 0 0), color_index (-1) in
  textout_centre_ex screen font ("Driver: "^ digi_driver_name())  (screen_w/2) (screen_h/3) black transp;
  textout_centre_ex screen font ("Playing "^ Sys.argv.(1))        (screen_w/2) (screen_h/2) black transp;
  textout_centre_ex screen font "Use the arrow keys to adjust it" (screen_w/2) (screen_h*2/3) black transp;

  (* start up the sample *)
  ignore(play_sample the_sample 255 !pan !pitch true);

  while not(key_esc()) do
     poll_keyboard();

     (* alter the pan position? *)
     if (key_left() && (!pan > 0)) then
       decr pan
     else if (key_right() && (!pan < 255)) then
       incr pan;

     (* alter the pitch? *)
     if (key_up() && (!pitch < 16384)) then
        pitch := ((!pitch * 513) / 512) + 1
     else if (key_down() && (!pitch > 64)) then
        pitch := ((!pitch * 511) / 512) - 1; 

     (* adjust the sample *)
     adjust_sample the_sample 255 !pan !pitch true;

     (* delay a bit *)
     rest 2;
  done;

  (* destroy the sample *)
  destroy_sample the_sample;
;;

