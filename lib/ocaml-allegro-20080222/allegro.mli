# 1 "allegro.ml.cpp"
# 1 "<built-in>"
# 1 "<command-line>"
# 1 "allegro.ml.cpp"
(* {{{ COPYING *(

 +-------------------------------------------------------------------+
 | OCaml-Allegro, OCaml bindings for the Allegro library. |
 | Copyright (C) 2007 Florent Monnier |
 +-------------------------------------------------------------------+
 | |
 | This program is free software: you can redistribute it and/or |
 | modify it under the terms of the GNU General Public License |
 | as published by the Free Software Foundation, either version 3 |
 | of the License, or (at your option) any later version. |
 | |
 | This program is distributed in the hope that it will be useful, |
 | but WITHOUT ANY WARRANTY; without even the implied warranty of |
 | MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the |
 | GNU General Public License for more details. |
 | |
 | You should have received a copy of the GNU General Public License |
 | along with this program. If not, see |
 | <http://www.gnu.org/licenses/>.                                   |
 +-------------------------------------------------------------------+
 | Author: Florent Monnier <fmonnier@1 -nantes.org> |
 +-------------------------------------------------------------------+

)* }}} *)

(**
  This is an OCaml binding for Allegro.
  Allegro is a cross-platform library intended for use in computer games
  and other types of multimedia programming.
*)

