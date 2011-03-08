type windowModes = Windowed | FullScreen;;
type mouseButton = MouseButtonLeft | MouseButtonRight

let seconds time = time;;
let milliseconds time = time /. 1000.0;;

type key = KeyA | KeyB | KeyC | KeyUnknown

let index_to_key index =
  match index with
  | 1 -> KeyA
  | 2 -> KeyB
  | 3 -> KeyC
  | _ -> KeyUnknown

  (*
   * 
  __allegro_KEY_A            = 1,
  __allegro_KEY_B            = 2,
  __allegro_KEY_C            = 3,
  __allegro_KEY_D            = 4,
  __allegro_KEY_E            = 5,
  __allegro_KEY_F            = 6,
  __allegro_KEY_G            = 7,
  __allegro_KEY_H            = 8,
  __allegro_KEY_I            = 9,
  __allegro_KEY_J            = 10,
  __allegro_KEY_K            = 11,
  __allegro_KEY_L            = 12,
  __allegro_KEY_M            = 13,
  __allegro_KEY_N            = 14,
  __allegro_KEY_O            = 15,
  __allegro_KEY_P            = 16,
  __allegro_KEY_Q            = 17,
  __allegro_KEY_R            = 18,
  __allegro_KEY_S            = 19,
  __allegro_KEY_T            = 20,
  __allegro_KEY_U            = 21,
  __allegro_KEY_V            = 22,
  __allegro_KEY_W            = 23,
  __allegro_KEY_X            = 24,
  __allegro_KEY_Y            = 25,
  __allegro_KEY_Z            = 26,
  __allegro_KEY_0            = 27,
  __allegro_KEY_1            = 28,
  __allegro_KEY_2            = 29,
  __allegro_KEY_3            = 30,
  __allegro_KEY_4            = 31,
  __allegro_KEY_5            = 32,
  __allegro_KEY_6            = 33,
  __allegro_KEY_7            = 34,
  __allegro_KEY_8            = 35,
  __allegro_KEY_9            = 36,
  __allegro_KEY_0_PAD        = 37,
  __allegro_KEY_1_PAD        = 38,
  __allegro_KEY_2_PAD        = 39,
  __allegro_KEY_3_PAD        = 40,
  __allegro_KEY_4_PAD        = 41,
  __allegro_KEY_5_PAD        = 42,
  __allegro_KEY_6_PAD        = 43,
  __allegro_KEY_7_PAD        = 44,
  __allegro_KEY_8_PAD        = 45,
  __allegro_KEY_9_PAD        = 46,
  __allegro_KEY_F1           = 47,
  __allegro_KEY_F2           = 48,
  __allegro_KEY_F3           = 49,
  __allegro_KEY_F4           = 50,
  __allegro_KEY_F5           = 51,
  __allegro_KEY_F6           = 52,
  __allegro_KEY_F7           = 53,
  __allegro_KEY_F8           = 54,
  __allegro_KEY_F9           = 55,
  __allegro_KEY_F10          = 56,
  __allegro_KEY_F11          = 57,
  __allegro_KEY_F12          = 58,
  __allegro_KEY_ESC          = 59,
  __allegro_KEY_TILDE        = 60,
  __allegro_KEY_MINUS        = 61,
  __allegro_KEY_EQUALS       = 62,
  __allegro_KEY_BACKSPACE    = 63,
  __allegro_KEY_TAB          = 64,
  __allegro_KEY_OPENBRACE    = 65,
  __allegro_KEY_CLOSEBRACE   = 66,
  __allegro_KEY_ENTER        = 67,
  __allegro_KEY_COLON        = 68,
  __allegro_KEY_QUOTE        = 69,
  __allegro_KEY_BACKSLASH    = 70,
  __allegro_KEY_BACKSLASH2   = 71,
  __allegro_KEY_COMMA        = 72,
  __allegro_KEY_STOP         = 73,
  __allegro_KEY_SLASH        = 74,
  __allegro_KEY_SPACE        = 75,
  __allegro_KEY_INSERT       = 76,
  __allegro_KEY_DEL          = 77,
  __allegro_KEY_HOME         = 78,
  __allegro_KEY_END          = 79,
  __allegro_KEY_PGUP         = 80,
  __allegro_KEY_PGDN         = 81,
  __allegro_KEY_LEFT         = 82,
  __allegro_KEY_RIGHT        = 83,
  __allegro_KEY_UP           = 84,
  __allegro_KEY_DOWN         = 85,
  __allegro_KEY_SLASH_PAD    = 86,
  __allegro_KEY_ASTERISK     = 87,
  __allegro_KEY_MINUS_PAD    = 88,
  __allegro_KEY_PLUS_PAD     = 89,
  __allegro_KEY_DEL_PAD      = 90,
  __allegro_KEY_ENTER_PAD    = 91,
  __allegro_KEY_PRTSCR       = 92,
  __allegro_KEY_PAUSE        = 93,
  __allegro_KEY_ABNT_C1      = 94,
  __allegro_KEY_YEN          = 95,
  __allegro_KEY_KANA         = 96,
  __allegro_KEY_CONVERT      = 97,
  __allegro_KEY_NOCONVERT    = 98,
  __allegro_KEY_AT           = 99,
  __allegro_KEY_CIRCUMFLEX   = 100,
  __allegro_KEY_COLON2       = 101,
  __allegro_KEY_KANJI        = 102,
  __allegro_KEY_EQUALS_PAD   = 103,  /* MacOS X */
  __allegro_KEY_BACKQUOTE    = 104,  /* MacOS X */
  __allegro_KEY_SEMICOLON    = 105,  /* MacOS X */
  __allegro_KEY_COMMAND      = 106,  /* MacOS X */
  __allegro_KEY_UNKNOWN1     = 107,
  __allegro_KEY_UNKNOWN2     = 108,
  __allegro_KEY_UNKNOWN3     = 109,
  __allegro_KEY_UNKNOWN4     = 110,
  __allegro_KEY_UNKNOWN5     = 111,
  __allegro_KEY_UNKNOWN6     = 112,
  __allegro_KEY_UNKNOWN7     = 113,
  __allegro_KEY_UNKNOWN8     = 114,
  __allegro_KEY_MODIFIERS    = 115,
  __allegro_KEY_LSHIFT       = 115,
  __allegro_KEY_RSHIFT       = 116,
  __allegro_KEY_LCONTROL     = 117,
  __allegro_KEY_RCONTROL     = 118,
  __allegro_KEY_ALT          = 119,
  __allegro_KEY_ALTGR        = 120,
  __allegro_KEY_LWIN         = 121,
  __allegro_KEY_RWIN         = 122,
  __allegro_KEY_MENU         = 123,
  __allegro_KEY_SCRLOCK      = 124,
  __allegro_KEY_NUMLOCK      = 125,
  __allegro_KEY_CAPSLOCK     = 126,
  __allegro_KEY_MAX          = 127
   *)


