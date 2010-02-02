(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to use an offscreen part of
 *  the video memory to store source graphics for a hardware
 *  accelerated graphics driver. The example loads the `mysha.pcx'
 *  file and then blits it several times on the screen. Depending
 *  on whether you have enough video memory and Allegro supports
 *  the hardware acceleration features of your card, your success
 *  running this example may be none at all, sluggish performance
 *  due to software emulation, or flicker free smooth hardware
 *  accelerated animation.
 *)

open Allegro ;;

let max_images = 256 ;;


(* structure to hold the current position and velocity of an image *)
type image =
  {
    mutable x:float;
    mutable y:float;
    mutable dx:float;
    mutable dy:float;
  }


(* initialises an image structure to a random position and velocity *)
let init_image ~images ~i =
  images.(i) <- {
      x = (float)(Random.int 704);
      y = (float)(Random.int 568);
      dx = (float)((Random.int 255) - 127) /. 32.0;
      dy = (float)((Random.int 255) - 127) /. 32.0;
    }
;;



(* called once per frame to bounce an image around the screen *)
let update_image ~image =

  image.x <- image.x +. image.dx;
  image.y <- image.y +. image.dy;

  if (((image.x < 0.) && (image.dx < 0.)) ||
      ((image.x > 703.) && (image.dx > 0.))) then
     image.dx <- image.dx *. -1.;

  if (((image.y < 0.) && (image.dy < 0.)) ||
      ((image.y > 567.) && (image.dy > 0.))) then
     image.dy <- image.dy *. -1.;
;;



let () =
  allegro_init();
  install_keyboard(); 
  install_timer();
  Random.self_init();

  (* see comments in exflip.c *)
  begin
    try
      if allegro_vram_single_surface()
      then set_gfx_mode GFX_AUTODETECT_WINDOWED 1024 768 0 (2 * 768 + 200)
      else set_gfx_mode GFX_AUTODETECT_WINDOWED 1024 768 0 0
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error setting graphics mode\n"^ (get_allegro_error()) ^"\n");
      exit 1;
  end;

  (* read in the source graphic *)
  let buf = replace_filename Sys.argv.(0) "mysha.pcx" in
  let pal = new_palette() in
  let image =
    try load_bitmap buf pal;
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Error reading "^ buf ^"!\n");
      exit 1;
  in

  set_palette pal;

  let images = Array.make max_images {x=0.; y=0.; dx=0.; dy=0.} in

  (* initialise the images to random positions *)
  for i=0 to pred max_images do
     init_image images i;
  done;

  let screen_w, screen_h = get_screen_width(), get_screen_height() in
  let image_w, image_h = get_bitmap_width image, get_bitmap_height image in

  let page, vimage =
    try
      (* create two video memory bitmaps for page flipping *)
      let page =
        Array.init 2 (fun i -> create_video_bitmap screen_w screen_h)
      in

      (* create a video memory bitmap to store our picture *)
      let vimage = create_video_bitmap image_w image_h in

      (page, vimage)
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message("Not enough video memory (need two 1024x768 pages "^
        	      "and a 320x200 image)\n");
      exit 1;
  in

  (* copy the picture into offscreen video memory *)
  blit image vimage 0 0 0 0 image_w image_h;

  let font = get_font() in
  let num_images = ref 4 in
  let page_num = ref 1 in
  let _done = ref false in

  while not(!_done) do
    acquire_bitmap page.(!page_num);

    (* clear the screen *)
    clear_bitmap page.(!page_num);

    (* draw onto it *)
    for i=0 to pred !num_images do
      blit vimage page.(!page_num) 0 0 (int_of_float images.(i).x) (int_of_float images.(i).y) image_w image_h;
    done;

    let transparent = transparent() in

    textout_ex page.(!page_num) font
               ("Images: "^ (string_of_int !num_images) ^" (arrow keys to change)")
               0 0 (color_index 255) transparent;

    let gfx_capabilities = get_gfx_capabilities() in

    (* tell the user which functions are being done in hardware *)
    if (List.mem GFX_HW_FILL gfx_capabilities) then
       textout_ex page.(!page_num) font "Clear: hardware accelerated" 0 16 (color_index 255) transparent
    else
       textout_ex page.(!page_num) font "Clear: software (urgh, this is not good!)" 0 16 (color_index 255) transparent;

    if (List.mem GFX_HW_VRAM_BLIT gfx_capabilities) then
       textout_ex page.(!page_num) font "Blit: hardware accelerated" 0 32 (color_index 255) transparent
    else
       textout_ex page.(!page_num) font ("Blit: software (urgh, this program "^
      	    "will run too sloooooowly without hardware acceleration!)")
      	    0 32 (color_index 255) transparent;

    release_bitmap page.(!page_num);

    (* page flip *)
    show_video_bitmap page.(!page_num);
    page_num := 1 - !page_num;

    (* deal with keyboard input *)
    while keypressed() do
       match readkey_scancode() with
       | KEY_UP
       | KEY_RIGHT ->
           if !num_images < max_images then incr num_images;

       | KEY_DOWN
       | KEY_LEFT ->
           if !num_images > 0 then decr num_images;

       | KEY_ESC ->
             _done := true;

       | _ -> ()
    done;

    (* bounce the images around the screen *)
    for i=0 to pred !num_images do
      update_image images.(i);
    done;
  done;

  destroy_bitmap image;
  destroy_bitmap vimage;
  destroy_bitmap page.(0);
  destroy_bitmap page.(1);
;;

