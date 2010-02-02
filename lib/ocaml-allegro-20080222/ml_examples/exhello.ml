(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This is a very simple program showing how to get into graphics
 *  mode and draw text onto the screen.
 *)

open Allegro ;;

let () =
   (* you should always do this at the start of Allegro programs *)
   allegro_init();

   (* set up the keyboard handler *)
   install_keyboard(); 

   (* set a graphics mode sized 320x200 *)
   begin
     try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 200 0 0;
     with _ ->
       try set_gfx_mode GFX_SAFE 320 200 0 0;
       with _ ->
	 set_gfx_mode GFX_TEXT 0 0 0 0;
	 allegro_message("Unable to set any graphic mode\n"^ (get_allegro_error()) ^"\n");
	 exit 1;
   end;

   (* set the color palette *)
   set_palette(get_desktop_palette());

   (* clear the screen to white *)
   let screen = get_screen() in
   clear_to_color screen (makecol 255 255 255);

   (* you don't need to do this, but on some platforms (eg. Windows) things
    * will be drawn more quickly if you always acquire the screen before
    * trying to draw onto it.
    *)
   acquire_screen();

   let (screen_w, screen_h) = get_screen_width(), get_screen_height() in
   (* write some text to the screen with black letters and transparent background *)
   textout_centre_ex screen (get_font()) "Hello, world!" (screen_w/2) (screen_h/2) (makecol 0 0 0) (color_index(-1));

   (* you must always release bitmaps before calling any input functions *)
   release_screen();

   (* wait for a key press *)
   ignore(readkey());
;;

