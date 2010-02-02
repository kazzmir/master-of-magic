(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates the use of patterned drawing and sub-bitmaps.
 *)

open Allegro ;;


let draw_pattern ~bitmap ~message ~color =

  acquire_bitmap bitmap;

  let screen = get_screen() in
  let font = get_font() in

  (* create a pattern bitmap *)
  let pattern = create_bitmap (text_length font message) (text_height font) in
  clear_to_color pattern (bitmap_mask_color pattern);
  textout_ex pattern font message 0 0 (palette_color 255) (bitmap_mask_color screen);

  let bitmap_w, bitmap_h = get_bitmap_width bitmap, get_bitmap_height bitmap in

  (* cover the bitmap with the pattern *)
  drawing_mode(DRAW_MODE_MASKED_PATTERN(pattern, 0, 0));
  rectfill bitmap 0 0 bitmap_w bitmap_h (palette_color color);
  solid_mode();

  (* destroy the pattern bitmap *)
  destroy_bitmap pattern;

  release_bitmap bitmap;
;;


let () =
  allegro_init();
  install_keyboard();

  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 200 0 0
    with _ ->
      try set_gfx_mode GFX_SAFE 320 200 0 0
      with _ ->
        set_gfx_mode GFX_TEXT 0 0 0 0;
        allegro_message("Unable to set any graphic mode\n"^(get_allegro_error())^"\n");
        exit 1;
  end;

  let screen = get_screen() in

  set_palette(get_desktop_palette());
  clear_to_color screen (makecol 255 255 255);

  (* first cover the whole screen with a pattern *)
  draw_pattern screen "<screen>" 255;

  (* draw the pattern onto a memory bitmap and then blit it to the screen *)
  let bitmap = create_bitmap 128 32 in
  clear_to_color bitmap (makecol 255 255 255);
  draw_pattern bitmap "<memory>" 1;
  masked_blit bitmap screen 0 0 32 32 128 32;
  destroy_bitmap bitmap;

  (* or we could use a sub-bitmap. These share video memory with their
   * parent, so the drawing will be visible without us having to blit it
   * across onto the screen.
   *)
  let bitmap = create_sub_bitmap screen 224 64 64 128 in
  rectfill screen 224 64 286 192 (makecol 255 255 255);
  draw_pattern bitmap "<subbmp>" 4;
  destroy_bitmap bitmap;

  ignore(readkey());
;;