let key_name key =
  match key with
  | KeyA -> "a"
  | KeyB -> "b"
  | KeyC -> "c"
  | KeyUnknown -> "unknown"

let float_equals a b =
  abs_float(a -. b) < 0.00001
;;

let mouse_button_name button =
  match button with
  | MouseButtonLeft -> "left button"
  | MouseButtonRight -> "right button"
;;

class virtual eventHandler = object
  method virtual mouse_down: float -> mouseButton -> float -> float -> unit
  method virtual mouse_up: float -> mouseButton -> float -> float -> unit
  method virtual key_down: key -> unit
  method virtual key_up: key -> unit
  method virtual mouse_click: float -> mouseButton -> float -> float -> unit
  method virtual mouse_hover: float -> float -> float -> unit
  method virtual mouse_move: float -> float -> float -> unit
  method virtual keypress: unit
end;;

let now = Unix.gettimeofday;;

class timer (delay: float) (callback: (float -> unit)) = object
  val mutable running = false
  val mutable last_time_fired = 0.0
  val callback = callback
  val delay = delay
  method should_fire time = running && time -. last_time_fired > delay
  method fire time = callback time; last_time_fired <- time
  method start = last_time_fired <- now (); running <- true
  method stop = running <- false
end;;

module type GraphicsSignature = sig
  class graphics: object
    method initialize: windowModes -> int -> int -> unit
    method event_loop: eventHandler -> unit
    method addTimer: float -> (float -> unit) -> timer
    method removeTimer: timer -> bool
		method fillBox: int -> int -> int -> int -> int -> unit
  end
end;;

type mouseButtonState = {held : int; last_pressed_time : float;}
type keyboardState = {keys : bool array}
type mousePosition = {x : float; y: float; last_move_time : float; hovered : bool}
type eventLoopState = {mouse_left_button : mouseButtonState;
                       mouse_right_button : mouseButtonState;
                       mouse_position : mousePosition;
                       keyboard : keyboardState}

