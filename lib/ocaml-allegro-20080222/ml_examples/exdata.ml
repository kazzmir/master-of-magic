(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to access the contents of an
 *  Allegro datafile (created by the grabber utility). The example
 *  loads the file `example.dat', then blits a bitmap and shows
 *  a font, both from this datafile.
 *)

open Allegro ;;


(* the grabber allegro program produces a C header "example.h",
 * which contains defines for the names of all the objects in
 * the datafile (BIG_FONT, SILLY_BITMAP, etc), that are here used
 * as indexes.
 *)
let big_font = 0 ;;
let silly_bitmap = 1 ;;
let the_palette = 2 ;;


let () =
  allegro_init();
  install_keyboard();

  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 200 0 0
    with _ ->
      try set_gfx_mode GFX_SAFE 320 200 0 0
      with _ ->
        set_gfx_mode GFX_TEXT 0 0 0 0;
        allegro_message("Unable to set any graphic mode\n"^ (get_allegro_error()) ^"\n");
        exit 1;
  end;

  (* we still don't have a palette => Don't let Allegro twist colors *)
  set_color_conversion COLORCONV_NONE;

  (* load the datafile into memory *)
  let buf = replace_filename Sys.argv.(0) "example.dat" in
  let datafile =
    try load_datafile buf
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error loading "^ buf ^"!\n");
      exit 1;
  in

  (* select the palette which was loaded from the datafile *)
  set_palette ((item_dat ~dat:datafile ~idx:the_palette):palette);

  (* aha, set a palette and let Allegro convert colors when blitting *)
  set_color_conversion COLORCONV_TOTAL;

  let screen, font = get_screen(), get_font() in

  (* display the bitmap from the datafile *)
  textout_ex screen font "This is the bitmap:" 32 16 (makecol 255 255 255) (color_index(-1));
  blit (bitmap_dat ~dat:datafile ~idx:silly_bitmap) screen  0 0 64 32 64 64;

  (* and use the font from the datafile *)
  textout_ex screen (font_dat ~dat:datafile ~idx:big_font) "And this is a big font!"
             32 128 (makecol 0 255 0) (color_index(-1));

  ignore(readkey());

  (* unload the datafile when we are finished with it *)
  unload_datafile datafile;
;;

