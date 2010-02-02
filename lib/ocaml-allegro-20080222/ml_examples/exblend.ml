(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to use the translucency functions
 *  in truecolor video modes. Two image files are loaded from
 *  disk and displayed moving slowly around the screen. One of
 *  the images will be tinted to different colors. The other
 *  image will be faded out with a varying alpha strength, and
 *  drawn on top of the other image.
 *)

open Allegro ;;

let () =
  let bpp = ref (-1) in

  allegro_init();
  install_keyboard(); 
  install_timer();

  (* what color depth should we use? *)
  if (Array.length Sys.argv) > 1 then
  begin
    if ((Sys.argv.(1).[0] = '-') || (Sys.argv.(1).[0] = '/')) then
      let len = pred(String.length Sys.argv.(1)) in
      bpp := int_of_string(String.sub Sys.argv.(1) 1 len);

    if ((!bpp <> 15) && (!bpp <> 16) && (!bpp <> 24) && (!bpp <> 32)) then begin
      allegro_message("Invalid color depth '"^ Sys.argv.(1) ^"'\n");
      exit 1;
    end
  end;

  begin
    try
      if !bpp > 0 then begin
        (* set a user-requested color depth *)
        set_color_depth !bpp;
        set_gfx_mode GFX_AUTODETECT_WINDOWED 640 480 0 0;
      end
      else begin
        (* autodetect what color depths are available *)
        let rec iter_depths = function
          depth::tail ->
            bpp := depth;
            set_color_depth !bpp;
            begin try set_gfx_mode GFX_AUTODETECT_WINDOWED 640 480 0 0;
            with _ -> iter_depths tail end
        | [] ->
            failwith "no depth works"
        in
        iter_depths [16; 15; 32; 24];
      end;
    with _ ->
      (* did the video mode set properly? *)
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error setting "^ (string_of_int !bpp) ^
                      " bit graphics mode\n"^
        	      (get_allegro_error()) ^"\n");
      exit 1;
  end;

  (* specify that images should be loaded in a truecolor pixel format *)
  set_color_conversion COLORCONV_TOTAL;

  let pal = new_palette() in

  (* load the first picture *)
  let buf = replace_filename Sys.argv.(0) "allegro.pcx" in
  let image1 =
    try load_bitmap buf pal
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error reading "^ buf ^"!\n");
      exit 1;
  in

  (* load the second picture *)
  let buf = replace_filename Sys.argv.(0) "mysha.pcx" in
  let image2 =
    try load_bitmap buf pal
    with _ ->
      destroy_bitmap image1;
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error reading "^ buf ^"!\n");
      exit 1;
  in

  (* create a double buffer bitmap *)
  let screen_w, screen_h = get_screen_width(), get_screen_height() in
  let buffer = create_bitmap screen_w screen_h in

  let screen, font = get_screen(), get_font() in

  (* Note that because we loaded the images as truecolor bitmaps, we don't
   * need to bother setting the palette, and we can display both on screen
   * at the same time even though the source files use two different 256
   * color palettes...
   *)

  textout_ex screen font ((string_of_int !bpp) ^" bpp") 0 (screen_h-8) (makecol 255 255 255) (color_index 0);

  let rec main_loop prevx1 prevy1 prevx2 prevy2 =
    let timer = retrace_count() in
    clear_bitmap buffer;

    (* the first image moves in a slow circle while being tinted to 
     * different colors...
     *)
    (*
      x1= 160+fixtoi(fixsin(itofix(timer)/16)*160);
      y1= 140-fixtoi(fixcos(itofix(timer)/16)*140);
      r = 127-fixtoi(fixcos(itofix(timer)/6)*127);
      g = 127-fixtoi(fixcos(itofix(timer)/7)*127);
      b = 127-fixtoi(fixcos(itofix(timer)/8)*127);
      a = 127-fixtoi(fixcos(itofix(timer)/9)*127);
    *)
    let x1= 160 + fixtoi(fixmul(fixsin(fixdiv(itofix timer) (itofix 16))) (itofix 160)) in
    let y1= 140 - fixtoi(fixmul(fixcos(fixdiv(itofix timer) (itofix 16))) (itofix 140)) in
    let r = 127 - fixtoi(fixmul(fixcos(fixdiv(itofix timer) (itofix  6))) (itofix 127)) in
    let g = 127 - fixtoi(fixmul(fixcos(fixdiv(itofix timer) (itofix  7))) (itofix 127)) in
    let b = 127 - fixtoi(fixmul(fixcos(fixdiv(itofix timer) (itofix  8))) (itofix 127)) in
    let a = 127 - fixtoi(fixmul(fixcos(fixdiv(itofix timer) (itofix  9))) (itofix 127)) in
    set_trans_blender r g b 0;
    draw_lit_sprite buffer image1 x1 y1 (color_index a);
    textout_ex screen font (Printf.sprintf "light: %d " a) 0 0 (makecol r g b) (color_index 0);

    (* the second image moves in a faster circle while the alpha value
     * fades in and out...
     *)
    (*
      x2= 160 + fixtoi(fixsin(fixdiv (itofix timer) (itofix 10)) * 160);
      y2= 140 - fixtoi(fixcos(fixdiv (itofix timer) (itofix 10)) * 140);
      a = 127 - fixtoi(fixcos(fixdiv (itofix timer) (itofix  4)) * 127);
    *)
    let x2= 160 + fixtoi(fixmul(fixsin(fixdiv(itofix timer) (itofix 10))) (itofix 160)) in
    let y2= 140 - fixtoi(fixmul(fixcos(fixdiv(itofix timer) (itofix 10))) (itofix 140)) in
    let a = 127 - fixtoi(fixmul(fixcos(fixdiv(itofix timer) (itofix  4))) (itofix 127)) in
    set_trans_blender 0 0 0 a;
    draw_trans_sprite buffer image2 x2 y2;
    textout_ex screen font (Printf.sprintf "alpha: %d " a) 0 8 (makecol a a a) (color_index 0);

    (* copy the double buffer across to the screen *)
    vsync();

    let x = min x1 prevx1 in
    let y = min y1 prevy1 in
    let w = (max x1 prevx1) + 320 - x in
    let h = (max y1 prevy1) + 200 - y in
    blit buffer screen x y x y w h;

    let x = min x2 prevx2 in
    let y = min y2 prevy2 in
    let w = (max x2 prevx2) + 320 - x in
    let h = (max y2 prevy2) + 200 - y in
    blit buffer screen x y x y w h;

    if not(keypressed()) then main_loop x1 y1 x2 y2
  in
  main_loop 0 0 0 0;

  clear_keybuf();

  destroy_bitmap image1;
  destroy_bitmap image2;
  destroy_bitmap buffer;
;;

