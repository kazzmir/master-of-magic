(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates the use of triple buffering. Several
 *  triangles are displayed rotating and bouncing on the screen
 *  until you press a key. Note that on some platforms you
 *  can't get real hardware triple buffering.  The Allegro code
 *  remains the same, but most likely the graphic driver will
 *  emulate it. Unfortunately, in these cases you can't expect
 *  the animation to be completely smooth and flicker free.
 *)

open Allegro ;;

let num_shapes = 16 ;;

type shape =
  {
    color:color;           (* color of the shape *)
    mutable x:fixed;
    mutable y:fixed;       (* centre of the shape *)
    mutable dir1:fixed;
    mutable dir2:fixed;
    mutable dir3:fixed;    (* directions to the three corners *)
    dist1:fixed; dist2:fixed; dist3:fixed;  (* distances to the three corners *)
    mutable xc:fixed;
    mutable yc:fixed;
    mutable ac:fixed;     (* position and angle change values *)
  }


let shapes =
  let null = itofix 0 in
  Array.make num_shapes {
    color=color_index 0;
    x=null; y=null;
    dir1=null; dir2=null; dir3=null;
    dist1=null; dist2=null; dist3=null;
    xc=null; yc=null; ac=null;
  }
;;

let triplebuffer_not_available = ref false ;;

let screen_w = ref 0 ;;
let screen_h = ref 0 ;;


(* randomly initialises a shape structure *)
let init_shape shape i =

  shape.(i) <-
    {
      color = color_index (1+(Random.int 15));

      (* randomly position the corners *)
      dir1 = itofix(Random.int 256);
      dir2 = itofix(Random.int 256);
      dir3 = itofix(Random.int 256);

      dist1 = itofix(Random.int 64);
      dist2 = itofix(Random.int 64);
      dist3 = itofix(Random.int 64);

      (* rand centre position and movement speed/direction *)
      x = itofix(Random.int !screen_w);
      y = itofix(Random.int !screen_h);
      ac = itofix((Random.int 9)-4);
      xc = itofix((Random.int 7)-2);
      yc = itofix((Random.int 7)-2);
    }
;;



(* updates the position of a shape structure *)
let move_shape shape =

  shape.x <- fixadd shape.x shape.xc;
  shape.y <- fixadd shape.y shape.yc;

  shape.dir1 <- fixadd shape.dir1 shape.ac;
  shape.dir2 <- fixadd shape.dir2 shape.ac;
  shape.dir3 <- fixadd shape.dir3 shape.ac;

  if (((fixtoi shape.x <= 0) && (fixtoi shape.xc < 0)) ||
      ((fixtoi shape.x >= !screen_w) && (fixtoi shape.xc > 0))) then
  begin
    shape.xc <- fixsub (itofix 0) shape.xc;
    shape.ac <- itofix((Random.int 9) - 4);
  end;

  if (((fixtoi shape.y <= 0) && (fixtoi shape.yc < 0)) ||
      ((fixtoi shape.y >= !screen_h) && (fixtoi shape.yc > 0))) then
  begin
    shape.yc <- fixsub (itofix 0) shape.yc;
    shape.ac <- itofix((Random.int 9) - 4);
  end
;;



(* draws a frame of the animation *)
let draw b =
  acquire_bitmap b;
  clear_bitmap b;

  for c=0 to pred num_shapes do
    let x = shapes.(c).x
    and y = shapes.(c).y
    in
    let dist1 = shapes.(c).dist1
    and dist2 = shapes.(c).dist2
    and dist3 = shapes.(c).dist3
    in
    let dir1 = shapes.(c).dir1
    and dir2 = shapes.(c).dir2
    and dir3 = shapes.(c).dir3
    in
    triangle b
             (fixtoi(fixadd x (fixmul dist1 (fixcos dir1))))
             (fixtoi(fixadd y (fixmul dist1 (fixsin dir1))))
             (fixtoi(fixadd x (fixmul dist2 (fixcos dir2))))
             (fixtoi(fixadd y (fixmul dist2 (fixsin dir2))))
             (fixtoi(fixadd x (fixmul dist3 (fixcos dir3))))
             (fixtoi(fixadd y (fixmul dist3 (fixsin dir3))))
             shapes.(c).color;

    move_shape shapes.(c);
  done;

  let message =
    if !triplebuffer_not_available
    then "Simulated triple buffering"
    else "Real triple buffering"
  in

  textout_ex b (get_font()) message 0 0 (color_index 255) (color_index(-1));

  release_bitmap b;
;;


type next = A | B | C

(* main animation control loop *)
let triple_buffer ~page1 ~page2 ~page3 =

  let rec loop page active_page =

    (* draw a frame *)
    draw active_page;

    (* make sure the last flip request has actually happened *)
    while poll_scroll() do () done;

    (* post a request to display the page we just drew *)
    begin
      try request_video_bitmap active_page;
      with _ -> ()
    end;
    (*
    blit active_page screen 0 0 0 0 !screen_w !screen_h;
    *)

    (* update counters to point to the next page *)
    let (page, active_page) =
      match page with
      | A -> (B, page2)
      | B -> (C, page3)
      | C -> (A, page1)
    in

    if not(keypressed()) then loop page active_page
  in
  loop A page1;

  clear_keybuf();
;;



let () =
  Random.self_init();
  allegro_init();
  install_timer();
  install_keyboard();
  ignore(install_mouse());

  let (w, h) =
    if allegro_dos()
    then (320, 240)
    else (640, 480)
  in

  (* see comments in exflip.c *)
  begin
    try
      if allegro_vram_single_surface()
      then set_gfx_mode GFX_AUTODETECT_WINDOWED w h 0 (h * 3)
      else set_gfx_mode GFX_AUTODETECT_WINDOWED w h 0 0
    with _ ->
      try set_gfx_mode GFX_SAFE w h 0 0
      with _ ->
        set_gfx_mode GFX_TEXT 0 0 0 0;
        allegro_message("Unable to set any graphic mode\n"^
                        (get_allegro_error()) ^"\n");
        exit 1;
  end;

  screen_w := get_screen_width();
  screen_h := get_screen_height();

  set_palette(get_desktop_palette());

  (* if triple buffering isn't available, try to enable it *)
  if not(List.mem GFX_CAN_TRIPLE_BUFFER (get_gfx_capabilities())) then
    ignore(enable_triple_buffer());

  (* if that didn't work, give up *)
  if not(List.mem GFX_CAN_TRIPLE_BUFFER (get_gfx_capabilities())) then
    triplebuffer_not_available := true;

  (* allocate three sub bitmaps to access pages of the screen *)
  let page1, page2, page3 =
    try
    ( create_video_bitmap !screen_w !screen_h,
      create_video_bitmap !screen_w !screen_h,
      create_video_bitmap !screen_w !screen_h )
    with _ ->
      set_gfx_mode GFX_TEXT 0 0 0 0;
      allegro_message "Unable to create three video memory pages\n";
      exit 1;
  in

  (* initialise the shapes *)
  for c=0 to pred num_shapes do
    init_shape shapes c;
  done;

  triple_buffer page1 page2 page3;

  destroy_bitmap page1;
  destroy_bitmap page2;
  destroy_bitmap page3;
;;