(**
{ul
  {- {{:#glf}General Functions}}
  {- {{:#gfx}Graphics modes}}
  {- {{:#col}Truecolor pixel formats}}
  {- {{:#pal}Palette routines}}
  {- {{:#colconv}Converting between color formats}}
  {- {{:#bmp}Bitmap objects}}
  {- {{:#load}Loading image files}}
  {- {{:#keyboard}Keyboard routines}}
  {- {{:#timer}Timer routines}}
  {- {{:#mouse}Mouse routines}}
  {- {{:#fix}Fixed point number}}
  {- {{:#drawprim}Drawing primitives}}
  {- {{:#sprites}Sprites}}
  {- {{:#rle}RLE Sprites}}
  {- {{:#comp}Compiled Sprites}}
  {- {{:#cout}Text output}}
  {- {{:#transp}Transparency and patterned drawing}}
  {- {{:#sndini}Sound init routines}}
  {- {{:#digispl}Digital sample routines}}
  {- {{:#filefuns}File and compression routines}}
  {- {{:#dat}Datafile routines}}
  {- {{:#poly}Polygon rendering}}
  {- {{:#math3d}3D math routines}}
  {- {{:#quater}Quaternion math routines}}
  {- {{:#gui}GUI routines}}
}
*)

(** {3:glf General Functions} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg000.html}
    Allegro API documentation for this module} *)

external allegro_init: unit -> unit = "ml_allegro_init"
external allegro_exit: unit -> unit = "ml_allegro_exit"
external allegro_message: string -> unit = "ml_allegro_message"

external get_allegro_error: unit -> string = "ml_get_allegro_error"

external allegro_id: unit -> string = "ml_get_allegro_id"

external allegro_version: unit -> int * int * int * string * int * int * int = "ml_alleg_version"
(**
  The three first integers contain the major, middle and minor version numbers.
  The text string contains all version numbers and maybe some additional text (for exemple "4.2.1 (SVN)").
  And the release date of Allegro (year, month, day).
*)

external set_window_title: name:string -> unit = "ml_set_window_title"

external cpu_vendor: unit -> string = "ml_cpu_vendor"

type cpu_family =
  | CPU_FAMILY_UNKNOWN
  | CPU_FAMILY_I386
  | CPU_FAMILY_I486
  | CPU_FAMILY_I586
  | CPU_FAMILY_I686
  | CPU_FAMILY_ITANIUM
  | CPU_FAMILY_POWERPC
  | CPU_FAMILY_EXTENDED
  | CPU_FAMILY_NOT_FOUND

external get_cpu_family: unit -> cpu_family = "ml_get_cpu_family"

external os_version: unit -> int = "ml_os_version"
external os_revision: unit -> int = "ml_os_revision"
external os_multitasking: unit -> bool = "ml_os_multitasking"

external desktop_color_depth: unit -> int = "ml_desktop_color_depth"

external get_desktop_resolution: unit -> int * int = "ml_get_desktop_resolution"

(* TODO
cpu_capabilities
cpu_model
*)



(** {3:gfx Graphics Modes} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg008.html}
    Allegro API documentation for this module} *)

external set_color_depth: depth:int -> unit = "ml_set_color_depth"
external get_color_depth: unit -> int = "ml_get_color_depth"

external get_screen_width: unit -> int = "ml_get_screen_w"
external get_screen_height: unit -> int = "ml_get_screen_h"

external get_virtual_width: unit -> int = "ml_get_virtual_w"
external get_virtual_height: unit -> int = "ml_get_virtual_h"


external request_refresh_rate: rate:int -> unit = "ml_request_refresh_rate"
external get_refresh_rate: unit -> int = "ml_get_refresh_rate"


type gfx_driver =
  | GFX_AUTODETECT
  | GFX_AUTODETECT_FULLSCREEN
  | GFX_AUTODETECT_WINDOWED
  | GFX_SAFE
  | GFX_TEXT

external set_gfx_mode: gfx_driver:gfx_driver ->
    width:int -> height:int -> virtual_width:int -> virtual_height:int -> unit
    = "ml_set_gfx_mode"
(** raises Failure "set_gfx_mode" when it fails *)

external get_gfx_driver_name: unit -> string = "get_gfx_driver_name"

external enable_triple_buffer: unit -> bool = "ml_enable_triple_buffer"
external vsync: unit -> unit = "ml_vsync"
external scroll_screen: x:int -> y:int -> unit = "ml_scroll_screen"
external request_scroll: x:int -> y:int -> unit = "ml_request_scroll"
external poll_scroll: unit -> bool = "ml_poll_scroll"

(* {{{ type gfx_capabilities *)

type gfx_capabilities =
  | GFX_CAN_SCROLL
  | GFX_CAN_TRIPLE_BUFFER
  | GFX_HW_CURSOR
  | GFX_SYSTEM_CURSOR
  | GFX_HW_HLINE
  | GFX_HW_HLINE_XOR
  | GFX_HW_HLINE_SOLID_PATTERN
  | GFX_HW_HLINE_COPY_PATTERN
  | GFX_HW_FILL
  | GFX_HW_FILL_XOR
  | GFX_HW_FILL_SOLID_PATTERN
  | GFX_HW_FILL_COPY_PATTERN
  | GFX_HW_LINE
  | GFX_HW_LINE_XOR
  | GFX_HW_TRIANGLE
  | GFX_HW_TRIANGLE_XOR
  | GFX_HW_GLYPH
  | GFX_HW_VRAM_BLIT
  | GFX_HW_VRAM_BLIT_MASKED
  | GFX_HW_MEM_BLIT
  | GFX_HW_MEM_BLIT_MASKED
  | GFX_HW_SYS_TO_VRAM_BLIT
  | GFX_HW_SYS_TO_VRAM_BLIT_MASKED
  | GFX_HW_VRAM_STRETCH_BLIT
  | GFX_HW_SYS_STRETCH_BLIT
  | GFX_HW_VRAM_STRETCH_BLIT_MASKED
  | GFX_HW_SYS_STRETCH_BLIT_MASKED

(* }}} *)
external get_gfx_capabilities: unit -> gfx_capabilities list = "get_gfx_capabilities"
(*
external gfx_capabilities: unit -> int = "ml_gfx_capabilities"
*)

type switch_mode =
  | SWITCH_NONE
  | SWITCH_PAUSE
  | SWITCH_AMNESIA
  | SWITCH_BACKGROUND
  | SWITCH_BACKAMNESIA

external set_display_switch_mode: switch_mode -> unit = "ml_set_display_switch_mode"
external get_display_switch_mode: unit -> switch_mode = "ml_get_display_switch_mode"

external is_windowed_mode: unit -> bool = "ml_is_windowed_mode"

external allegro_vram_single_surface: unit -> bool = "ml_vram_single_surface"
external allegro_dos: unit -> bool = "ml_allegro_dos"



(** {3:col Truecolor Pixel Formats} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg012.html}
    Allegro API documentation for this module} *)

type color
external makecol: r:int -> g:int -> b:int -> color = "ml_makecol"
external makecol_depth: depth:int -> r:int -> g:int -> b:int -> color = "ml_makecol_depth"

external transparent: unit -> color = "get_transparent"
external makeacol: r:int -> g:int -> b:int -> a:int -> color = "ml_makeacol"
external makeacol_depth: depth:int -> r:int -> g:int -> b:int -> a:int -> color = "ml_makeacol_depth"
external color_index: int -> color = "%identity"

external getr: color -> int = "ml_getr"
external getg: color -> int = "ml_getg"
external getb: color -> int = "ml_getb"
external geta: color -> int = "ml_geta"

external palette_color: int -> color = "ml_palette_color"

(** {4 Display Dependent Pixel Format} *)

external makecol8: r:int -> g:int -> b:int -> color = "ml_makecol8"
external makecol15: r:int -> g:int -> b:int -> color = "ml_makecol15"
external makecol16: r:int -> g:int -> b:int -> color = "ml_makecol16"
external makecol24: r:int -> g:int -> b:int -> color = "ml_makecol24"
external makecol32: r:int -> g:int -> b:int -> color = "ml_makecol32"

external makeacol32: r:int -> g:int -> b:int -> a:int -> color = "ml_makeacol32"

external makecol15_dither: r:int -> g:int -> b:int -> x:int -> y:int -> color = "ml_makecol15_dither"
external makecol16_dither: r:int -> g:int -> b:int -> x:int -> y:int -> color = "ml_makecol16_dither"


external getr8: color -> int = "ml_getr8"
external getg8: color -> int = "ml_getg8"
external getb8: color -> int = "ml_getb8"

external getr15: color -> int = "ml_getr15"
external getg15: color -> int = "ml_getg15"
external getb15: color -> int = "ml_getb15"

external getr16: color -> int = "ml_getr16"
external getg16: color -> int = "ml_getg16"
external getb16: color -> int = "ml_getb16"

external getr24: color -> int = "ml_getr24"
external getg24: color -> int = "ml_getg24"
external getb24: color -> int = "ml_getb24"

external getr32: color -> int = "ml_getr32"
external getg32: color -> int = "ml_getg32"
external getb32: color -> int = "ml_getb32"
external geta32: color -> int = "ml_geta32"


external getr_depth: depth:int -> color -> int = "ml_getr_depth"
external getg_depth: depth:int -> color -> int = "ml_getg_depth"
external getb_depth: depth:int -> color -> int = "ml_getb_depth"
external geta_depth: depth:int -> color -> int = "ml_geta_depth"




(** {3:pal Palette Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg011.html}
    Allegro API documentation for this module} *)

type palette
external new_palette: unit -> palette = "new_palette"
external free_palette: pal:palette -> unit = "free_palette"
(** free palettes got from the new_palette function *)
external set_palette: pal:palette -> unit = "ml_set_palette"
external get_palette: pal:palette -> unit = "ml_get_palette"
external set_color: index:int -> r:int -> g:int -> b:int -> a:int -> unit = "ml_set_color"
external get_desktop_palette: unit -> palette = "ml_get_desktop_palette"
external generate_332_palette: pal:palette -> unit = "ml_generate_332_palette"
external select_palette: pal:palette -> unit = "ml_select_palette"
external unselect_palette: unit -> unit = "ml_unselect_palette"

external palette_set_rgb: pal:palette -> i:int -> r:int -> g:int -> b:int -> unit = "ml_palette_set_rgb"
external palette_set_rgba: pal:palette -> i:int -> r:int -> g:int -> b:int -> a:int -> unit
    = "ml_palette_set_rgba_bytecode"
      "ml_palette_set_rgba_native"
(** sets the index [i] of the palette *)

external palette_get_rgb: pal:palette -> i:int -> int * int * int = "ml_palette_get_rgb"
external palette_get_rgba: pal:palette -> i:int -> int * int * int * int = "ml_palette_get_rgba"
(** gets the index [i] of the palette *)

external palette_copy_index: src:palette -> dst:palette -> i:int -> unit = "ml_palette_copy_index"
(** copy an entry of a palette to another *)

(* TODO
_set_color
black_palette
default_palette
desktop_palette
fade_from
fade_from_range
fade_in
fade_in_range
fade_interpolate
fade_out
fade_out_range
generate_optimized_palette
get_color
get_palette_range
set_palette_range
*)



(** {3:colconv Converting Between Color Formats} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg021.html}
    Allegro API documentation for this module} *)

external hsv_to_rgb: h:float -> s:float -> v:float -> int * int * int = "ml_hsv_to_rgb"
external rgb_to_hsv: r:int -> g:int -> b:int -> float * float * float = "ml_rgb_to_hsv"

(*
TODO
int bestfit_color(const PALETTE pal, int r, int g, int b);
void create_rgb_table(RGB_MAP *table, const PALETTE pal, void ( *callback)(int pos));
*)


(** {3:bmp Bitmap Objects} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg009.html}
    Allegro API documentation for this module} *)

type bitmap
(* external check_bitmap: bmp:bitmap -> unit = "ml_check_bitmap" *)

external get_screen: unit -> bitmap = "ml_get_screen"
external create_bitmap: width:int -> height:int -> bitmap = "ml_create_bitmap"
external create_sub_bitmap: parent:bitmap -> x:int -> y:int -> width:int -> height:int -> bitmap = "ml_create_sub_bitmap"
external create_video_bitmap: width:int -> height:int -> bitmap = "ml_create_video_bitmap"
external create_system_bitmap: width:int -> height:int -> bitmap = "ml_create_system_bitmap"
external create_bitmap_ex: color_depth:int -> width:int -> height:int -> bitmap = "ml_create_bitmap_ex"
external bitmap_color_depth: bmp:bitmap -> int = "ml_bitmap_color_depth"
external destroy_bitmap: bmp:bitmap -> unit = "ml_destroy_bitmap"
external get_bitmap_width: bmp:bitmap -> int = "get_bitmap_width"
external get_bitmap_height: bmp:bitmap -> int = "get_bitmap_height"
external acquire_bitmap: bmp:bitmap -> unit = "ml_acquire_bitmap"
(* Locks the bitmap before drawing onto it. *)
external release_bitmap: bmp:bitmap -> unit = "ml_release_bitmap"
external acquire_screen: unit -> unit = "ml_acquire_screen"
external release_screen: unit -> unit = "ml_release_screen"

external bitmap_mask_color: bmp:bitmap -> color = "ml_bitmap_mask_color"
external is_same_bitmap: bmp1:bitmap -> bmp2:bitmap -> bool = "ml_is_same_bitmap"

external show_video_bitmap: bmp:bitmap -> unit = "ml_show_video_bitmap"
external request_video_bitmap: bmp:bitmap -> unit = "ml_request_video_bitmap"

external clear_bitmap: bmp:bitmap -> unit = "ml_clear_bitmap"
external clear_to_color: bmp:bitmap -> color:color -> unit = "ml_clear_to_color"
external add_clip_rect: bmp:bitmap -> x1:int -> y1:int -> x2:int -> y2:int -> unit = "ml_add_clip_rect"
external set_clip_rect: bmp:bitmap -> x1:int -> y1:int -> x2:int -> y2:int -> unit = "ml_set_clip_rect"
external set_clip_state: bmp:bitmap -> bool -> unit = "ml_set_clip_state"
(* Turns on or off the clipping of a bitmap. *)
external get_clip_state: bmp:bitmap -> bool = "ml_get_clip_state"

external is_planar_bitmap: bmp:bitmap -> bool = "ml_is_planar_bitmap"
external is_linear_bitmap: bmp:bitmap -> bool = "ml_is_linear_bitmap"
external is_memory_bitmap: bmp:bitmap -> bool = "ml_is_memory_bitmap"
external is_screen_bitmap: bmp:bitmap -> bool = "ml_is_screen_bitmap"
external is_video_bitmap: bmp:bitmap -> bool = "ml_is_video_bitmap"
external is_system_bitmap: bmp:bitmap -> bool = "ml_is_system_bitmap"
external is_sub_bitmap: bmp:bitmap -> bool = "ml_is_sub_bitmap"

(*
get_clip_rect : Returns the clipping rectangle of a bitmap.

is_inside_bitmap : Tells if a point is inside a bitmap.

lock_bitmap : Locks the memory used by a bitmap.
screen_h : Global define to obtain the size of the screen.
screen_w : Global define to obtain the size of the screen.
virtual_h : Global define to obtain the virtual size of the screen.
virtual_w : Global define to obtain the virtual size of the screen.
*)



(** {3:load Loading Image Files} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg010.html}
    Allegro API documentation for this module} *)

external load_bitmap: string -> palette -> bitmap = "ml_load_bitmap"
external save_bitmap: string -> bitmap -> bool = "ml_save_bitmap"
external load_bmp: string -> palette -> bitmap = "ml_load_bmp"
external load_lbm: string -> palette -> bitmap = "ml_load_lbm"
external load_pcx: string -> palette -> bitmap = "ml_load_pcx"
external load_tga: string -> palette -> bitmap = "ml_load_tga"


external blit: src:bitmap -> dest:bitmap ->
    src_x:int -> src_y:int -> dest_x:int -> dest_y:int -> width:int -> height:int -> unit
    = "ml_blit_bytecode"
      "ml_blit_native"

external masked_blit: src:bitmap -> dest:bitmap ->
    src_x:int -> src_y:int -> dest_x:int -> dest_y:int -> width:int -> height:int -> unit
    = "ml_masked_blit_bytecode"
      "ml_masked_blit_native"

external stretch_blit: src:bitmap -> dest:bitmap ->
    src_x:int -> src_y:int -> src_width:int -> src_height:int ->
    dest_x:int -> dest_y:int -> dest_width:int -> dest_height:int -> unit
    = "ml_stretch_blit_bytecode"
      "ml_stretch_blit_native"

external masked_stretch_blit: src:bitmap -> dest:bitmap ->
    src_x:int -> src_y:int -> src_width:int -> src_height:int ->
    dest_x:int -> dest_y:int -> dest_width:int -> dest_height:int -> unit
    = "ml_masked_stretch_blit_bytecode"
      "ml_masked_stretch_blit_native"


type color_conversion =
  | COLORCONV_NONE
  | COLORCONV_8_TO_15
  | COLORCONV_8_TO_16
  | COLORCONV_8_TO_24
  | COLORCONV_8_TO_32
  | COLORCONV_15_TO_8
  | COLORCONV_15_TO_16
  | COLORCONV_15_TO_24
  | COLORCONV_15_TO_32
  | COLORCONV_16_TO_8
  | COLORCONV_16_TO_15
  | COLORCONV_16_TO_24
  | COLORCONV_16_TO_32
  | COLORCONV_24_TO_8
  | COLORCONV_24_TO_15
  | COLORCONV_24_TO_16
  | COLORCONV_24_TO_32
  | COLORCONV_32_TO_8
  | COLORCONV_32_TO_15
  | COLORCONV_32_TO_16
  | COLORCONV_32_TO_24
  | COLORCONV_32A_TO_8
  | COLORCONV_32A_TO_15
  | COLORCONV_32A_TO_16
  | COLORCONV_32A_TO_24
  | COLORCONV_DITHER_PAL
  | COLORCONV_DITHER_HI
  | COLORCONV_KEEP_TRANS

  | COLORCONV_EXPAND_256
  | COLORCONV_REDUCE_TO_256
  | COLORCONV_EXPAND_15_TO_16
  | COLORCONV_REDUCE_16_TO_15
  | COLORCONV_EXPAND_HI_TO_TRUE
  | COLORCONV_REDUCE_TRUE_TO_HI
  | COLORCONV_24_EQUALS_32
  | COLORCONV_TOTAL
  | COLORCONV_PARTIAL
  | COLORCONV_MOST
  | COLORCONV_DITHER
  (*
  | COLORCONV_KEEP_ALPHA
  *)

external set_color_conversion: color_conversion -> unit = "ml_set_color_conversion"
external get_color_conversion: unit -> color_conversion = "ml_get_color_conversion"


(** {3:keyboard Keyboard Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg006.html}
    Allegro API documentation for this module} *)

external install_keyboard: unit -> unit = "ml_install_keyboard"
(*
TODO
void set_config_string(const char *section, const char *name, const char *val);
*)

external poll_keyboard: unit -> unit = "ml_poll_keyboard"
(*
TODO
keyboard_needs_poll()
*)


external keypressed: unit -> bool = "ml_keypressed"
external readkey: unit -> char = "ml_readkey"

external key_esc: unit -> bool = "ml_key_key_esc"
external key_enter: unit -> bool = "ml_key_key_enter"
external key_left: unit -> bool = "ml_key_key_left"
external key_right: unit -> bool = "ml_key_key_right"
external key_up: unit -> bool = "ml_key_key_up"
external key_down: unit -> bool = "ml_key_key_down"

type kb_flag =
  | KB_SHIFT_FLAG
  | KB_CTRL_FLAG
  | KB_ALT_FLAG
  | KB_LWIN_FLAG
  | KB_RWIN_FLAG
  | KB_MENU_FLAG
  | KB_COMMAND_FLAG
  | KB_SCROLOCK_FLAG
  | KB_NUMLOCK_FLAG
  | KB_CAPSLOCK_FLAG
  | KB_INALTSEQ_FLAG
  | KB_ACCENT1_FLAG
  | KB_ACCENT2_FLAG
  | KB_ACCENT3_FLAG
  | KB_ACCENT4_FLAG

external get_kb_flag: kb_flag -> bool = "get_kb_flag"



(* {{{ type scancode *)

type scancode =
  | KEY_A | KEY_B | KEY_C | KEY_D | KEY_E | KEY_F
  | KEY_G | KEY_H | KEY_I | KEY_J | KEY_K
  | KEY_L | KEY_M | KEY_N | KEY_O | KEY_P | KEY_Q
  | KEY_R | KEY_S | KEY_T | KEY_U | KEY_V | KEY_W
  | KEY_X | KEY_Y | KEY_Z

  | KEY_0 | KEY_1 | KEY_2 | KEY_3 | KEY_4
  | KEY_5 | KEY_6 | KEY_7 | KEY_8 | KEY_9

  | KEY_0_PAD | KEY_1_PAD | KEY_2_PAD | KEY_3_PAD
  | KEY_4_PAD | KEY_5_PAD | KEY_6_PAD | KEY_7_PAD
  | KEY_8_PAD | KEY_9_PAD

  | KEY_F1 | KEY_F2 | KEY_F3 | KEY_F4 | KEY_F5
  | KEY_F6 | KEY_F7 | KEY_F8 | KEY_F9
  | KEY_F10 | KEY_F11 | KEY_F12

  | KEY_ESC | KEY_TILDE | KEY_MINUS | KEY_EQUALS
  | KEY_BACKSPACE | KEY_TAB | KEY_OPENBRACE | KEY_CLOSEBRACE
  | KEY_ENTER | KEY_COLON | KEY_QUOTE | KEY_BACKSLASH
  | KEY_BACKSLASH2 | KEY_COMMA | KEY_STOP | KEY_SLASH
  | KEY_SPACE

  | KEY_INSERT | KEY_DEL | KEY_HOME | KEY_END | KEY_PGUP
  | KEY_PGDN | KEY_LEFT | KEY_RIGHT | KEY_UP | KEY_DOWN

  | KEY_SLASH_PAD | KEY_ASTERISK | KEY_MINUS_PAD
  | KEY_PLUS_PAD | KEY_DEL_PAD | KEY_ENTER_PAD

  | KEY_PRTSCR | KEY_PAUSE

  | KEY_ABNT_C1 | KEY_YEN | KEY_KANA | KEY_CONVERT | KEY_NOCONVERT
  | KEY_AT | KEY_CIRCUMFLEX | KEY_COLON2 | KEY_KANJI

  | KEY_LSHIFT | KEY_RSHIFT
  | KEY_LCONTROL | KEY_RCONTROL
  | KEY_ALT | KEY_ALTGR
  | KEY_LWIN | KEY_RWIN | KEY_MENU
  | KEY_SCRLOCK | KEY_NUMLOCK | KEY_CAPSLOCK

  | KEY_EQUALS_PAD | KEY_BACKQUOTE | KEY_SEMICOLON | KEY_COMMAND

(* }}} *)
external readkey_scancode: unit -> scancode = "ml_readkey_scancode"

external clear_keybuf: unit -> unit = "ml_clear_keybuf"
external set_keyboard_rate: delay:int -> repeat:int -> unit = "ml_set_keyboard_rate"

external remove_keyboard: unit -> unit = "ml_remove_keyboard"


(** {3:timer Timer Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg005.html}
    Allegro API documentation for this module} *)

external install_timer: unit -> unit = "ml_install_timer"
external retrace_count: unit -> int = "get_retrace_count"
external rest: time:int -> unit = "ml_rest"



type cb_id
val install_param_int: ms:int -> cb:(param:'a -> unit) -> param:'a -> cb_id
(** returns the id of the callback which can be used with [remove_param_int] *)
val remove_param_int: id:cb_id -> unit
# 638 "allegro.ml.cpp"
(** {3:mouse Mouse Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg004.html}
    Allegro API documentation for this module} *)

external install_mouse: unit -> int = "ml_install_mouse"
external poll_mouse: unit -> unit = "ml_poll_mouse"

external mouse_driver_name: unit -> string = "get_mouse_driver_name"

external enable_hardware_cursor: unit -> unit = "ml_enable_hardware_cursor"
external disable_hardware_cursor: unit -> unit = "ml_disable_hardware_cursor"

type mouse_cursor =
  | MOUSE_CURSOR_NONE
  | MOUSE_CURSOR_ALLEGRO
  | MOUSE_CURSOR_ARROW
  | MOUSE_CURSOR_BUSY
  | MOUSE_CURSOR_QUESTION
  | MOUSE_CURSOR_EDIT

external select_mouse_cursor: mouse_cursor -> unit = "ml_select_mouse_cursor"
external set_mouse_cursor_bitmap: mouse_cursor -> bmp:bitmap -> unit = "ml_set_mouse_cursor_bitmap"
external show_mouse: bmp:bitmap -> unit = "ml_show_mouse"
external hide_mouse: unit -> unit = "ml_hide_mouse"
external scare_mouse: unit -> unit = "ml_scare_mouse"
external scare_mouse_area: x:int -> y:int -> w:int -> h:int -> unit = "ml_scare_mouse_area"
external unscare_mouse: unit -> unit = "ml_unscare_mouse"

external position_mouse: x:int -> y:int -> unit = "ml_position_mouse"
external position_mouse_z: z:int -> unit = "ml_position_mouse_z"
(*
external position_mouse_w: w:int -> unit = "ml_position_mouse_w"
*)
external set_mouse_range: x1:int -> y1:int -> x2:int -> y2:int -> unit = "ml_set_mouse_range"
external set_mouse_speed: xspeed:int -> yspeed:int -> unit = "ml_set_mouse_speed"


external set_mouse_sprite: sprite:bitmap -> unit = "ml_set_mouse_sprite"
external set_mouse_sprite_focus: x:int -> y:int -> unit = "ml_set_mouse_sprite_focus"

external get_mouse_mickeys: unit -> int * int = "ml_get_mouse_mickeys"

external get_mouse_x: unit -> int = "ml_mouse_x"
external get_mouse_y: unit -> int = "ml_mouse_y"
external get_mouse_z: unit -> int = "ml_mouse_z"
(*
external get_mouse_w: unit -> int = "ml_mouse_w"
*)

external get_mouse_x_focus: unit -> int = "ml_mouse_x_focus"
external get_mouse_y_focus: unit -> int = "ml_mouse_y_focus"

external left_button_pressed: unit -> bool = "left_button_pressed"
external right_button_pressed: unit -> bool = "right_button_pressed"
external middle_button_pressed: unit -> bool = "middle_button_pressed"

external get_mouse_b: unit -> int = "get_mouse_b"

external get_mouse_pos: unit -> int * int = "get_mouse_pos"

(*
TODO
extern void ( *mouse_callback)(int flags);
*)


(** {3:fix Fixed Point Number} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg032.html}
    Allegro API documentation for this module} *)

type fixed
external itofix: int -> fixed = "ml_itofix"
external fixtoi: fixed -> int = "ml_fixtoi"
external ftofix: float -> fixed = "ml_ftofix"
external fixtof: fixed -> float = "ml_fixtof"


external fixadd: fixed -> fixed -> fixed = "ml_fixadd"
external fixsub: fixed -> fixed -> fixed = "ml_fixsub"
external fixdiv: fixed -> fixed -> fixed = "ml_fixdiv"
external fixmul: fixed -> fixed -> fixed = "ml_fixmul"
external fixhypot: fixed -> fixed -> fixed = "ml_fixhypot"

external fixceil : fixed -> int = "ml_fixceil"
external fixfloor: fixed -> int = "ml_fixfloor"

external fixsin : fixed -> fixed = "ml_fixsin"
external fixcos : fixed -> fixed = "ml_fixcos"
external fixtan : fixed -> fixed = "ml_fixtan"
external fixasin: fixed -> fixed = "ml_fixasin"
external fixacos: fixed -> fixed = "ml_fixacos"
external fixatan: fixed -> fixed = "ml_fixatan"
external fixsqrt: fixed -> fixed = "ml_fixsqrt"

external to_rad: fixed -> fixed = "ml_fixtorad"
external of_rad: fixed -> fixed = "ml_fixofrad"

external fixminus: fixed -> fixed = "ml_fixminus"

external fixatan2: y:fixed -> x:fixed -> fixed = "ml_fixatan2"



(** {3:drawprim Drawing Primitives} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg013.html}
    Allegro API documentation for this module} *)

external putpixel: bmp:bitmap -> x:int -> y:int -> color:color -> unit = "ml_putpixel"

external rect: bmp:bitmap -> x1:int -> y1:int -> x2:int -> y2:int -> color:color -> unit
    = "ml_rect_bytecode"
      "ml_rect_native"
external rectfill: bmp:bitmap -> x1:int -> y1:int -> x2:int -> y2:int -> color:color -> unit
    = "ml_rectfill_bytecode"
      "ml_rectfill_native"

external arc: bmp:bitmap -> x:int -> y:int -> angle1:fixed -> angle2:fixed -> r:int -> color:color -> unit
    = "ml_arc_bytecode"
      "ml_arc_native"

external floodfill: bmp:bitmap -> x:int -> y:int -> color:color -> unit = "ml_floodfill"

external spline: bmp:bitmap ->
    x1:int -> y1:int -> x2:int -> y2:int -> x3:int -> y3:int -> x4:int -> y4:int -> color:color -> unit
    = "ml_spline_bytecode"
      "ml_spline_native"


external circle: bmp:bitmap -> x:int -> y:int -> radius:int -> color:color -> unit = "ml_circle"
external circlefill: bmp:bitmap -> x:int -> y:int -> radius:int -> color:color -> unit = "ml_circlefill"

external ellipse: bmp:bitmap -> x:int -> y:int -> rx:int -> ry:int -> color:color -> unit
    = "ml_ellipse_bytecode"
      "ml_ellipse_native"

external ellipsefill: bmp:bitmap -> x:int -> y:int -> rx:int -> ry:int -> color:color -> unit
    = "ml_ellipsefill_bytecode"
      "ml_ellipsefill_native"

external triangle: bmp:bitmap -> x1:int -> y1:int -> x2:int -> y2:int -> x3:int -> y3:int -> color:color -> unit
    = "ml_triangle_bytecode"
      "ml_triangle_native"

external line: bmp:bitmap -> x1:int -> y1:int -> x2:int -> y2:int -> color:color -> unit
    = "ml_line_bytecode"
      "ml_line_native"

external fastline: bmp:bitmap -> x1:int -> y1:int -> x2:int -> y2:int -> color:color -> unit
    = "ml_fastline_bytecode"
      "ml_fastline_native"


external vline: bmp:bitmap -> x:int -> y1:int -> y2:int -> color:color -> unit = "ml_vline"
external hline: bmp:bitmap -> x1:int -> y:int -> x2:int -> color:color -> unit = "ml_hline"

external getpixel: bmp:bitmap -> x:int -> y:int -> color = "ml_getpixel"

(* {{{ do_circle *)


val do_circle: bmp:bitmap -> x:int -> y:int -> radius:int -> d:int ->
      proc:(bmp:bitmap -> x:int -> y:int -> d:int -> unit) -> unit
# 815 "allegro.ml.cpp"
(* }}} *)
(* {{{ do_ellipse *)


val do_ellipse: bmp:bitmap -> x:int -> y:int -> rx:int -> ry:int -> d:int ->
      proc:(bmp:bitmap -> x:int -> y:int -> d:int -> unit) -> unit
# 835 "allegro.ml.cpp"
(* }}} *)

(* TODO
calc_spline
*)



(** {3:sprites Sprites} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg014.html}
    Allegro API documentation for this module} *)

external draw_sprite: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> unit = "ml_draw_sprite"
external draw_sprite_v_flip: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> unit = "draw_sprite_v_flip"
external draw_sprite_h_flip: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> unit = "draw_sprite_h_flip"
external draw_sprite_vh_flip: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> unit = "draw_sprite_vh_flip"


external rotate_sprite: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> angle:fixed -> unit = "ml_rotate_sprite"
external rotate_sprite_v_flip: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> angle:fixed -> unit = "ml_rotate_sprite_v_flip"
external rotate_scaled_sprite: bmp:bitmap -> sprite:bitmap ->
    x:int -> y:int -> angle:fixed -> scale:fixed -> unit
    = "ml_rotate_scaled_sprite_bytecode"
      "ml_rotate_scaled_sprite_native"

external rotate_scaled_sprite_v_flip: bmp:bitmap -> sprite:bitmap ->
    x:int -> y:int -> angle:fixed -> scale:fixed -> unit
    = "ml_rotate_scaled_sprite_v_flip_bytecode"
      "ml_rotate_scaled_sprite_v_flip_native"


external pivot_sprite: bmp:bitmap -> sprite:bitmap ->
    x:int -> y:int -> cx:int -> cy:int -> angle:fixed -> unit
    = "ml_pivot_sprite_bytecode"
      "ml_pivot_sprite_native"

external pivot_sprite_v_flip: bmp:bitmap -> sprite:bitmap ->
    x:int -> y:int -> cx:int -> cy:int -> angle:fixed -> unit
    = "ml_pivot_sprite_v_flip_bytecode"
      "ml_pivot_sprite_v_flip_native"

external pivot_scaled_sprite: bmp:bitmap -> sprite:bitmap ->
    x:int -> y:int -> cx:int -> cy:int -> angle:fixed -> scale:fixed -> unit
    = "ml_pivot_scaled_sprite_bytecode"
      "ml_pivot_scaled_sprite_native"

external pivot_scaled_sprite_v_flip: bmp:bitmap -> sprite:bitmap ->
    x:int -> y:int -> cx:int -> cy:int -> angle:fixed -> scale:fixed -> unit
    = "ml_pivot_scaled_sprite_v_flip_bytecode"
      "ml_pivot_scaled_sprite_v_flip_native"


external stretch_sprite: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> w:int -> h:int -> unit
    = "ml_stretch_sprite_bytecode"
      "ml_stretch_sprite_native"

external draw_character_ex: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> color:color -> bg:color -> unit
    = "ml_draw_character_ex_bytecode"
      "ml_draw_character_ex_native"

external draw_lit_sprite: bmp:bitmap -> sprite:bitmap ->
    x:int -> y:int -> color:color -> unit = "ml_draw_lit_sprite"

external draw_trans_sprite: bmp:bitmap -> sprite:bitmap -> x:int -> y:int -> unit = "ml_draw_trans_sprite"

external draw_gouraud_sprite: bmp:bitmap -> sprite:bitmap -> x:int -> y:int ->
    c1:color -> c2:color -> c3:color -> c4:color -> unit
    = "ml_draw_gouraud_sprite_bytecode"
      "ml_draw_gouraud_sprite_native"



(** {3:rle RLE Sprites} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg015.html}
    Allegro API documentation for this module} *)

type rle_sprite

external get_rle_sprite: tmp:bitmap -> rle_sprite = "ml_get_rle_sprite"
external destroy_rle_sprite: rle_sprite -> unit = "ml_destroy_rle_sprite"
external draw_rle_sprite: bmp:bitmap -> sprite:rle_sprite -> x:int -> y:int -> unit = "ml_draw_rle_sprite"
external draw_trans_rle_sprite: bmp:bitmap -> sprite:rle_sprite ->
    x:int -> y:int -> unit = "ml_draw_trans_rle_sprite"
external draw_lit_rle_sprite: bmp:bitmap -> sprite:rle_sprite ->
    x:int -> y:int -> color:color -> unit = "ml_draw_lit_rle_sprite"


(** {3:comp Compiled Sprites} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg016.html}
    Allegro API documentation for this module} *)

type compiled_sprite

external get_compiled_sprite: bmp:bitmap -> planar:bool -> compiled_sprite = "ml_get_compiled_sprite"
external destroy_compiled_sprite: sprite:compiled_sprite -> unit = "ml_destroy_compiled_sprite"
external draw_compiled_sprite: bmp:bitmap -> sprite:compiled_sprite -> x:int -> y:int -> unit = "ml_draw_compiled_sprite"


(** {3:cout Text Output} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg018.html}
    Allegro API documentation for this module} *)

type font

external get_font: unit -> font = "get_font"
external allegro_404_char: char -> unit = "ml_allegro_404_char"
external text_length: f:font -> string -> int = "ml_text_length"
external text_height: f:font -> int = "ml_text_height"

external textout_ex: bmp:bitmap -> f:font -> str:string ->
    x:int -> y:int -> color:color -> bg:color -> unit
    = "ml_textout_ex_bytecode"
      "ml_textout_ex_native"

external textout_centre_ex: bmp:bitmap -> f:font -> str:string ->
    x:int -> y:int -> color:color -> bg:color -> unit
    = "ml_textout_centre_ex_bytecode"
      "ml_textout_centre_ex_native"

external textout_justify_ex: bmp:bitmap -> f:font -> str:string ->
    x1:int -> x2:int -> y:int -> diff:int -> color:color -> bg:color -> unit
    = "ml_textout_justify_ex_bytecode"
      "ml_textout_justify_ex_native"

external textout_right_ex: bmp:bitmap -> f:font -> str:string ->
    x:int -> y:int -> color:color -> bg:color -> unit
    = "ml_textout_right_ex_bytecode"
      "ml_textout_right_ex_native"

(* {{{ textprintf_ex *)


(*
val textprintf_ex: bmp:bitmap -> f:font ->
    x:int -> y:int -> color:color -> bg:color ->
    ('a, unit, string, unit) format4 -> 'a
*)
val textprintf_ex: bitmap -> font ->
    int -> int -> color -> color ->
    ('a, unit, string, unit) format4 -> 'a
# 998 "allegro.ml.cpp"
(* }}} *)
(* {{{ textprintf_centre_ex *)


val textprintf_centre_ex: bitmap -> font ->
    int -> int -> color -> color ->
    ('a, unit, string, unit) format4 -> 'a
# 1015 "allegro.ml.cpp"
(* }}} *)
(* {{{ textprintf_justify_ex *)


val textprintf_justify_ex: bitmap -> font ->
    int -> int -> int -> int -> color -> color ->
    ('a, unit, string, unit) format4 -> 'a
# 1032 "allegro.ml.cpp"
(* }}} *)
(* {{{ textprintf_right_ex *)


val textprintf_right_ex: bitmap -> font ->
    int -> int -> color -> color ->
    ('a, unit, string, unit) format4 -> 'a
# 1049 "allegro.ml.cpp"
(* }}} *)


(** {3:fonts Fonts} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg017.html}
    Allegro API documentation for this module} *)

external load_font: filename:string -> font * palette = "ml_load_font"
(* XXX
external extract_font_range: f:font -> start:int -> finish:int -> font = "ml_extract_font_range"
*)
external merge_fonts: f1:font -> f2:font -> font = "ml_merge_fonts"
external destroy_font: f:font -> unit = "ml_destroy_font"


(** {3:transp Transparency and Patterned Drawing} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg020.html}
    Allegro API documentation for this module} *)

(* {{{ drawing_mode *)


type draw_mode =
  | DRAW_MODE_SOLID (* the default, solid color drawing *)
  | DRAW_MODE_XOR (* exclusive-or drawing *)
  | DRAW_MODE_COPY_PATTERN of bitmap * int * int (* multicolored pattern fill *)
  | DRAW_MODE_SOLID_PATTERN of bitmap * int * int (* single color pattern fill *)
  | DRAW_MODE_MASKED_PATTERN of bitmap * int * int (* masked pattern fill *)
  | DRAW_MODE_TRANS (* translucent color blending *)

val drawing_mode: draw_mode:draw_mode -> unit
# 1116 "allegro.ml.cpp"
(* }}} *)

external xor_mode: on:bool -> unit = "ml_xor_mode"
external solid_mode: unit -> unit = "ml_solid_mode"
external set_trans_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_trans_blender"
external set_alpha_blender: unit -> unit = "ml_set_alpha_blender"
external set_write_alpha_blender: unit -> unit = "ml_set_write_alpha_blender"
external set_add_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_add_blender"
external set_burn_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_burn_blender"
external set_color_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_color_blender"
external set_difference_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_difference_blender"
external set_dissolve_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_dissolve_blender"
external set_dodge_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_dodge_blender"
external set_hue_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_hue_blender"
external set_invert_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_invert_blender"
external set_luminance_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_luminance_blender"
external set_multiply_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_multiply_blender"
external set_saturation_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_saturation_blender"
external set_screen_blender: r:int -> g:int -> b:int -> a:int -> unit = "ml_set_screen_blender"


external digi_driver_name: unit -> string = "ml_digi_driver_name"

(*
TODO
void set_blender_mode(BLENDER_FUNC b15, b16, b24, int r, g, b, a);
void set_blender_mode_ex(BLENDER_FUNC b15, b16, b24, b32, b15x, b16x, b24x, int r, g, b, a);

void create_blender_table(COLOR_MAP *table, const PALETTE pal, void ( *callback)(int pos));
void create_color_table(COLOR_MAP *table, const PALETTE pal,
                        void ( *blend)(PALETTE pal, int x, int y, RGB *rgb),
                        void ( *callback)(int pos));
*)


(** {3:sndini Sound Init Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg024.html}
    Allegro API documentation for this module} *)

type digi =
  | DIGI_AUTODETECT
  | DIGI_NONE

type midi =
  | MIDI_AUTODETECT
  | MIDI_NONE

external install_sound: digi -> midi -> unit = "ml_install_sound"
external remove_sound: unit -> unit = "ml_remove_sound"
external reserve_voices: digi_voices:int -> midi_voices:int -> unit = "ml_reserve_voices"
external set_volume_per_voice: scale:int -> unit = "ml_set_volume_per_voice"
external set_volume: digi_volume:int -> midi_volume:int -> unit = "ml_set_volume"
external set_hardware_volume: digi_volume:int -> midi_volume:int -> unit = "ml_set_hardware_volume"
(*
external get_volume: unit -> int * int = "ml_get_volume"
external get_hardware_volume: unit -> int * int = "ml_get_hardware_volume"
*)

(*
TODO
int detect_digi_driver(int driver_id);
int detect_midi_driver(int driver_id);
*)



(** {3:digispl Digital Sample Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg026.html}
    Allegro API documentation for this module} *)

type sample
external load_sample: filename:string -> sample = "ml_load_sample"
external destroy_sample: spl:sample -> unit = "ml_destroy_sample"
external adjust_sample: spl:sample -> vol:int -> pan:int -> freq:int -> loop:bool -> unit = "ml_adjust_sample"
external play_sample: spl:sample -> vol:int -> pan:int -> freq:int -> loop:bool -> int = "ml_play_sample"
external stop_sample: spl:sample -> unit = "ml_stop_sample"


(* * TODO {3:astrm Audio Stream Routines} *)


(** {3:filefuns File and Compression Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg030.html}
    Allegro API documentation for this module} *)

external replace_filename: path:string -> filename:string -> string = "ml_replace_filename"
(*
PACKFILE *pack_fopen(const char *filename, const char *mode);
int pack_fclose(PACKFILE *f);
long pack_fwrite(const void *p, long n, PACKFILE *f);
*)



(** {3:dat Datafile Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg031.html}
    Allegro API documentation for this module} *)

type datafile
external load_datafile: filename:string -> datafile = "ml_load_datafile"
external unload_datafile: dat:datafile -> unit = "ml_unload_datafile"
external fixup_datafile: dat:datafile -> unit = "ml_fixup_datafile"

external palette_dat: dat:datafile -> idx:int -> palette = "datafile_index"
external bitmap_dat : dat:datafile -> idx:int -> bitmap = "datafile_index"
external font_dat : dat:datafile -> idx:int -> font = "datafile_index"
external sample_dat : dat:datafile -> idx:int -> sample = "datafile_index"
external item_dat : dat:datafile -> idx:int -> 'a      = "datafile_index"


(** {3:gui GUI Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg035.html}
    Allegro API documentation for this module} *)

external gfx_mode_select_ex: unit -> gfx_driver * int * int * int = "ml_gfx_mode_select_ex"


(*
TODO
int alert(const char *s1, *s2, *s3, const char *b1, *b2, int c1, c2);
*)



(** {3:poly Polygon Rendering} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg019.html}
    Allegro API documentation for this module} *)

type polytype =
  | POLYTYPE_ATEX
  | POLYTYPE_ATEX_LIT
  | POLYTYPE_ATEX_MASK
  | POLYTYPE_ATEX_MASK_LIT
  | POLYTYPE_ATEX_MASK_TRANS
  | POLYTYPE_ATEX_TRANS
  | POLYTYPE_FLAT
  | POLYTYPE_GCOL
  | POLYTYPE_GRGB
  | POLYTYPE_PTEX
  | POLYTYPE_PTEX_LIT
  | POLYTYPE_PTEX_MASK
  | POLYTYPE_PTEX_MASK_LIT
  | POLYTYPE_PTEX_MASK_TRANS
  | POLYTYPE_PTEX_TRANS

type 'a _v3d = {
    x: 'a; y: 'a; z: 'a;  (** position *)
    u: 'a; v: 'a; (** texture map coordinates *)
    c:color; (** color *)
  }

type v3d = fixed _v3d
type v3d_f = float _v3d


external triangle3d: bmp:bitmap -> polytype:polytype -> tex:bitmap ->
    v1:v3d -> v2:v3d -> v3:v3d -> unit
    = "ml_triangle3d_bytecode"
      "ml_triangle3d_native"

external triangle3d_f: bmp:bitmap -> polytype:polytype -> tex:bitmap ->
    v1:v3d_f -> v2:v3d_f -> v3:v3d_f -> unit
    = "ml_triangle3d_f_bytecode"
      "ml_triangle3d_f_native"

external quad3d: bmp:bitmap -> polytype:polytype -> tex:bitmap ->
    v1:v3d -> v2:v3d -> v3:v3d -> v4:v3d -> unit
    = "ml_quad3d_bytecode"
      "ml_quad3d_native"

external quad3d_f: bmp:bitmap -> polytype:polytype -> tex:bitmap ->
    v1:v3d_f -> v2:v3d_f -> v3:v3d_f -> v4:v3d_f -> unit
    = "ml_quad3d_f_bytecode"
      "ml_quad3d_f_native"

external clear_scene: bmp:bitmap -> unit = "ml_clear_scene"
external render_scene: unit -> unit = "ml_render_scene"
external create_scene: nedge:int -> npoly:int -> int = "ml_create_scene"
external destroy_scene: unit -> unit = "ml_destroy_scene"

(* TODO
int clip3d_f(int type, float min_z, float max_z, int vc, const V3D_f *vtx[], V3D_f *vout[], V3D_f *vtmp[], int out[]);
int clip3d(int type, fixed min_z, fixed max_z, int vc, const V3D *vtx[], V3D *vout[], V3D *vtmp[], int out[]);

int scene_polygon3d(int type, BITMAP *texture, int vc, V3D *vtx[]);
int scene_polygon3d_f(int type, BITMAP *texture, int vc, V3D_f *vtx[]);
*)

type zbuffer

external create_zbuffer: bmp:bitmap -> zbuffer = "ml_create_zbuffer"

external create_sub_zbuffer: parent:zbuffer -> x:int -> y:int -> width:int -> height:int -> zbuffer
    = "ml_create_sub_zbuffer"

external set_zbuffer: zbuffer:zbuffer -> unit = "ml_set_zbuffer"

external clear_zbuffer: zbuffer:zbuffer -> z:float -> unit = "ml_clear_zbuffer"

external destroy_zbuffer: zbuffer:zbuffer -> unit = "ml_destroy_zbuffer"



(** {3:math3d 3D Math Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg033.html}
    Allegro API documentation for this module} *)

type matrix
type matrix_f

external get_identity_matrix: unit -> matrix = "ml_get_identity_matrix"
external get_identity_matrix_f: unit -> matrix_f = "ml_get_identity_matrix_f"
external free_matrix: matrix -> unit = "ml_free_matrix"

external make_matrix:
    v:(float * float * float) *
      (float * float * float) *
      (float * float * float) ->
    t:(float * float * float) -> matrix = "ml_make_matrix"

external make_matrix_f:
    v:(float * float * float) *
      (float * float * float) *
      (float * float * float) ->
    t:(float * float * float) -> matrix_f = "ml_make_matrix_f"
(**
  use [free_matrix] when not used any more.
  @param v 3x3 scaling and rotation component
  @param t x/y/z translation component
*)

external new_matrix: unit -> matrix = "ml_new_matrix"
external new_matrix_f: unit -> matrix_f = "ml_new_matrix_f"
(** [new_matrix] functions provide uninitialised matrices made to be set
    with the [get_*_matrix] functions. The choice have been made to not make
    the [get_*_matrix] functions return a fresh malloc'ed matrix, because 
    these are mainly to be used in the display loop, so alloc the matrices
    with [new_matrix] before the loop, manipulate and use these in the loop,
    and finaly free it with [free_matrix] at the end of the display loop. *)

external get_translation_matrix: m:matrix -> x:fixed -> y:fixed -> z:fixed -> unit = "ml_get_translation_matrix"
external get_translation_matrix_f: m:matrix -> x:float -> y:float -> z:float -> unit = "ml_get_translation_matrix_f"

external get_scaling_matrix: m:matrix -> x:fixed -> y:fixed -> z:fixed -> unit = "ml_get_scaling_matrix"
external get_scaling_matrix_f: m:matrix -> x:float -> y:float -> z:float -> unit = "ml_get_scaling_matrix_f"

external get_x_rotate_matrix: m:matrix -> r:fixed -> unit = "ml_get_x_rotate_matrix"
external get_x_rotate_matrix_f: m:matrix -> r:float -> unit = "ml_get_x_rotate_matrix_f"

external get_y_rotate_matrix: m:matrix -> r:fixed -> unit = "ml_get_y_rotate_matrix"
external get_y_rotate_matrix_f: m:matrix -> r:float -> unit = "ml_get_y_rotate_matrix_f"

external get_z_rotate_matrix: m:matrix -> r:fixed -> unit = "ml_get_z_rotate_matrix"
external get_z_rotate_matrix_f: m:matrix -> r:float -> unit = "ml_get_z_rotate_matrix_f"

external get_rotation_matrix: m:matrix -> x:fixed -> y:fixed -> z:fixed -> unit = "ml_get_rotation_matrix"
external get_rotation_matrix_f: m:matrix -> x:float -> y:float -> z:float -> unit = "ml_get_rotation_matrix_f"

external get_align_matrix: m:matrix -> xfront:fixed -> yfront:fixed -> zfront:fixed ->
    xup:fixed -> yup:fixed -> zup:fixed -> unit
    = "ml_get_align_matrix_bytecode"
      "ml_get_align_matrix_native"
external get_align_matrix_f: m:matrix_f -> xfront:float -> yfront:float -> zfront:float ->
    xup:float -> yup:float -> zup:float -> unit
    = "ml_get_align_matrix_f_bytecode"
      "ml_get_align_matrix_f_native"

external get_vector_rotation_matrix: m:matrix -> x:fixed -> y:fixed -> z:fixed -> a:fixed -> unit = "ml_get_vector_rotation_matrix"
external get_vector_rotation_matrix_f: m:matrix_f -> x:float -> y:float -> z:float -> a:float -> unit = "ml_get_vector_rotation_matrix_f"

external get_transformation_matrix: m:matrix -> scale:fixed ->
    xrot:fixed -> yrot:fixed -> zrot:fixed ->
    x:fixed -> y:fixed -> z:fixed -> unit
    = "ml_get_transformation_matrix_bytecode"
      "ml_get_transformation_matrix_native"
external get_transformation_matrix_f: m:matrix_f -> scale:float ->
    xrot:float -> yrot:float -> zrot:float ->
    x:float -> y:float -> z:float -> unit
    = "ml_get_transformation_matrix_f_bytecode"
      "ml_get_transformation_matrix_f_native"

external get_camera_matrix: m:matrix -> x:fixed -> y:fixed -> z:fixed ->
    xfront:fixed -> yfront:fixed -> zfront:fixed ->
    xup:fixed -> yup:fixed -> zup:fixed ->
    fov:fixed -> aspect:fixed -> unit
    = "ml_get_camera_matrix_bytecode"
      "ml_get_camera_matrix_native"
external get_camera_matrix_f: m:matrix_f -> x:float -> y:float -> z:float ->
    xfront:float -> yfront:float -> zfront:float ->
    xup:float -> yup:float -> zup:float ->
    fov:float -> aspect:float -> unit
    = "ml_get_camera_matrix_f_bytecode"
      "ml_get_camera_matrix_f_native"

external qtranslate_matrix: m:matrix -> x:fixed -> y:fixed -> z:fixed -> unit = "ml_qtranslate_matrix"
external qtranslate_matrix_f: m:matrix_f -> x:float -> y:float -> z:float -> unit = "ml_qtranslate_matrix_f"

external qscale_matrix: m:matrix -> scale:fixed -> unit = "ml_qscale_matrix"
external qscale_matrix_f: m:matrix_f -> scale:float -> unit = "ml_qscale_matrix_f"

external matrix_mul: m1:matrix -> m2:matrix -> out:matrix -> unit = "ml_matrix_mul"
external matrix_mul_f: m1:matrix_f -> m2:matrix_f -> out:matrix_f -> unit = "ml_matrix_mul_f"

external vector_length: x:fixed -> y:fixed -> z:fixed -> fixed = "ml_vector_length"
external vector_length_f: x:float -> y:float -> z:float -> float = "ml_vector_length_f"

(* TODO
void normalize_vector(fixed *x, fixed *y, fixed *z);
void normalize_vector_f(float *x, float *y, float *z);
fixed dot_product(fixed x1, y1, z1, x2, y2, z2);
float dot_product_f(float x1, y1, z1, x2, y2, z2);
void cross_product(fixed x1, y1, z1, x2, y2, z2, *xout, *yout, *zout);
void cross_product_f(float x1, y1, z1, x2, y2, z2, *xout, *yout, *zout);
fixed polygon_z_normal(const V3D *v1, const V3D *v2, const V3D *v3);
float polygon_z_normal_f(const V3D_f *v1, const V3D_f *v2, const V3D_f *v3);
*)
external apply_matrix: m:matrix -> x:fixed -> y:fixed -> z:fixed -> fixed * fixed * fixed = "ml_apply_matrix"
external apply_matrix_f: m:matrix_f -> x:float -> y:float -> z:float -> float * float * float = "ml_apply_matrix_f"
external set_projection_viewport: x:int -> y:int -> w:int -> h:int -> unit = "ml_set_projection_viewport"
external persp_project: x:fixed -> y:fixed -> z:fixed -> fixed * fixed = "ml_persp_project"
external persp_project_f: x:float -> y:float -> z:float -> float * float = "ml_persp_project_f"



(** {3:quater Quaternion Math Routines} *)

(** {{:http://alleg.sourceforge.net/stabledocs/en/alleg034.html}
    Allegro API documentation for this module} *)

type quat
external make_quat: w:float -> x:float -> y:float -> z:float -> quat = "ml_make_quat"
external free_quat: q:quat -> unit = "ml_free_quat"
external get_identity_quat: unit -> quat = "ml_get_identity_quat"
(** use [free_quat] at the end as with make_quat *)

external get_x_rotate_quat: q:quat -> r:float -> unit = "ml_get_x_rotate_quat"
external get_y_rotate_quat: q:quat -> r:float -> unit = "ml_get_y_rotate_quat"
external get_z_rotate_quat: q:quat -> r:float -> unit = "ml_get_z_rotate_quat"
(** the rotation is applied on the quat parameter (thus modiling it) *)

external get_rotation_quat: q:quat -> x:float -> y:float -> z:float -> unit = "ml_get_rotation_quat"
external get_vector_rotation_quat: q:quat -> x:float -> y:float -> z:float -> a:float -> unit = "ml_get_vector_rotation_quat"

external quat_to_matrix: q:quat -> matrix_f = "ml_quat_to_matrix"
external matrix_to_quat: m:matrix_f -> quat = "ml_matrix_to_quat"

external quat_mul: p:quat -> q:quat -> quat = "ml_quat_mul"
external apply_quat: q:quat -> x:float -> y:float -> z:float -> float * float * float = "ml_apply_quat"
external quat_interpolate: from:quat -> to_:quat -> t:float -> quat = "ml_quat_interpolate"

type how_slerp =
  | QUAT_SHORT (* like [quat_interpolate], use shortest path *)
  | QUAT_LONG (* rotation will be greater than 180 degrees *)
  | QUAT_CW (* rotate clockwise when viewed from above *)
  | QUAT_CCW (* rotate counterclockwise when viewed from above *)
  | QUAT_USER (* the quaternions are interpolated exactly as given *)

external quat_slerp: from:quat -> to_:quat -> t:float -> how:how_slerp -> quat = "ml_quat_slerp"


(* vim: sw=2 sts=2 ts=2 et fdm=marker filetype=ocaml
 *)
