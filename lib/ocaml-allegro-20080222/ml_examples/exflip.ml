(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program moves a circle across the screen, first with a
 *  double buffer and then using page flips.
 *)

open Allegro ;;

let () =
  allegro_init();
  install_timer();
  install_keyboard();

  (* Some platforms do page flipping by making one large screen that you
   * can then scroll, while others give you several smaller, unique
   * surfaces. If you use the create_video_bitmap() function, the same
   * code can work on either kind of platform, but you have to be careful
   * how you set the video mode in the first place. We want two pages of
   * 320x200 video memory, but if we just ask for that, on DOS Allegro
   * might use a VGA driver that won't later be able to give us a second
   * page of vram. But if we ask for the full 320x400 virtual screen that
   * we want, the call will fail when using DirectX drivers that can't do
   * this. So we try two different mode sets, first asking for the 320x400
   * size, and if that doesn't work, for 320x200.
   *)
  begin
    try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 200 0 400;
    with _ -> try set_gfx_mode GFX_AUTODETECT_WINDOWED 320 200 0 0;
      with _ -> try set_gfx_mode GFX_SAFE 320 200 0 0;
        with _ ->
          set_gfx_mode GFX_TEXT 0 0 0 0;
          allegro_message("Unable to set any graphic mode\n"^ (get_allegro_error()) ^"\n");
          exit 1;
  end;

  set_palette(get_desktop_palette());

  (* allocate the memory buffer *)
  let (screen_w, screen_h) = (get_screen_width(), get_screen_height()) in
  let buffer = create_bitmap screen_w screen_h in

  let font = get_font() in
  let screen = get_screen() in
  let break() = () in

  (* first with a double buffer... *)
  clear_keybuf();
  let c = (retrace_count())+32 in
  while (retrace_count())-c <= screen_w+32 do
    clear_to_color buffer (makecol 255 255 255);
    circlefill buffer ((retrace_count())-c) (screen_h/2) 32 (makecol 0 0 0);
    textout_ex buffer font "Double buffered" 0 0 (makecol 0 0 0) (transparent());
    blit buffer screen 0 0 0 0 screen_w screen_h;

    if keypressed() then break();
  done;

  destroy_bitmap buffer;

  (* now create two video memory bitmaps for the page flipping *)
  let (page1, page2) =
    try (create_video_bitmap screen_w screen_h,
         create_video_bitmap screen_w screen_h)
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message "Unable to create two video memory pages\n";
      exit 1;
  in

  let active_page = ref page2 in

  (* do the animation using page flips... *)
  clear_keybuf();
  for c = (-32) to screen_w+32 do
    clear_to_color !active_page (makecol 255 255 255);
    circlefill !active_page c (screen_h/2) 32 (makecol 0 0 0);
    textout_ex !active_page font "Page flipping" 0 0 (makecol 0 0 0) (transparent());
    show_video_bitmap !active_page;

    if !active_page == page1
    then active_page := page2
    else active_page := page1;

    if keypressed() then break();
  done;

  destroy_bitmap page1;
  destroy_bitmap page2;
;;

(* vim: sw=2 sts=2 ts=2 et fdm=marker
 *)
