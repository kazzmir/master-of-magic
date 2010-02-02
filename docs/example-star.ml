(*
let speed_counter = ref 0;;
let dont_care ~param : unit =
  Printf.printf "hello %d\n" !speed_counter;
  speed_counter := !speed_counter + 1;
  ignore()
;;
*)

let gfx_x = 640;;
let gfx_y = 480;;
let max_stars = 1500;;

let init () =
  Allegro.allegro_init ();
  Allegro.set_color_depth 16;
  Allegro.set_gfx_mode Allegro.GFX_AUTODETECT_WINDOWED gfx_x gfx_y 0 0;
  Allegro.install_timer ();
  Allegro.install_keyboard ();
  (* bar 0 (fun ~param -> ()); *)
  (*
  Allegro.install_param_int 60 dont_care 0;
  *)
  (*
  let foo = Allegro.install_param_int 60
  in ();
  *)
;;

(* generic allegro timer loop *)
let run world (logic : 'a -> unit) (draw : 'a -> Allegro.bitmap -> unit)  =
  let current_time () = Unix.gettimeofday () in
  let fps n = 1.0 /. n in
  (* todo: pass fps number as a parameter *)
  let min_count = fps 35.0 in
  let (screen_w, screen_h) = (Allegro.get_screen_width(), Allegro.get_screen_height()) in
  let buffer = Allegro.create_bitmap screen_w screen_h in
  let main_loop () =
    let rec loop start now dodraw =
      if now -. start > min_count then begin
        (logic world);
        (* if you are worried that `logic' might take longer to execute
         * than `min_count' then pass `now` instead of (current_time ()).
         * this will ensure termination of the logic loop, but the game
         * will look choppy.
         *)
        if not (Allegro.key_esc()) then
          loop (start +. min_count) (current_time ()) true;
      end else begin
        if dodraw then begin
          (Allegro.clear_bitmap buffer);
          (draw world buffer);
          (Allegro.blit buffer (Allegro.get_screen ()) 0 0 0 0 screen_w
          screen_h)
        end;
        Allegro.rest 10;
        loop start (current_time ()) false
      end
    in
    loop (current_time ()) (current_time ()) false
  in
  ignore (main_loop ())
;;

(* structure that defines a star *)
type star = {
  front_z : float ref;
  back_z : float ref;
  speed : float;
  center_x : float;
  center_y : float;
  x : float ref;
  y : float ref;
}

let blend_palette colors r1 g1 b1 r2 g2 b2 =
  let make_color n =
    let j = (float n) /. (float colors) in
    let r = 0.5 +. (float r1) +. ((float r2) -. (float r1)) *. j in
    let g = 0.5 +. (float g1) +. ((float g2) -. (float g1)) *. j in
    let b = 0.5 +. (float b1) +. ((float b2) -. (float b1)) *. j in
    Allegro.makecol (truncate r) (truncate g) (truncate b)
  in
  let rec all num =
    match num with
    | 0 -> []
    | n -> (make_color n) :: (all (num - 1))
  in
  all colors
;;

let absF (f:float) = if f > 0.0 then f else (f *. -1.0);;

class color_changer = object(self)
  val mutable r1 = 0.0
  val r1_max = 64.0
  val mutable r2 = 255.0
  val r2_min = 128.0
  val mutable g1 = 0.0
  val g1_max = 64.0
  val mutable g2 = 255.0
  val g2_min = 128.0
  val mutable b1 = 0.0
  val b1_max = 64.0
  val mutable b2 = 255.0
  val b2_min = 128.0

  val r1_velocity = ref 0.0
  val r1_want_velocity = ref 0.0
  val r2_velocity = ref 0.0
  val r2_want_velocity = ref 0.0
  val g1_velocity = ref 0.0
  val g1_want_velocity = ref 0.0
  val g2_velocity = ref 0.0
  val g2_want_velocity = ref 0.0
  val b1_velocity = ref 0.0
  val b1_want_velocity = ref 0.0
  val b2_velocity = ref 0.0
  val b2_want_velocity = ref 0.0

  method check color min max =
    if color < min then
      min
    else if color > max then
      max
    else
      color

  method get_colors n = blend_palette n (truncate r1) (truncate g1) (truncate
  b1) (truncate r2) (truncate g2) (truncate b2)

  method update_velocity velocity want_velocity =
    let new_velocity () = (Random.float 6.0) -. 3.0 in
    let epsilon = 0.1 in
    let acceleration = 0.1 in
    if (absF (!velocity -. !want_velocity)) < epsilon then
      velocity := (new_velocity ())
    else if !velocity > !want_velocity then
      velocity := !velocity -. acceleration
    else if !velocity < !want_velocity then
      velocity := !velocity +. acceleration

  method update =
    r1 <- self#check (r1 +. !r1_velocity) 0.0 r1_max;
    r2 <- self#check (r2 +. !r2_velocity) r2_min 255.0;
    g1 <- self#check (g1 +. !g1_velocity) 0.0 g1_max;
    g2 <- self#check (g2 +. !g2_velocity) g2_min 255.0;
    b1 <- self#check (b1 +. !b1_velocity) 0.0 b1_max;
    b2 <- self#check (b2 +. !b2_velocity) b2_min 255.0;

    self#update_velocity r1_velocity r1_want_velocity;
    self#update_velocity r2_velocity r2_want_velocity;
    self#update_velocity g1_velocity g1_want_velocity;
    self#update_velocity g2_velocity g2_want_velocity;
    self#update_velocity b1_velocity b1_want_velocity;
    self#update_velocity b2_velocity b2_want_velocity;

