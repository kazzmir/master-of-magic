(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates the use of memory bitmaps. It creates
 *  a small temporary bitmap in memory, draws some circles onto it,
 *  and then blits lots of copies of it onto the screen.
 *)

open Allegro ;;

let () =
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

  set_palette(get_desktop_palette());

  (* make a memory bitmap sized 20x20 *)
  let memory_bitmap = create_bitmap 20 20 in

  (* draw some circles onto it *)
  clear_bitmap memory_bitmap;
  for x=0 to pred 16 do
    circle memory_bitmap 10 10 x (palette_color x);
  done;

  (* blit lots of copies of it onto the screen *)
  acquire_screen();

  let (screen_w, screen_h) = get_screen_width(), get_screen_height() in
  let screen = get_screen() in

  for y=0 to pred(screen_h / 20) do
    for x=0 to pred(screen_w / 20) do
      blit memory_bitmap screen 0 0 (x*20) (y*20) 20 20;
    done;
  done;

  release_screen();

  (* free the memory bitmap *)
  destroy_bitmap memory_bitmap;

  ignore(readkey());
;;