module AllegroGraphics: GraphicsSignature = struct
  class graphics = object(self)
    val mutable timers : timer list = []
    method initialize (mode : windowModes) width height =
      Allegro.allegro_init ();
      Allegro.set_color_depth 16;
      Allegro.install_timer ();
      Allegro.install_keyboard ();
      ignore(Allegro.install_mouse ());
      match mode with
      | Windowed -> Allegro.set_gfx_mode Allegro.GFX_AUTODETECT_WINDOWED width height 0 0;
      | FullScreen -> Allegro.set_gfx_mode Allegro.GFX_AUTODETECT_FULLSCREEN width height 0 0
      (*
      Printf.printf "initialize! %d %d\n"
      width height
      *)

		method fillBox (x:int) (y:int) (w:int) (h:int) (rgb:int) =
			();

    method private mouse_left_down () =
      (Allegro.left_button_pressed ())
    
    method private mouse_right_down () =
      (Allegro.right_button_pressed ())

    method private add_timer timer = timers <- timer :: timers

    method addTimer (delay : float) (callback : (float -> unit)) =
      let new_timer = new timer (milliseconds delay) callback in
      ignore(self#add_timer new_timer);
      new_timer#start;
      new_timer

    method removeTimer (timer : timer) =
      let removed = List.exists (fun element -> element == timer) timers in
      timers <- List.filter (fun element -> element != timer) timers;
      (timer#stop);
      removed

    method event_loop (handler : eventHandler) =
      let mouse_delta = milliseconds 100.0 in
      let check_mouse_button checker button state time x y =
        match checker () with
        | true -> begin match state.held with
                        | 0 -> (handler#mouse_down time button x y); {state with held 
                        = 1; last_pressed_time = time}
                        | n -> state
                  end
        | false -> begin match state.held with
                         | 0 -> {state with held = 0; last_pressed_time = 0.0}
                         | n -> (handler#mouse_up time button x y);
                                if time -. state.last_pressed_time < mouse_delta then
                                  (handler#mouse_click time button x y)
                                else
                                  ();
                                  {state with held = 0; last_pressed_time = 0.0}
                         end
      in
      (* TODO: abstract check_left and check_right with currying *)
      let check_right state time = 
        {state with mouse_right_button = check_mouse_button (fun () -> self#mouse_right_down ()) MouseButtonRight state.mouse_right_button time state.mouse_position.x state.mouse_position.y}
      in
      let check_left state time =
        {state with mouse_left_button = check_mouse_button (fun () -> self#mouse_left_down ()) MouseButtonLeft state.mouse_left_button time state.mouse_position.x state.mouse_position.y}
      in
      let check_mouse_position state time =
        let update state =
          let hover_delta = seconds 2.0 in
          let x = float(Allegro.get_mouse_x ()) in
          let y = float(Allegro.get_mouse_y ()) in
          if not (float_equals x state.x) || not (float_equals y state.y) then begin
            handler#mouse_move time x y;
            {x = x; y = y; last_move_time = time; hovered = false};
        end else begin
          if state.hovered = false && time -. state.last_move_time > hover_delta then begin
            (handler#mouse_hover time state.x state.y);
            {state with hovered = true;}
          end else begin
            state
          end
        end
        in {state with mouse_position = update state.mouse_position}
      in
      let check_keyboard_state state time =
        let update_key index state =
          match (state, Allegro.key_is_down index) with
          | true, false -> handler#key_up (index_to_key index); false
          | false, true -> handler#key_down (index_to_key index); true

          (*
          | true, true -> ...
          | false, false -> ...
          | _ -> state
          *)
          | _ -> state
        in
        {state with keyboard = {keys = Array.mapi (fun index state -> update_key index state) state.keyboard.keys}}
        (*
        if Allegro.key_is_down 1 then
          handler#key_down KeyA
        else
          ();
        state
        *)
      in
      let check_timers state time =
        let do_timer timer =
          if timer#should_fire time then
            timer#fire time
          else
            ()
        in
        List.iter do_timer timers;
        state
      in
      let rec loop state : unit =
        Allegro.rest 10;
        let new_time = Unix.gettimeofday () in
        let checks = [check_mouse_position; check_left; check_right; check_keyboard_state; check_timers] in
        loop (List.fold_left (fun old_state check -> check old_state new_time) state 
        checks)
      in
      (loop {mouse_left_button = {held = 0; last_pressed_time = 0.0};
             mouse_right_button = {held = 0; last_pressed_time = 0.0};
             mouse_position = {x = 0.0; y = 0.0; last_move_time = 0.0; hovered = false};
             keyboard = {keys = Array.make 128 false}});
  end
end;;

(*
module WindowSystem = struct
  class virtual widget parent x y z width height = object
    (*
    method virtual draw: (graphics: GraphicsSignature.graphics) -> unit
    method virtual send: unit
    *)

    val parent = parent
    val x = x
    val y = y
    val z = z
    val width = width
    val height = height
  end;;

  class manager = object(self)
    inherit eventHandler
    method mouse_down time button x y = ()
    method mouse_up time button x y = ()
    method key_down key = ()
    method key_up key = ()
    method mouse_click time button x y = ()
    method mouse_hover time x y = ()
    method mouse_move time x y = ()
    method keypress = ()

    method draw (graphics: GraphicsSignature.graphics) = ()

    val mutable widgets: widget list = []
  end;;

  (* type Event = Show | Clicked | ... *)
end;;
  *)