end;;

type world = {
  stars : star list;
  colors : color_changer;
}

let min_z = 500.0;;
let star_radius = 75.0;;

(* create one star *)
let make_star center_x center_y () =
  let x = (Random.float (star_radius *. 2.0)) -. star_radius in
  let y = (Random.float (star_radius *. 2.0)) -. star_radius in
  let speed = (Random.float 8.0) +. 8.0 in
  let z = (Random.float (min_z -. 20.0)) +. 20.0 in
  {front_z = ref z;
   back_z = ref (z +. (Random.float 20.0) +. 10.0);
   speed = speed;
   center_x = (float center_x);
   center_y = (float center_y);
   x = ref x;
   y = ref y;
   };;

let star_logic (world : world) =
  let update_star (star:star) =
    star.front_z := !(star.front_z) -. star.speed;
    star.back_z := !(star.front_z) -. star.speed;
    if !(star.front_z) < 20.0 then begin
      let x = (Random.float (star_radius *. 2.0)) -. star_radius in
      let y = (Random.float (star_radius *. 2.0)) -. star_radius in
        star.x := x;
        star.y := y;
        star.front_z := min_z;
        star.back_z := min_z;
    end
  in
  List.iter update_star world.stars;
  (world.colors#update);
  ;;

let star_draw world screen =
  let diver = 1024.0 in
  (* project from R^3 to R^2 and translate by `center' *)
  (* let colors = blend_palette (truncate min_z) 0 32 12 220 82 180 in *)
  let colors = (world.colors#get_colors ((truncate min_z) + 20)) in
  let project element z center =
    element *. diver /. z +. center in
  let draw_star star =
    let color (z:float) =
      let c =
        let okz = (truncate z) in
        List.nth colors okz
        (*
        List.nth colors (if okz >= (truncate min_z) then (truncate min_z) - 1
        else okz)
        *)
        (*
        let x = (min_z +. 20.0 -. z) /. 6.0 in
        if x > 255.0 then 255.0 else x
        in
        (Allegro.makecol (truncate c) (truncate c) (truncate c)) in
      *)
      in
      c
    in
    let x1 = project !(star.x) !(star.front_z) star.center_x in
    let y1 = project !(star.y) !(star.front_z) star.center_y in
    let x2 = project !(star.x) !(star.back_z) star.center_x in
    let y2 = project !(star.y) !(star.back_z) star.center_y in
    (*
    Printf.printf "draw line at %f,%f %f,%f %f\n" x1 y1 x2 y2 !(star.front.z);
    flush stdout;
    *)
    ignore(Allegro.line screen (truncate x1) (truncate y1) (truncate x2)
    (truncate y2) (color !(star.front_z)));
  in
  List.iter draw_star world.stars
  ;;

(* make a list of things by invoking `func' a number of times given by `times'
 *)
let make_list (func : unit -> star) times =
  let rec all num =
    match num with
    | 0 -> []
    | n -> (func ()) :: (all (num - 1))
  in
  all times
;;

ignore(init ());
ignore(run {stars = (make_list (make_star (gfx_x / 2) (gfx_y / 2)) max_stars);
            colors = new color_changer;}
           star_logic star_draw)
