(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to load and display bitmap files
 *  in truecolor video modes, and how to crossfade between them.
 *  You have to use this example from the command line to specify
 *  as parameters a number of graphic files. Use at least two
 *  files to see the graphical effect. The example will crossfade
 *  from one image to another with each key press until you press
 *  the ESC key.
 *)

open Allegro ;;

type return_type = Show_Error | Show_OK | Quit_request ;;
exception Show_return of return_type ;;

let show ~name =
  let return r = raise(Show_return r) in
  try
    let pal = new_palette() in

    (* load the file *)
    let bmp =
      try load_bitmap name pal
      with _ -> return Show_Error
    in

    let screen = get_screen() in
    let screen_w, screen_h = get_screen_width(), get_screen_height() in

    let buffer = create_bitmap screen_w screen_h in
    blit screen buffer 0 0 0 0 screen_w screen_h;

    set_palette pal;

    let bmp_w, bmp_h = get_bitmap_width ~bmp, get_bitmap_height ~bmp in

    (* fade it in on top of the previous picture *)
    let speed = 4 in
    for a=0 to 256/speed do
      let alpha = a * speed in
      set_trans_blender 0 0 0 alpha;
      draw_trans_sprite buffer bmp ((screen_w - bmp_w)/2) ((screen_h - bmp_h)/2);
      vsync();
      blit buffer screen 0 0 0 0 screen_w screen_h;
      if keypressed() then
        begin
          destroy_bitmap bmp;
          destroy_bitmap buffer;
          if key_esc()
          then return Quit_request
          else return Show_OK
        end
    done;

    blit bmp screen 0 0 ((screen_w - bmp_w)/2) ((screen_h - bmp_h)/2) bmp_w bmp_h;

    destroy_bitmap bmp;
    destroy_bitmap buffer;

    if key_esc()
    then return Quit_request
    else return Show_OK
  with
  | Show_return r -> r
  | x -> raise x
;;



let () =
   allegro_init();

   let argc = Array.length Sys.argv in
   if argc < 2 then begin
      allegro_message "Usage: 'exxfade files.[bmp|lbm|pcx|tga]'\n";
      exit 1;
   end;

   install_keyboard(); 

   (* set the best color depth that we can find *)
   let gfx_driver = GFX_AUTODETECT_WINDOWED in
   begin
     set_color_depth 16;
     try set_gfx_mode gfx_driver 640 480 0 0
     with _ ->
       set_color_depth 15;
       try set_gfx_mode gfx_driver 640 480 0 0
       with _ ->
         set_color_depth 32;
         try set_gfx_mode gfx_driver 640 480 0 0
         with _ ->
           set_color_depth 24;
           try set_gfx_mode gfx_driver 640 480 0 0
           with _ ->
             set_gfx_mode GFX_TEXT 0 0 0 0;
             allegro_message("Error setting graphics mode\n"^ (get_allegro_error()) ^"\n");
             exit 1;
   end;

   (* load all images in the same color depth as the display *)
   set_color_conversion COLORCONV_TOTAL;

   (* process all the files on our command line *)
   let rec loop i =
      match (show Sys.argv.(i)) with
      | Show_Error ->
          set_gfx_mode GFX_TEXT 0 0 0 0;
          allegro_message("Error loading image file '"^ Sys.argv.(i) ^"'\n");
          exit 1;

      | Show_OK ->
          (* next! *)
          if (succ i >= argc) then loop 1 else loop(succ i)

      | Quit_request ->
          allegro_exit();
   in
   loop 1;
;;

