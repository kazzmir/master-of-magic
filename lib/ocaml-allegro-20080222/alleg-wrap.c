/* {{{ COPYING *\

 +-------------------------------------------------------------------+
 | OCaml-Allegro, OCaml bindings for the Allegro library.            |
 | Copyright (C) 2007  Florent Monnier                               |
 +-------------------------------------------------------------------+
 |                                                                   |
 | This program is free software: you can redistribute it and/or     |
 | modify it under the terms of the GNU General Public License       |
 | as published by the Free Software Foundation, either version 3    |
 | of the License, or (at your option) any later version.            |
 |                                                                   |
 | This program is distributed in the hope that it will be useful,   |
 | but WITHOUT ANY WARRANTY; without even the implied warranty of    |
 | MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the     |
 | GNU General Public License for more details.                      |
 |                                                                   |
 | You should have received a copy of the GNU General Public License |
 | along with this program.  If not, see                             |
 | <http://www.gnu.org/licenses/>.                                   |
 +-------------------------------------------------------------------+
 | Author: Florent Monnier  <fmonnier@linux-nantes.org>              |
 +-------------------------------------------------------------------+

\* }}} */

#include <allegro.h>
#include <string.h>
#include <caml/mlvalues.h>
#include <caml/memory.h>
#include <caml/callback.h>
#include <caml/fail.h>
#include <caml/alloc.h>


CAMLprim value ml_allegro_init( value unit )
{
    if (allegro_init() != 0) caml_failwith("allegro_init");
    return Val_unit;
}


CAMLprim value ml_allegro_exit( value unit )
{
    allegro_exit();
    return Val_unit;
}


CAMLprim value ml_allegro_message( value msg )
{
    allegro_message( String_val(msg) );
    return Val_unit;
}


CAMLprim value ml_get_allegro_error( value unit )
{
    CAMLparam0();
    CAMLreturn( caml_copy_string(allegro_error) );
}


CAMLprim value ml_get_allegro_id( value unit )
{
    CAMLparam0();
    CAMLreturn( caml_copy_string(allegro_id) );
}


CAMLprim value ml_alleg_version( value unit )
{
    CAMLparam0();
    CAMLlocal1( version );

    const int year = ALLEGRO_DATE / 10000;
    const int month = (ALLEGRO_DATE / 100) % 100;
    const int day = ALLEGRO_DATE % 100;

    version = caml_alloc(7, 0);

    Store_field( version, 0, Val_int(ALLEGRO_VERSION) );
    Store_field( version, 1, Val_int(ALLEGRO_SUB_VERSION) );
    Store_field( version, 2, Val_int(ALLEGRO_WIP_VERSION) );

    Store_field( version, 3, caml_copy_string(ALLEGRO_VERSION_STR) );

    Store_field( version, 4, Val_int(year) );
    Store_field( version, 5, Val_int(month) );
    Store_field( version, 6, Val_int(day) );

    CAMLreturn( version );
}


CAMLprim value ml_set_window_title( value name )
{
    set_window_title( String_val(name) );
    return Val_unit;
}


CAMLprim value ml_cpu_vendor( value unit )
{
    CAMLparam0();
    check_cpu();
    CAMLreturn( caml_copy_string(cpu_vendor) );
}


CAMLprim value ml_get_cpu_family( value unit )
{
    check_cpu();
    switch (cpu_family)
    {
        case CPU_FAMILY_UNKNOWN:  return Val_int(0);
        case CPU_FAMILY_I386:     return Val_int(1);
        case CPU_FAMILY_I486:     return Val_int(2);
        case CPU_FAMILY_I586:     return Val_int(3);
        case CPU_FAMILY_I686:     return Val_int(4);
        case CPU_FAMILY_ITANIUM:  return Val_int(5);
        case CPU_FAMILY_POWERPC:  return Val_int(6);
        case CPU_FAMILY_EXTENDED: return Val_int(7);
        default: return Val_int(8);
    }
}

CAMLprim value ml_os_version( value unit ) { return Val_int(os_version); }
CAMLprim value ml_os_revision( value unit ) { return Val_int(os_revision); }

CAMLprim value ml_os_multitasking( value unit ) { if (os_multitasking) return Val_true; else return Val_false; }

CAMLprim value ml_desktop_color_depth( value unit ) { return Val_int(desktop_color_depth()); }


CAMLprim value ml_get_desktop_resolution( value unit )
{
    CAMLparam0();
    CAMLlocal1( resolution );
    int width, height;

    if (get_desktop_resolution( &width, &height ) != 0)
    {
        caml_failwith("get_desktop_resolution");
    }

    resolution = caml_alloc(2, 0);

    Store_field( resolution, 0, Val_int(width) );
    Store_field( resolution, 1, Val_int(height) );

    CAMLreturn( resolution );
}



CAMLprim value ml_set_color_depth( value depth )
{
    set_color_depth(Int_val(depth));
    return Val_unit;
}


CAMLprim value ml_get_color_depth( value unit )
{
    return Val_int( get_color_depth() );
}


CAMLprim value ml_get_screen_w( value unit ) { return Val_int(SCREEN_W); }
CAMLprim value ml_get_screen_h( value unit ) { return Val_int(SCREEN_H); }

CAMLprim value ml_get_virtual_w( value unit ) { return Val_int(VIRTUAL_W); }
CAMLprim value ml_get_virtual_h( value unit ) { return Val_int(VIRTUAL_H); }


CAMLprim value ml_request_refresh_rate( value rate )
{
    request_refresh_rate(Int_val(rate));
    return Val_unit;
}


CAMLprim value ml_get_refresh_rate( value unit )
{
    return Val_int( get_refresh_rate() );
}



CAMLprim value ml_set_gfx_mode( value driver, value w, value h, value v_w, value v_h )
{
    int _driver;
    switch (Int_val(driver))
    {
        case 0: _driver = GFX_AUTODETECT;            break;
        case 1: _driver = GFX_AUTODETECT_FULLSCREEN; break;
        case 2: _driver = GFX_AUTODETECT_WINDOWED;   break;
        case 3: _driver = GFX_SAFE;                  break;
        case 4: _driver = GFX_TEXT;                  break;
    }

    if (set_gfx_mode( _driver, Int_val(w), Int_val(h), Int_val(v_w), Int_val(v_h) ) != 0)
    {
        caml_failwith("set_gfx_mode");
    }
    return Val_unit;
}


CAMLprim value get_gfx_driver_name( value unit )
{
    CAMLparam0();
    CAMLreturn( caml_copy_string( gfx_driver->name ) );
}

CAMLprim value ml_enable_triple_buffer( value unit )
{
    if (enable_triple_buffer()) return Val_true; else return Val_false;
}


CAMLprim value ml_vsync( value unit )
{
    vsync();
    return Val_unit;
}


CAMLprim value ml_scroll_screen( value x, value y )
{
    if (scroll_screen( Int_val(x), Int_val(y) ) != 0)
    {
        caml_failwith("scroll_screen");
    }
    return Val_unit;
}


CAMLprim value ml_request_scroll( value x, value y )
{
    if (request_scroll( Int_val(x), Int_val(y) ) != 0)
    {
        caml_failwith("request_scroll");
    }
    return Val_unit;
}


CAMLprim value ml_poll_scroll( value unit )
{
    if (poll_scroll()) return Val_true; else return Val_false;
}


#define cons_cell( i, capability )     \
    if (gfx_capabilities & capability) \
    {                                  \
        cell = caml_alloc_small(2, 0); \
        Field(cell, 0) = Val_int(i);   \
        Field(cell, 1) = gfx_cap;      \
        gfx_cap = cell;                \
    }

CAMLprim value get_gfx_capabilities( value unit )
{
    CAMLparam0();
    CAMLlocal2( gfx_cap, cell );

    gfx_cap = Val_emptylist;

    cons_cell(  0, GFX_CAN_SCROLL );
    cons_cell(  1, GFX_CAN_TRIPLE_BUFFER);
    cons_cell(  2, GFX_HW_CURSOR );
    cons_cell(  3, GFX_SYSTEM_CURSOR );
    cons_cell(  4, GFX_HW_HLINE );
    cons_cell(  5, GFX_HW_HLINE_XOR );
    cons_cell(  6, GFX_HW_HLINE_SOLID_PATTERN );
    cons_cell(  7, GFX_HW_HLINE_COPY_PATTERN );
    cons_cell(  8, GFX_HW_FILL );
    cons_cell(  9, GFX_HW_FILL_XOR );
    cons_cell( 10, GFX_HW_FILL_SOLID_PATTERN );
    cons_cell( 11, GFX_HW_FILL_COPY_PATTERN );
    cons_cell( 12, GFX_HW_LINE );
    cons_cell( 13, GFX_HW_LINE_XOR );
    cons_cell( 14, GFX_HW_TRIANGLE );
    cons_cell( 15, GFX_HW_TRIANGLE_XOR );
    cons_cell( 16, GFX_HW_GLYPH );
    cons_cell( 17, GFX_HW_VRAM_BLIT );
    cons_cell( 18, GFX_HW_VRAM_BLIT_MASKED );
    cons_cell( 19, GFX_HW_MEM_BLIT );
    cons_cell( 20, GFX_HW_MEM_BLIT_MASKED );
    cons_cell( 21, GFX_HW_SYS_TO_VRAM_BLIT );
    cons_cell( 22, GFX_HW_SYS_TO_VRAM_BLIT_MASKED );
  //cons_cell( 23, GFX_HW_VRAM_STRETCH_BLIT );
  //cons_cell( 24, GFX_HW_SYS_STRETCH_BLIT );
  //cons_cell( 25, GFX_HW_VRAM_STRETCH_BLIT_MASKED );
  //cons_cell( 26, GFX_HW_SYS_STRETCH_BLIT_MASKED );

    CAMLreturn( gfx_cap );
}

//CAMLprim value ml_gfx_capabilities(value unit) { return Val_int(gfx_capabilities); }


CAMLprim value ml_set_display_switch_mode( value ml_mode )
{
    int mode;

    switch (Int_val(ml_mode))
    {
        case 0: mode = SWITCH_NONE; break;
        case 1: mode = SWITCH_PAUSE; break;
        case 2: mode = SWITCH_AMNESIA; break;
        case 3: mode = SWITCH_BACKGROUND; break;
        case 4: mode = SWITCH_BACKAMNESIA; break;
    }

    switch (set_display_switch_mode(mode))
    {
        case 0: return Val_unit;
        case -1: caml_failwith("set_display_switch_mode: requested mode not currently possible");
        default: caml_failwith("set_display_switch_mode");
    }
}

CAMLprim value ml_get_display_switch_mode( value unit )
{
    switch (get_display_switch_mode())
    {
        case SWITCH_NONE:        return Val_int(0);
        case SWITCH_PAUSE:       return Val_int(1);
        case SWITCH_AMNESIA:     return Val_int(2);
        case SWITCH_BACKGROUND:  return Val_int(3);
        case SWITCH_BACKAMNESIA: return Val_int(4);
        default: caml_failwith("get_display_switch_mode");
    }
}


CAMLprim value ml_is_windowed_mode( value unit )
{
    if (is_windowed_mode()) return Val_true; else return Val_false;
}


CAMLprim value ml_vram_single_surface( value unit )
{
#ifdef ALLEGRO_VRAM_SINGLE_SURFACE
    return Val_true;
#else
    return Val_false;
#endif
}

CAMLprim value ml_allegro_dos( value unit )
{
#ifdef ALLEGRO_DOS
    return Val_true;
#else
    return Val_false;
#endif
}


CAMLprim value ml_makecol(value r, value g, value b) { return Val_int( makecol( Int_val(r), Int_val(g), Int_val(b) )); }

CAMLprim value ml_makecol_depth(value depth, value r, value g, value b) { return Val_int( makecol_depth( Int_val(depth), Int_val(r), Int_val(g), Int_val(b) )); }

CAMLprim value ml_makecol8(value r, value g, value b) { return Val_int( makecol8( Int_val(r), Int_val(g), Int_val(b) )); }
CAMLprim value ml_makecol15(value r, value g, value b) { return Val_int( makecol15( Int_val(r), Int_val(g), Int_val(b) )); }
CAMLprim value ml_makecol16(value r, value g, value b) { return Val_int( makecol16( Int_val(r), Int_val(g), Int_val(b) )); }
CAMLprim value ml_makecol24(value r, value g, value b) { return Val_int( makecol24( Int_val(r), Int_val(g), Int_val(b) )); }
CAMLprim value ml_makecol32(value r, value g, value b) { return Val_int( makecol32( Int_val(r), Int_val(g), Int_val(b) )); }


CAMLprim value ml_makecol15_dither( value r, value g, value b, value x, value y )
{
    return Val_int(
        makecol15_dither( Int_val(r), Int_val(g), Int_val(b), Int_val(x), Int_val(y) ));
}

CAMLprim value ml_makecol16_dither( value r, value g, value b, value x, value y )
{
    return Val_int(
        makecol16_dither( Int_val(r), Int_val(g), Int_val(b), Int_val(x), Int_val(y) ));
}


CAMLprim value get_transparent( value unit ) { return Val_int( -1 ); }

CAMLprim value ml_makeacol(value r, value g, value b, value a) {
    return Val_int(
        makeacol( Int_val(r), Int_val(g), Int_val(b), Int_val(a) ));
}

CAMLprim value ml_makeacol32(value r, value g, value b, value a)
{
    return Val_int(
        makeacol32( Int_val(r), Int_val(g), Int_val(b), Int_val(a) ));
}

CAMLprim value ml_makeacol_depth(value depth, value r, value g, value b, value a)
{
    return Val_int(
        makeacol_depth( Int_val(depth), Int_val(r), Int_val(g), Int_val(b), Int_val(a) ));
}

CAMLprim value ml_getr( value color ) { return Val_int( getr( Int_val(color) )); }
CAMLprim value ml_getg( value color ) { return Val_int( getg( Int_val(color) )); }
CAMLprim value ml_getb( value color ) { return Val_int( getb( Int_val(color) )); }
CAMLprim value ml_geta( value color ) { return Val_int( geta( Int_val(color) )); }

CAMLprim value ml_getr8( value color ) { return Val_int( getr8( Int_val(color) )); }
CAMLprim value ml_getg8( value color ) { return Val_int( getg8( Int_val(color) )); }
CAMLprim value ml_getb8( value color ) { return Val_int( getb8( Int_val(color) )); }

CAMLprim value ml_getr15( value color ) { return Val_int( getr15( Int_val(color) )); }
CAMLprim value ml_getg15( value color ) { return Val_int( getg15( Int_val(color) )); }
CAMLprim value ml_getb15( value color ) { return Val_int( getb15( Int_val(color) )); }

CAMLprim value ml_getr16( value color ) { return Val_int( getr16( Int_val(color) )); }
CAMLprim value ml_getg16( value color ) { return Val_int( getg16( Int_val(color) )); }
CAMLprim value ml_getb16( value color ) { return Val_int( getb16( Int_val(color) )); }

CAMLprim value ml_getr24( value color ) { return Val_int( getr24( Int_val(color) )); }
CAMLprim value ml_getg24( value color ) { return Val_int( getg24( Int_val(color) )); }
CAMLprim value ml_getb24( value color ) { return Val_int( getb24( Int_val(color) )); }

CAMLprim value ml_getr32( value color ) { return Val_int( getr32( Int_val(color) )); }
CAMLprim value ml_getg32( value color ) { return Val_int( getg32( Int_val(color) )); }
CAMLprim value ml_getb32( value color ) { return Val_int( getb32( Int_val(color) )); }
CAMLprim value ml_geta32( value color ) { return Val_int( geta32( Int_val(color) )); }


CAMLprim value ml_getr_depth( value depth, value color ) { return Val_int( getr_depth( Int_val(depth), Int_val(color) )); }
CAMLprim value ml_getg_depth( value depth, value color ) { return Val_int( getg_depth( Int_val(depth), Int_val(color) )); }
CAMLprim value ml_getb_depth( value depth, value color ) { return Val_int( getb_depth( Int_val(depth), Int_val(color) )); }
CAMLprim value ml_geta_depth( value depth, value color ) { return Val_int( geta_depth( Int_val(depth), Int_val(color) )); }


CAMLprim value ml_palette_color( value i ) { return Val_int( palette_color[Int_val(i)] ); }


/* Palette Routines */

#define Val_palette(b) ((value) &b)
#define Palette_val(v) (*((PALETTE *) v))

CAMLprim value new_palette( value unit )
{
    PALETTE *palette;
    palette = malloc(sizeof(PALETTE));
    return ((value) palette);
}

CAMLprim value free_palette( value pal )
{
    free( Palette_val(pal) );
    return Val_unit;
}

CAMLprim value ml_set_palette( value pal )
{
    set_palette( Palette_val(pal) );
    return Val_unit;
}

CAMLprim value ml_get_palette( value pal )
{
    get_palette( Palette_val(pal) );
    return Val_unit;
}

CAMLprim value ml_set_color( value index, value r, value g, value b, value a )
{
    RGB rgb;
    rgb.r      = (unsigned char) Int_val(r);
    rgb.g      = (unsigned char) Int_val(g);
    rgb.b      = (unsigned char) Int_val(b);
    rgb.filler = (unsigned char) Int_val(a);

    set_color( Int_val(index), &rgb );
    return Val_unit;
}


CAMLprim value ml_palette_set_rgb( value _pal, value i, value r, value g, value b )
{
    PALETTE *pal;
    pal = (PALETTE *) _pal;
    (*pal)[Int_val(i)].r = Int_val(r);
    (*pal)[Int_val(i)].g = Int_val(g);
    (*pal)[Int_val(i)].b = Int_val(b);
    return Val_unit;
}

CAMLprim value ml_palette_set_rgba_native( value _pal, value i, value r, value g, value b, value a )
{
    PALETTE *pal;
    pal = (PALETTE *) _pal;
    (*pal)[Int_val(i)].r = Int_val(r);
    (*pal)[Int_val(i)].g = Int_val(g);
    (*pal)[Int_val(i)].b = Int_val(b);
    (*pal)[Int_val(i)].filler = Int_val(a);
    return Val_unit;
}
CAMLprim value ml_palette_set_rgba_bytecode(value * argv, int argn)
{
    return ml_palette_set_rgba_native(argv[0], argv[1], argv[2],
                                      argv[3], argv[4], argv[5]);
}

CAMLprim value ml_palette_get_rgb( value _pal, value i )
{
    CAMLparam2( _pal, i );
    CAMLlocal1( rgb );

    PALETTE *pal;
    pal = (PALETTE *) _pal;

    rgb = caml_alloc(3, 0);

    Store_field( rgb, 0, Val_int( (*pal)[Int_val(i)].r ) );
    Store_field( rgb, 1, Val_int( (*pal)[Int_val(i)].g ) );
    Store_field( rgb, 2, Val_int( (*pal)[Int_val(i)].b ) );

    CAMLreturn( rgb );
}

CAMLprim value ml_palette_get_rgba( value _pal, value i )
{
    CAMLparam2( _pal, i );
    CAMLlocal1( rgba );

    PALETTE *pal;
    pal = (PALETTE *) _pal;

    rgba = caml_alloc(4, 0);

    Store_field( rgba, 0, Val_int( (*pal)[Int_val(i)].r ) );
    Store_field( rgba, 1, Val_int( (*pal)[Int_val(i)].g ) );
    Store_field( rgba, 2, Val_int( (*pal)[Int_val(i)].b ) );
    Store_field( rgba, 3, Val_int( (*pal)[Int_val(i)].filler ) );

    CAMLreturn( rgba );
}

CAMLprim value ml_palette_copy_index( value src, value dst, value i )
{
    Palette_val(dst)[Int_val(i)] = Palette_val(src)[Int_val(i)];
    return Val_unit;
}


CAMLprim value ml_get_desktop_palette( value unit ) { return Val_palette( desktop_palette ); }

CAMLprim value ml_generate_332_palette( value pal ) { generate_332_palette( Palette_val(pal) ); return Val_unit; }

CAMLprim value ml_select_palette( value pal ) { select_palette( Palette_val(pal) ); return Val_unit; }

CAMLprim value ml_unselect_palette( value unit ) { unselect_palette(); return Val_unit; }


/*
TODO
// Converting between color formats
int bestfit_color(const PALETTE pal, int r, int g, int b);
*/

CAMLprim value ml_hsv_to_rgb( value h, value s, value v )
{
    CAMLparam3( h, s, v );
    CAMLlocal1( rgb );
    int r, g, b;

    hsv_to_rgb( Double_val(h), Double_val(s), Double_val(v), &r, &g, &b);

    rgb = caml_alloc(3, 0);

    Store_field( rgb, 0, Val_int(r) );
    Store_field( rgb, 1, Val_int(g) );
    Store_field( rgb, 2, Val_int(b) );

    CAMLreturn( rgb );
}


CAMLprim value ml_rgb_to_hsv( value r, value g, value b )
{
    CAMLparam3( r, g, b );
    CAMLlocal1( hsv );
    float h, s, v;

    rgb_to_hsv( Int_val(r), Int_val(g), Int_val(b), &h, &s, &v);

    hsv = caml_alloc(3, 0);

    Store_field( hsv, 0, caml_copy_double(h) );
    Store_field( hsv, 1, caml_copy_double(s) );
    Store_field( hsv, 2, caml_copy_double(v) );

    CAMLreturn( hsv );
}


#define Val_bitmap(b) ((value) b)
#define Bitmap_val(v) ((BITMAP *) v)

CAMLprim value ml_get_screen( value unit )
{
    if (screen != NULL)
        return Val_bitmap( screen );
    else
        caml_failwith("get_screen");
}


CAMLprim value ml_create_bitmap( value width, value height )
{
    BITMAP *b;
    b = create_bitmap( Int_val(width), Int_val(height) );
    if (!b) caml_failwith("create_bitmap");
    return Val_bitmap(b);
}

/*
CAMLprim value ml_check_bitmap( const value b )
{
    printf(" bitmap: %d\n", (int) b);
    return Val_unit;
}
*/

CAMLprim value ml_create_sub_bitmap( value parent, value x, value y, value width, value height )
{
    BITMAP *b;
    b = create_sub_bitmap( Bitmap_val(parent), Int_val(x), Int_val(y), Int_val(width), Int_val(height) );
    if (!b) caml_failwith("create_sub_bitmap");
    return Val_bitmap(b);
}


CAMLprim value ml_create_video_bitmap( value width, value height)
{
    BITMAP *b;
    b = create_video_bitmap( Int_val(width), Int_val(height) );
    if (!b) caml_failwith("create_video_bitmap");
    return Val_bitmap(b);
}


CAMLprim value ml_create_system_bitmap( value width, value height)
{
    BITMAP *b;
    b = create_system_bitmap( Int_val(width), Int_val(height) );
    if (!b) caml_failwith("create_system_bitmap");
    return Val_bitmap(b);
}


CAMLprim value ml_create_bitmap_ex( value color_depth, value width, value height )
{
    BITMAP *b;
    b = create_bitmap_ex( Int_val(color_depth), Int_val(width), Int_val(height) );
    if (!b) caml_failwith("create_bitmap_ex");
    return Val_bitmap(b);
}


CAMLprim value ml_bitmap_color_depth( value bmp )
{
    return Val_int( bitmap_color_depth(Bitmap_val(bmp)) );
}


CAMLprim value ml_destroy_bitmap( value bmp )
{
    destroy_bitmap( Bitmap_val(bmp) );
    return Val_unit;
}


CAMLprim value get_bitmap_width( value bmp ) { return Val_int( (Bitmap_val(bmp))->w ); }
CAMLprim value get_bitmap_height( value bmp ) { return Val_int( (Bitmap_val(bmp))->h ); }


CAMLprim value ml_acquire_bitmap( value bmp ) { acquire_bitmap( Bitmap_val(bmp) ); return Val_unit; }
CAMLprim value ml_release_bitmap( value bmp ) { release_bitmap( Bitmap_val(bmp) ); return Val_unit; }

CAMLprim value ml_acquire_screen( value unit ) { acquire_screen(); return Val_unit;}
CAMLprim value ml_release_screen( value unit ) { release_screen(); return Val_unit;}


CAMLprim value ml_bitmap_mask_color( value bmp )
{
    return Val_int( bitmap_mask_color(Bitmap_val(bmp)) );
}

CAMLprim value ml_is_same_bitmap( value bmp1, value bmp2 )
{
    if (is_same_bitmap(Bitmap_val(bmp1), Bitmap_val(bmp2))) return Val_true; else return Val_false;
}


CAMLprim value ml_show_video_bitmap( value bmp )
{
    if (show_video_bitmap(Bitmap_val(bmp)) != 0) caml_failwith("show_video_bitmap");
    return Val_unit;
}




CAMLprim value ml_request_video_bitmap( value bmp )
{
    if (request_video_bitmap(Bitmap_val(bmp)) != 0) caml_failwith("request_video_bitmap");
    return Val_unit;
}


CAMLprim value ml_clear_bitmap( value bmp )
{
    clear_bitmap( Bitmap_val(bmp) );
    return Val_unit;
}


CAMLprim value ml_clear_to_color( value bmp, value color )
{
    clear_to_color( Bitmap_val(bmp), Int_val(color) );
    return Val_unit;
}


CAMLprim value ml_add_clip_rect( value bmp, value x1, value y1, value x2, value y2 )
{
    add_clip_rect( Bitmap_val(bmp), Int_val(x1), Int_val(y1), Int_val(x2), Int_val(y2) );
    return Val_unit;
}


CAMLprim value ml_set_clip_rect( value bmp, value x1, value y1, value x2, value y2 )
{
    set_clip_rect( Bitmap_val(bmp), Int_val(x1), Int_val(y1), Int_val(x2), Int_val(y2) );
    return Val_unit;
}


CAMLprim value ml_set_clip_state( value bmp, value state )
{
    set_clip_state( Bitmap_val(bmp), Bool_val(state) );
    return Val_unit;
}


CAMLprim value ml_get_clip_state( value bmp )
{
    return Val_bool( get_clip_state(Bitmap_val(bmp)) );
}


CAMLprim value ml_is_planar_bitmap( value bmp )
{
    if (is_planar_bitmap(Bitmap_val(bmp))) return Val_true; else return Val_false;
}

CAMLprim value ml_is_linear_bitmap( value bmp )
{
    if (is_linear_bitmap(Bitmap_val(bmp))) return Val_true; else return Val_false;
}

CAMLprim value ml_is_memory_bitmap( value bmp )
{
    if (is_memory_bitmap(Bitmap_val(bmp))) return Val_true; else return Val_false;
}

CAMLprim value ml_is_screen_bitmap( value bmp )
{
    if (is_screen_bitmap(Bitmap_val(bmp))) return Val_true; else return Val_false;
}

CAMLprim value ml_is_video_bitmap( value bmp )
{
    if (is_video_bitmap(Bitmap_val(bmp))) return Val_true; else return Val_false;
}

CAMLprim value ml_is_system_bitmap( value bmp )
{
    if (is_system_bitmap(Bitmap_val(bmp))) return Val_true; else return Val_false;
}

CAMLprim value ml_is_sub_bitmap( value bmp )
{
    if (is_sub_bitmap(Bitmap_val(bmp))) return Val_true; else return Val_false;
}

CAMLprim value ml_save_bitmap(value filename, value img){
    BITMAP * b;
    b = Bitmap_val(img);
    if (save_bitmap(String_val(filename), b, NULL) != 0){
        return Val_true;
    } else {
        return Val_false;
    }
}

CAMLprim value ml_load_bitmap( value img, value pal )
{
    BITMAP *b;
    b = load_bitmap( String_val(img), Palette_val(pal) );
    if (!b) caml_failwith("load_bitmap");

    return Val_bitmap(b);
}


CAMLprim value ml_load_bmp( value img, value pal )
{
    BITMAP *b;
    b = load_bmp( String_val(img), Palette_val(pal) );
    if (!b) caml_failwith("load_bmp");

    return Val_bitmap(b);
}


CAMLprim value ml_load_lbm( value img, value pal )
{
    BITMAP *b;
    b = load_lbm( String_val(img), Palette_val(pal) );
    if (!b) caml_failwith("load_lbm");

    return Val_bitmap(b);
}


CAMLprim value ml_load_pcx( value img, value pal )
{
    BITMAP *b;
    b = load_pcx( String_val(img), Palette_val(pal) );
    if (!b) caml_failwith("load_pcx");

    return Val_bitmap(b);
}


CAMLprim value ml_load_tga( value img, value pal )
{
    BITMAP *b;
    b = load_tga( String_val(img), Palette_val(pal) );
    if (!b) caml_failwith("load_tga");

    return Val_bitmap(b);
}


CAMLprim value ml_blit_native( value src, value dest, value src_x, value src_y,
                               value dest_x, value dest_y, value width, value height )
{
    blit( Bitmap_val(src), Bitmap_val(dest), Int_val(src_x), Int_val(src_y),
          Int_val(dest_x), Int_val(dest_y), Int_val(width), Int_val(height) );
    return Val_unit;
}
CAMLprim value ml_blit_bytecode(value * argv, int argn)
{
    return ml_blit_native(argv[0], argv[1], argv[2], argv[3],
                          argv[4], argv[5], argv[6], argv[7]);
}


CAMLprim value ml_masked_blit_native( value src, value dest, value src_x, value src_y,
                                      value dest_x, value dest_y, value width, value height )
{
    masked_blit( Bitmap_val(src), Bitmap_val(dest), Int_val(src_x), Int_val(src_y),
                 Int_val(dest_x), Int_val(dest_y), Int_val(width), Int_val(height) );
    return Val_unit;
}
CAMLprim value ml_masked_blit_bytecode(value * argv, int argn)
{
    return ml_masked_blit_native(argv[0], argv[1], argv[2], argv[3],
                                 argv[4], argv[5], argv[6], argv[7]);
}


CAMLprim value ml_stretch_blit_native( value src, value dest,
                                       value src_x, value src_y, value src_width, value src_height,
                                       value dst_x, value dst_y, value dst_width, value dst_height )
{
    stretch_blit( Bitmap_val(src), Bitmap_val(dest),
                  Int_val(src_x), Int_val(src_y), Int_val(src_width), Int_val(src_height),
                  Int_val(dst_x), Int_val(dst_y), Int_val(dst_width), Int_val(dst_height) );
    return Val_unit;
}
CAMLprim value ml_stretch_blit_bytecode(value * argv, int argn)
{
    return ml_stretch_blit_native(argv[0], argv[1], argv[2], argv[3], argv[4],
                                  argv[5], argv[6], argv[7], argv[8], argv[9]);
}


CAMLprim value ml_masked_stretch_blit_native( value source, value dest,
                                              value src_x, value src_y, value src_w, value src_h,
                                              value dst_x, value dst_y, value dst_w, value dst_h )
{
    masked_stretch_blit( Bitmap_val(source), Bitmap_val(dest),
                         Int_val(src_x), Int_val(src_y), Int_val(src_w), Int_val(src_h),
                         Int_val(dst_x), Int_val(dst_y), Int_val(dst_w), Int_val(dst_h) );
    return Val_unit;
}
CAMLprim value ml_masked_stretch_blit_bytecode(value * argv, int argn)
{
    return ml_masked_stretch_blit_native(argv[0], argv[1], argv[2], argv[3], argv[4],
                                         argv[5], argv[6], argv[7], argv[8], argv[9]);
}


CAMLprim value ml_set_color_conversion( value ml_mode )
{
    int mode;

    switch (Int_val(ml_mode))
    {
        case  0: mode = COLORCONV_NONE             ; break;
        case  1: mode = COLORCONV_8_TO_15          ; break;
        case  2: mode = COLORCONV_8_TO_16          ; break;
        case  3: mode = COLORCONV_8_TO_24          ; break;
        case  4: mode = COLORCONV_8_TO_32          ; break;
        case  5: mode = COLORCONV_15_TO_8          ; break;
        case  6: mode = COLORCONV_15_TO_16         ; break;
        case  7: mode = COLORCONV_15_TO_24         ; break;
        case  8: mode = COLORCONV_15_TO_32         ; break;
        case  9: mode = COLORCONV_16_TO_8          ; break;
        case 10: mode = COLORCONV_16_TO_15         ; break;
        case 11: mode = COLORCONV_16_TO_24         ; break;
        case 12: mode = COLORCONV_16_TO_32         ; break;
        case 13: mode = COLORCONV_24_TO_8          ; break;
        case 14: mode = COLORCONV_24_TO_15         ; break;
        case 15: mode = COLORCONV_24_TO_16         ; break;
        case 16: mode = COLORCONV_24_TO_32         ; break;
        case 17: mode = COLORCONV_32_TO_8          ; break;
        case 18: mode = COLORCONV_32_TO_15         ; break;
        case 19: mode = COLORCONV_32_TO_16         ; break;
        case 20: mode = COLORCONV_32_TO_24         ; break;
        case 21: mode = COLORCONV_32A_TO_8         ; break;
        case 22: mode = COLORCONV_32A_TO_15        ; break;
        case 23: mode = COLORCONV_32A_TO_16        ; break;
        case 24: mode = COLORCONV_32A_TO_24        ; break;
        case 25: mode = COLORCONV_DITHER_PAL       ; break;
        case 26: mode = COLORCONV_DITHER_HI        ; break;
        case 27: mode = COLORCONV_KEEP_TRANS       ; break;

        case 28: mode = COLORCONV_EXPAND_256       ; break;
        case 29: mode = COLORCONV_REDUCE_TO_256    ; break;
        case 30: mode = COLORCONV_EXPAND_15_TO_16  ; break;
        case 31: mode = COLORCONV_REDUCE_16_TO_15  ; break;
        case 32: mode = COLORCONV_EXPAND_HI_TO_TRUE; break;
        case 33: mode = COLORCONV_REDUCE_TRUE_TO_HI; break;
        case 34: mode = COLORCONV_24_EQUALS_32     ; break;
        case 35: mode = COLORCONV_TOTAL            ; break;
        case 36: mode = COLORCONV_PARTIAL          ; break;
        case 37: mode = COLORCONV_MOST             ; break;
        case 38: mode = COLORCONV_DITHER           ; break;
      //case 39: mode = COLORCONV_KEEP_ALPHA       ; break;

        default: caml_failwith("set_color_conversion");
    }

    set_color_conversion( mode );
    return Val_unit;
}


CAMLprim value ml_get_color_conversion( value unit )
{
    CAMLparam0();
    CAMLlocal2( modes_li, cons );

    const int modes[] =
    {
        COLORCONV_NONE,
        COLORCONV_8_TO_15,
        COLORCONV_8_TO_16,
        COLORCONV_8_TO_24,
        COLORCONV_8_TO_32,
        COLORCONV_15_TO_8,
        COLORCONV_15_TO_16,
        COLORCONV_15_TO_24,
        COLORCONV_15_TO_32,
        COLORCONV_16_TO_8,
        COLORCONV_16_TO_15,
        COLORCONV_16_TO_24,
        COLORCONV_16_TO_32,
        COLORCONV_24_TO_8,
        COLORCONV_24_TO_15,
        COLORCONV_24_TO_16,
        COLORCONV_24_TO_32,
        COLORCONV_32_TO_8,
        COLORCONV_32_TO_15,
        COLORCONV_32_TO_16,
        COLORCONV_32_TO_24,
        COLORCONV_32A_TO_8,
        COLORCONV_32A_TO_15,
        COLORCONV_32A_TO_16,
        COLORCONV_32A_TO_24,
        COLORCONV_DITHER_PAL,
        COLORCONV_DITHER_HI,
        COLORCONV_KEEP_TRANS,
        COLORCONV_EXPAND_256,
        COLORCONV_REDUCE_TO_256,
        COLORCONV_EXPAND_15_TO_16,
        COLORCONV_REDUCE_16_TO_15,
        COLORCONV_EXPAND_HI_TO_TRUE,
        COLORCONV_REDUCE_TRUE_TO_HI,
        COLORCONV_24_EQUALS_32,
        COLORCONV_TOTAL,
        COLORCONV_PARTIAL,
        COLORCONV_MOST,
        COLORCONV_DITHER,
        0xFFFFFF,
    };
    int i = 0;

    int mode;
    mode = get_color_conversion();

    modes_li = Val_emptylist;

    while (modes[i] != 0xFFFFFF)
    {
        if (mode & modes[i])
        {
            cons = caml_alloc(2, 0);

            Store_field( cons, 0, Int_val(i) );
            Store_field( cons, 1, modes_li );

            modes_li = cons;
        }
    }

    CAMLreturn( modes_li );
}


CAMLprim value ml_install_keyboard( value unit )
{
    if (install_keyboard() != 0) caml_failwith("install_keyboard");
    return Val_unit;
}


CAMLprim value ml_poll_keyboard( value unit )
{
    if (poll_keyboard() != 0) caml_failwith("poll_keyboard");
    return Val_unit;
}


CAMLprim value ml_keypressed( value unit )
{
    return ( keypressed() ? Val_true : Val_false );
}


CAMLprim value ml_readkey( value unit )
{
    return Val_int( readkey() & 0xff );
}


CAMLprim value ml_readkey_scancode( value unit )
{
    switch (readkey() >> 8)
    {
        case KEY_A         : return Val_int(0);
        case KEY_B         : return Val_int(1);
        case KEY_C         : return Val_int(2);
        case KEY_D         : return Val_int(3);
        case KEY_E         : return Val_int(4);
        case KEY_F         : return Val_int(5);
        case KEY_G         : return Val_int(6);
        case KEY_H         : return Val_int(7);
        case KEY_I         : return Val_int(8);
        case KEY_J         : return Val_int(9);
        case KEY_K         : return Val_int(10);
        case KEY_L         : return Val_int(11);
        case KEY_M         : return Val_int(12);
        case KEY_N         : return Val_int(13);
        case KEY_O         : return Val_int(14);
        case KEY_P         : return Val_int(15);
        case KEY_Q         : return Val_int(16);
        case KEY_R         : return Val_int(17);
        case KEY_S         : return Val_int(18);
        case KEY_T         : return Val_int(19);
        case KEY_U         : return Val_int(20);
        case KEY_V         : return Val_int(21);
        case KEY_W         : return Val_int(22);
        case KEY_X         : return Val_int(23);
        case KEY_Y         : return Val_int(24);
        case KEY_Z         : return Val_int(25);
        case KEY_0         : return Val_int(26);
        case KEY_1         : return Val_int(27);
        case KEY_2         : return Val_int(28);
        case KEY_3         : return Val_int(29);
        case KEY_4         : return Val_int(30);
        case KEY_5         : return Val_int(31);
        case KEY_6         : return Val_int(32);
        case KEY_7         : return Val_int(33);
        case KEY_8         : return Val_int(34);
        case KEY_9         : return Val_int(35);
        case KEY_0_PAD     : return Val_int(36);
        case KEY_1_PAD     : return Val_int(37);
        case KEY_2_PAD     : return Val_int(38);
        case KEY_3_PAD     : return Val_int(39);
        case KEY_4_PAD     : return Val_int(40);
        case KEY_5_PAD     : return Val_int(41);
        case KEY_6_PAD     : return Val_int(42);
        case KEY_7_PAD     : return Val_int(43);
        case KEY_8_PAD     : return Val_int(44);
        case KEY_9_PAD     : return Val_int(45);
        case KEY_F1        : return Val_int(46);
        case KEY_F2        : return Val_int(47);
        case KEY_F3        : return Val_int(48);
        case KEY_F4        : return Val_int(49);
        case KEY_F5        : return Val_int(50);
        case KEY_F6        : return Val_int(51);
        case KEY_F7        : return Val_int(52);
        case KEY_F8        : return Val_int(53);
        case KEY_F9        : return Val_int(54);
        case KEY_F10       : return Val_int(55);
        case KEY_F11       : return Val_int(56);
        case KEY_F12       : return Val_int(57);
        case KEY_ESC       : return Val_int(58);
        case KEY_TILDE     : return Val_int(59);
        case KEY_MINUS     : return Val_int(60);
        case KEY_EQUALS    : return Val_int(61);
        case KEY_BACKSPACE : return Val_int(62);
        case KEY_TAB       : return Val_int(63);
        case KEY_OPENBRACE : return Val_int(64);
        case KEY_CLOSEBRACE: return Val_int(65);
        case KEY_ENTER     : return Val_int(66);
        case KEY_COLON     : return Val_int(67);
        case KEY_QUOTE     : return Val_int(68);
        case KEY_BACKSLASH : return Val_int(69);
        case KEY_BACKSLASH2: return Val_int(70);
        case KEY_COMMA     : return Val_int(71);
        case KEY_STOP      : return Val_int(72);
        case KEY_SLASH     : return Val_int(73);
        case KEY_SPACE     : return Val_int(74);
        case KEY_INSERT    : return Val_int(75);
        case KEY_DEL       : return Val_int(76);
        case KEY_HOME      : return Val_int(77);
        case KEY_END       : return Val_int(78);
        case KEY_PGUP      : return Val_int(79);
        case KEY_PGDN      : return Val_int(80);
        case KEY_LEFT      : return Val_int(81);
        case KEY_RIGHT     : return Val_int(82);
        case KEY_UP        : return Val_int(83);
        case KEY_DOWN      : return Val_int(84);
        case KEY_SLASH_PAD : return Val_int(85);
        case KEY_ASTERISK  : return Val_int(86);
        case KEY_MINUS_PAD : return Val_int(87);
        case KEY_PLUS_PAD  : return Val_int(88);
        case KEY_DEL_PAD   : return Val_int(89);
        case KEY_ENTER_PAD : return Val_int(90);
        case KEY_PRTSCR    : return Val_int(91);
        case KEY_PAUSE     : return Val_int(92);
        case KEY_ABNT_C1   : return Val_int(93);
        case KEY_YEN       : return Val_int(94);
        case KEY_KANA      : return Val_int(95);
        case KEY_CONVERT   : return Val_int(96);
        case KEY_NOCONVERT : return Val_int(97);
        case KEY_AT        : return Val_int(98);
        case KEY_CIRCUMFLEX: return Val_int(99);
        case KEY_COLON2    : return Val_int(100);
        case KEY_KANJI     : return Val_int(101);
        case KEY_LSHIFT    : return Val_int(102);
        case KEY_RSHIFT    : return Val_int(103);
        case KEY_LCONTROL  : return Val_int(104);
        case KEY_RCONTROL  : return Val_int(105);
        case KEY_ALT       : return Val_int(106);
        case KEY_ALTGR     : return Val_int(107);
        case KEY_LWIN      : return Val_int(108);
        case KEY_RWIN      : return Val_int(109);
        case KEY_MENU      : return Val_int(110);
        case KEY_SCRLOCK   : return Val_int(111);
        case KEY_NUMLOCK   : return Val_int(112);
        case KEY_CAPSLOCK  : return Val_int(113);
        case KEY_EQUALS_PAD: return Val_int(114);
        case KEY_BACKQUOTE : return Val_int(115);
        case KEY_SEMICOLON : return Val_int(116);
        case KEY_COMMAND   : return Val_int(117);

        default: caml_failwith("readkey_scancode");
    }
}


CAMLprim value ml_clear_keybuf( value unit ) { clear_keybuf(); return Val_unit; }


CAMLprim value ml_set_keyboard_rate( value delay, value repeat )
{
    set_keyboard_rate( Int_val(delay), Int_val(repeat) );
    return Val_unit;
}


CAMLprim value ml_remove_keyboard( value unit ) { remove_keyboard(); return Val_unit; }


CAMLprim value ml_key_key_esc( value unit ) { return ( key[KEY_ESC] ? Val_true : Val_false ); }
CAMLprim value ml_key_key_enter( value unit ) { return ( key[KEY_ENTER] ? Val_true : Val_false ); }

CAMLprim value ml_key_key_left( value unit ) { return ( key[KEY_LEFT] ? Val_true : Val_false ); }
CAMLprim value ml_key_key_right( value unit ) { return ( key[KEY_RIGHT] ? Val_true : Val_false ); }
CAMLprim value ml_key_key_up( value unit ) { return ( key[KEY_UP] ? Val_true : Val_false ); }
CAMLprim value ml_key_key_down( value unit ) { return ( key[KEY_DOWN] ? Val_true : Val_false ); }


CAMLprim value get_kb_flag( value ml_flag )
{
    int kb_flag;
    switch (Int_val(ml_flag))
    {
        case 0: kb_flag = KB_SHIFT_FLAG; break;
        case 1: kb_flag = KB_CTRL_FLAG; break;
        case 2: kb_flag = KB_ALT_FLAG; break;
        case 3: kb_flag = KB_LWIN_FLAG; break;
        case 4: kb_flag = KB_RWIN_FLAG; break;
        case 5: kb_flag = KB_MENU_FLAG; break;
        case 6: kb_flag = KB_COMMAND_FLAG; break;
        case 7: kb_flag = KB_SCROLOCK_FLAG; break;
        case 8: kb_flag = KB_NUMLOCK_FLAG; break;
        case 9: kb_flag = KB_CAPSLOCK_FLAG; break;
        case 10: kb_flag = KB_INALTSEQ_FLAG; break;
        case 11: kb_flag = KB_ACCENT1_FLAG; break;
        case 12: kb_flag = KB_ACCENT2_FLAG; break;
        case 13: kb_flag = KB_ACCENT3_FLAG; break;
        case 14: kb_flag = KB_ACCENT4_FLAG; break;
    }

    if (key_shifts & kb_flag)
        return Val_true;
    else
        return Val_false;
}



CAMLprim value ml_install_timer( value unit )
{
    if (install_timer() != 0) caml_failwith("install_timer");
    return Val_unit;
}

CAMLprim value get_retrace_count( value unit ) { return Val_int( retrace_count); }


CAMLprim value ml_rest( value time ) { rest( Int_val(time) ); return Val_unit; }



/*
static value caml_TimerFunc_cb = 0;  // real_call_back

CAMLprim void init_timerfunc( value f )
{
  caml_TimerFunc_cb = f;
  caml_register_global_root( &caml_TimerFunc_cb );
}
*/
void alleg_timer_cb(void *param_idx)
{
    /*
    leave_blocking_section();
    callback( caml_TimerFunc_cb, (value) param_idx );
    enter_blocking_section();
    */
    static value * closure_f = NULL;
    if (closure_f == NULL) {
        closure_f = caml_named_value("Alleg timer callback");
    }
    caml_callback(*closure_f, (value) param_idx );
}
END_OF_FUNCTION(alleg_timer_cb); 

CAMLprim value
ml_install_param_int( value ms, value param )
{
    if (install_param_int(alleg_timer_cb, (void *)param, Int_val(ms)) < 0)
        caml_failwith("install_param_int");

    LOCK_VARIABLE(caml_TimerFunc_cb);
    LOCK_FUNCTION(alleg_timer_cb);

    return Val_unit;
}


CAMLprim value
ml_remove_param_int( value param )
{
    remove_param_int(alleg_timer_cb, (void *)param);
    return Val_unit;
}



CAMLprim value ml_install_mouse( value unit )
{
    int ret;
    ret = install_mouse();
    if (ret == -1) caml_failwith("install_mouse");
    return Val_int(ret);
}


CAMLprim value ml_poll_mouse( value unit )
{
    if (poll_mouse() != 0) caml_failwith("poll_mouse");
    return Val_unit;
}


CAMLprim value get_mouse_driver_name( value unit )
{
    CAMLparam0();
    CAMLreturn( caml_copy_string( mouse_driver->name ) );
}


CAMLprim value ml_enable_hardware_cursor( value unit ) { enable_hardware_cursor(); return Val_unit; }
CAMLprim value ml_disable_hardware_cursor( value unit ) { disable_hardware_cursor(); return Val_unit; }

CAMLprim value ml_select_mouse_cursor( value _cursor )
{
    int cursor;
    switch (Int_val(_cursor))
    {
        case 0: cursor = MOUSE_CURSOR_NONE;     break;
        case 1: cursor = MOUSE_CURSOR_ALLEGRO;  break;
        case 2: cursor = MOUSE_CURSOR_ARROW;    break;
        case 3: cursor = MOUSE_CURSOR_BUSY;     break;
        case 4: cursor = MOUSE_CURSOR_QUESTION; break;
        case 5: cursor = MOUSE_CURSOR_EDIT;     break;
    }
    select_mouse_cursor( cursor );
    return Val_unit;
}


CAMLprim value ml_set_mouse_cursor_bitmap( value _cursor, value bmp )
{
    int cursor;
    switch (Int_val(_cursor))
    {
        case 0: caml_invalid_argument("set_mouse_cursor_bitmap");
        case 1: cursor = MOUSE_CURSOR_ALLEGRO;  break;
        case 2: cursor = MOUSE_CURSOR_ARROW;    break;
        case 3: cursor = MOUSE_CURSOR_BUSY;     break;
        case 4: cursor = MOUSE_CURSOR_QUESTION; break;
        case 5: cursor = MOUSE_CURSOR_EDIT;     break;
    }
    set_mouse_cursor_bitmap( cursor, Bitmap_val(bmp) );
    return Val_unit;
}

CAMLprim value ml_show_mouse( value bmp ) { show_mouse( Bitmap_val(bmp) ); return Val_unit; }
CAMLprim value ml_hide_mouse( value unit ) { show_mouse(NULL); return Val_unit; }
CAMLprim value ml_scare_mouse( value unit ) { scare_mouse(); return Val_unit; }

CAMLprim value ml_scare_mouse_area( value x, value y, value w, value h)
{
    scare_mouse_area( Int_val(x), Int_val(y), Int_val(w), Int_val(h) );
    return Val_unit;
}

CAMLprim value ml_unscare_mouse( value unit ) { unscare_mouse(); return Val_unit; }

CAMLprim value ml_position_mouse( value x, value y )
{
    position_mouse( Int_val(x), Int_val(y) );
    return Val_unit;
}

CAMLprim value ml_position_mouse_z( value z )
{
    position_mouse_z( Int_val(z) );
    return Val_unit;
}

#if 0
CAMLprim value ml_position_mouse_w( value w )
{
    position_mouse_w( Int_val(w) );
    return Val_unit;
}
#endif

CAMLprim value ml_set_mouse_range( value x1, value y1, value x2, value y2 )
{
    set_mouse_range( Int_val(x1), Int_val(y1), Int_val(x2), Int_val(y2) );
    return Val_unit;
}

CAMLprim value ml_set_mouse_speed( value xspeed, value yspeed )
{
    set_mouse_speed( Int_val(xspeed), Int_val(yspeed) );
    return Val_unit;
}


CAMLprim value ml_set_mouse_sprite( value sprite )
{
    set_mouse_sprite( Bitmap_val(sprite) );
    return Val_unit;
}

CAMLprim value ml_set_mouse_sprite_focus( value x, value y )
{
    set_mouse_sprite_focus( Int_val(x), Int_val(y) );
    return Val_unit;
}

CAMLprim value ml_get_mouse_mickeys( value unit )
{
    CAMLparam0();
    CAMLlocal1( mickeys );

    int mickeyx, mickeyy;
    get_mouse_mickeys( &mickeyx, &mickeyy );

    mickeys = caml_alloc(2, 0);

    Store_field( mickeys, 0, Val_int(mickeyx) );
    Store_field( mickeys, 1, Val_int(mickeyy) );

    CAMLreturn( mickeys );
}

CAMLprim value ml_mouse_x( value unit ) { return Val_int( mouse_x ); }
CAMLprim value ml_mouse_y( value unit ) { return Val_int( mouse_y ); }
CAMLprim value ml_mouse_z( value unit ) { return Val_int( mouse_z ); }
#if 0
CAMLprim value ml_mouse_w( value unit ) { return Val_int( mouse_w ); }
#endif

CAMLprim value ml_mouse_x_focus( value unit ) { return Val_int( mouse_x_focus ); }
CAMLprim value ml_mouse_y_focus( value unit ) { return Val_int( mouse_y_focus ); }

CAMLprim value left_button_pressed( value unit ) { if (mouse_b & 1) return Val_true; else return Val_false; }
CAMLprim value right_button_pressed( value unit ) { if (mouse_b & 2) return Val_true; else return Val_false; }
CAMLprim value middle_button_pressed( value unit ) { if (mouse_b & 4) return Val_true; else return Val_false; }

CAMLprim value get_mouse_b( value unit ) { return Val_int(mouse_b); }

CAMLprim value get_mouse_pos( value unit )
{
    CAMLparam0();
    CAMLlocal1( ml_pos );

    int pos, x, y;

    pos = mouse_pos;
    x = pos >> 16;
    y = pos & 0x0000ffff;

    ml_pos = caml_alloc(2, 0);

    Store_field( ml_pos, 0, Val_int(x) );
    Store_field( ml_pos, 1, Val_int(y) );

    CAMLreturn( ml_pos );
}


/* {{{ Fixed point math routines */

#define Val_fixed(f) ((value) f)
#define Fixed_val(v) ((fixed) v)


CAMLprim value ml_itofix( value i ) { return Val_fixed( itofix(Int_val(i)) ); }
CAMLprim value ml_fixtoi( value f ) { return Val_int( fixtoi( Fixed_val(f) ) ); }
CAMLprim value ml_ftofix( value d ) { return Val_fixed( ftofix( Double_val(d) ) ); }


CAMLprim value ml_fixtof( value f )
{
    CAMLparam1( f );
    CAMLreturn( caml_copy_double(
                fixtof( Fixed_val(f) ) ) );
}


CAMLprim value ml_fixadd( value x, value y ) { return Val_fixed( fixadd( Fixed_val(x), Fixed_val(y) ) ); }
CAMLprim value ml_fixsub( value x, value y ) { return Val_fixed( fixsub( Fixed_val(x), Fixed_val(y) ) ); }
CAMLprim value ml_fixdiv( value x, value y ) { return Val_fixed( fixdiv( Fixed_val(x), Fixed_val(y) ) ); }
CAMLprim value ml_fixmul( value x, value y ) { return Val_fixed( fixmul( Fixed_val(x), Fixed_val(y) ) ); }
CAMLprim value ml_fixhypot( value x, value y ) { return Val_fixed( fixhypot( Fixed_val(x), Fixed_val(y) ) ); }


CAMLprim value ml_fixceil( value x ) { return Val_int( fixceil( Fixed_val(x) ) ); }
CAMLprim value ml_fixfloor( value x ) { return Val_int( fixfloor( Fixed_val(x) ) ); }

CAMLprim value ml_fixsin( value x ) { return Val_fixed( fixsin( Fixed_val(x) ) ); }
CAMLprim value ml_fixcos( value x ) { return Val_fixed( fixcos( Fixed_val(x) ) ); }
CAMLprim value ml_fixtan( value x ) { return Val_fixed( fixtan( Fixed_val(x) ) ); }
CAMLprim value ml_fixasin( value x ) { return Val_fixed( fixasin( Fixed_val(x) ) ); }
CAMLprim value ml_fixacos( value x ) { return Val_fixed( fixacos( Fixed_val(x) ) ); }
CAMLprim value ml_fixatan( value x ) { return Val_fixed( fixatan( Fixed_val(x) ) ); }
CAMLprim value ml_fixsqrt( value x ) { return Val_fixed( fixsqrt( Fixed_val(x) ) ); }

CAMLprim value ml_fixtorad( value x ) { return Val_fixed( fixmul( Fixed_val(x), fixtorad_r ) ); }
CAMLprim value ml_fixofrad( value x ) { return Val_fixed( fixmul( Fixed_val(x), radtofix_r ) ); }

CAMLprim value ml_fixminus( value x ) { return Val_fixed( - Fixed_val(x) ); }

CAMLprim value ml_fixatan2( value y, value x )
{
    //if (y == 0) caml_invalid_argument("fixatan2: first argument is zero");  // XXX
    if (y == 0 && x == 0) caml_invalid_argument("fixatan2: both arguments are zero");
    return Val_fixed( fixatan2( Fixed_val(y), Fixed_val(x) ) );
}

/* }}} */


CAMLprim value ml_putpixel( value bmp, value x, value y, value color )
{
    putpixel( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(color) );
    return Val_unit;
}


CAMLprim value ml_rect_native( value bmp, value x1, value y1, value x2, value y2, value color )
{
    rect( Bitmap_val(bmp), Int_val(x1), Int_val(y1), Int_val(x2), Int_val(y2), Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_rect_bytecode(value * argv, int argn)
{
    return ml_rect_native(argv[0], argv[1], argv[2],
                          argv[3], argv[4], argv[5]);
}


CAMLprim value ml_rectfill_native( value bmp, value x1, value y1, value x2, value y2, value color )
{
    rectfill( Bitmap_val(bmp), Int_val(x1), Int_val(y1), Int_val(x2), Int_val(y2), Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_rectfill_bytecode(value * argv, int argn)
{
    return ml_rectfill_native(argv[0], argv[1], argv[2],
                              argv[3], argv[4], argv[5]);
}


CAMLprim value ml_arc_native( value bmp, value x, value y, value ang1, value ang2, value r, value color )
{
    arc( Bitmap_val(bmp), Int_val(x), Int_val(y), Fixed_val(ang1), Fixed_val(ang2), Int_val(r), Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_arc_bytecode(value * argv, int argn)
{
    return ml_arc_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5], argv[6]);
}


CAMLprim value ml_spline_native( value bmp,
        value x1, value y1, value x2, value y2, value x3, value y3, value x4, value y4,
        value color )
{
    int points[8];

    points[0] = Int_val(x1);
    points[1] = Int_val(y1);
    points[2] = Int_val(x2);
    points[3] = Int_val(y2);
    points[4] = Int_val(x3);
    points[5] = Int_val(y3);
    points[6] = Int_val(x4);
    points[7] = Int_val(y4);

    spline( Bitmap_val(bmp), points, Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_spline_bytecode(value * argv, int argn)
{
    return ml_spline_native(argv[0], argv[1], argv[2], argv[3], argv[4],
                            argv[5], argv[6], argv[7], argv[8], argv[9]);
}



CAMLprim value ml_floodfill( value bmp, value x, value y, value color )
{
    floodfill( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(color) );
    return Val_unit;
}


CAMLprim value ml_circle( value bmp, value x, value y, value radius, value color )
{
    /*
    circle( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(radius), Int_val(color) );
    */
    _soft_circle( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(radius), Int_val(color) );
    return Val_unit;
}


CAMLprim value ml_circlefill( value bmp, value x, value y, value radius, value color )
{
    circlefill( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(radius), Int_val(color) );
    return Val_unit;
}



CAMLprim value ml_ellipse_native( value bmp, value x, value y, value rx, value ry, value color )
{
    ellipse( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(rx), Int_val(ry), Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_ellipse_bytecode(value * argv, int argn)
{
    return ml_ellipse_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5]);
}



CAMLprim value ml_ellipsefill_native( value bmp, value x, value y, value rx, value ry, value color )
{
    ellipsefill( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(rx), Int_val(ry), Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_ellipsefill_bytecode(value * argv, int argn)
{
    return ml_ellipsefill_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5]);
}



CAMLprim value ml_triangle_native( value bmp, value x1, value y1, value x2, value y2, value x3, value y3, value color )
{
    triangle( Bitmap_val(bmp), Int_val(x1), Int_val(y1), Int_val(x2), Int_val(y2), Int_val(x3), Int_val(y3), Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_triangle_bytecode(value * argv, int argn)
{
    return ml_triangle_native(argv[0], argv[1], argv[2], argv[3],
                              argv[4], argv[5], argv[6], argv[7]);
}



CAMLprim value ml_line_native( value bmp, value x1, value y1, value x2, value y2, value color )
{
    line( Bitmap_val(bmp), Int_val(x1), Int_val(y1), Int_val(x2), Int_val(y2), Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_line_bytecode(value * argv, int argn)
{
    return ml_line_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5]);
}



CAMLprim value ml_fastline_native( value bmp, value x1, value y1, value x2, value y2, value color )
{
    fastline( Bitmap_val(bmp), Int_val(x1), Int_val(y1), Int_val(x2), Int_val(y2), Int_val(color) );
    return Val_unit;
}
CAMLprim value ml_fastline_bytecode(value * argv, int argn)
{
    return ml_fastline_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5]);
}


#if 0
CAMLprim value ml_vline( value bmp, value x, value y1, value y2, value color )
{
    vline( Bitmap_val(bmp), Int_val(x), Int_val(y1), Int_val(y2), Int_val(color) );
    return Val_unit;
}

CAMLprim value ml_hline( value bmp, value x1, value y, value x2, value color )
{
    hline( Bitmap_val(bmp), Int_val(x1), Int_val(y), Int_val(x2), Int_val(color) );
    return Val_unit;
}
#endif

/* XXX
 *  On my computer vline() and hline() have problem when used from ocaml
 *  so below there are two possible turn arround: use line() for both,
 *  or switch to the specific internal functions according to the color depth.
 */

#if 0
/* proto of the internal optimised functions to avoid a warning */
void *_linear_vline8(struct BITMAP *bmp, int x, int y1, int y2, int color);
void *_linear_vline15(struct BITMAP *bmp, int x, int y1, int y2, int color);
void *_linear_vline16(struct BITMAP *bmp, int x, int y1, int y2, int color);
void *_linear_vline24(struct BITMAP *bmp, int x, int y1, int y2, int color);
void *_linear_vline32(struct BITMAP *bmp, int x, int y1, int y2, int color);

void *_linear_hline8(struct BITMAP *bmp, int x1, int y, int x2, int color);
void *_linear_hline15(struct BITMAP *bmp, int x1, int y, int x2, int color);
void *_linear_hline16(struct BITMAP *bmp, int x1, int y, int x2, int color);
void *_linear_hline24(struct BITMAP *bmp, int x1, int y, int x2, int color);
void *_linear_hline32(struct BITMAP *bmp, int x1, int y, int x2, int color);
#endif

CAMLprim value ml_vline( value bmp, value x, value y1, value y2, value color )
{
    line( Bitmap_val(bmp), Int_val(x), Int_val(y1), Int_val(x), Int_val(y2), Int_val(color) );
    /*
    switch(bitmap_color_depth(Bitmap_val(bmp)))
    {
        case 8:  _linear_vline8(Bitmap_val(bmp), Int_val(x), Int_val(y1), Int_val(y2), Int_val(color)); break;
        case 15: _linear_vline15(Bitmap_val(bmp), Int_val(x), Int_val(y1), Int_val(y2), Int_val(color)); break;
        case 16: _linear_vline16(Bitmap_val(bmp), Int_val(x), Int_val(y1), Int_val(y2), Int_val(color)); break;
        case 24: _linear_vline24(Bitmap_val(bmp), Int_val(x), Int_val(y1), Int_val(y2), Int_val(color)); break;
        case 32: _linear_vline32(Bitmap_val(bmp), Int_val(x), Int_val(y1), Int_val(y2), Int_val(color)); break;
    }   
    */
    return Val_unit;
}   

CAMLprim value ml_hline( value bmp, value x1, value y, value x2, value color )
{
    line( Bitmap_val(bmp), Int_val(x1), Int_val(y), Int_val(x2), Int_val(y), Int_val(color) );
    /*
    switch(bitmap_color_depth(Bitmap_val(bmp)))
    {
        case 8:  _linear_hline8(Bitmap_val(bmp), Int_val(x1), Int_val(y), Int_val(x2), Int_val(color)); break;
        case 15: _linear_hline15(Bitmap_val(bmp), Int_val(x1), Int_val(y), Int_val(x2), Int_val(color)); break;
        case 16: _linear_hline16(Bitmap_val(bmp), Int_val(x1), Int_val(y), Int_val(x2), Int_val(color)); break;
        case 24: _linear_hline24(Bitmap_val(bmp), Int_val(x1), Int_val(y), Int_val(x2), Int_val(color)); break;
        case 32: _linear_hline32(Bitmap_val(bmp), Int_val(x1), Int_val(y), Int_val(x2), Int_val(color)); break;
    }   
    */
    return Val_unit;
}

CAMLprim value ml_getpixel( value bmp, value x, value y )
{
    return Val_int( getpixel( Bitmap_val(bmp), Int_val(x), Int_val(y) ) );
}



void do_circle_closure( BITMAP *bmp, int x, int y, int d )
{
    value args[4];

    static value * closure_f = NULL;
    if (closure_f == NULL) {
        closure_f = (value *) caml_named_value("Alleg callback do_circle");
    }

    args[0] = Val_bitmap(bmp);
    args[1] = Val_int(x);
    args[2] = Val_int(y);
    args[3] = Val_int(d);

    caml_callbackN( *closure_f, 4, args);
}

CAMLprim value ml_do_circle( value bmp, value x, value y, value radius, value d )
{
    do_circle( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(radius), Int_val(d),
               do_circle_closure);

    return Val_unit;
}


void do_ellipse_closure( BITMAP *bmp, int x, int y, int d )
{
    value args[4];

    static value * closure_f = NULL;
    if (closure_f == NULL) {
        closure_f = (value *) caml_named_value("Alleg callback do_ellipse");
    }

    args[0] = Val_bitmap(bmp);
    args[1] = Val_int(x);
    args[2] = Val_int(y);
    args[3] = Val_int(d);

    caml_callbackN( *closure_f, 4, args);
}

CAMLprim value ml_do_ellipse_native( value bmp, value x, value y, value rx, value ry, value d )
{
    do_ellipse( Bitmap_val(bmp), Int_val(x), Int_val(y), Int_val(rx), Int_val(ry), Int_val(d),
                do_ellipse_closure);

    return Val_unit;
}
CAMLprim value ml_do_ellipse_bytecode(value * argv, int argn)
{
    return ml_do_ellipse_native(argv[0], argv[1], argv[2],
                                argv[3], argv[4], argv[5]);
}



/* Sprites */

CAMLprim value ml_draw_sprite( value bmp, value sprite, value x, value y )
{
    draw_sprite( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y) );
    return Val_unit;
}


CAMLprim value ml_draw_sprite_v_flip( value bmp, value sprite, value x, value y )
{
    draw_sprite_v_flip( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y) );
    return Val_unit;
}


CAMLprim value ml_draw_sprite_h_flip( value bmp, value sprite, value x, value y )
{
    draw_sprite_h_flip( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y) );
    return Val_unit;
}


CAMLprim value ml_draw_sprite_vh_flip( value bmp, value sprite, value x, value y )
{
    draw_sprite_vh_flip( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y) );
    return Val_unit;
}


CAMLprim value ml_rotate_sprite( value bmp, value sprite, value x, value y, value angle )
{
    rotate_sprite(Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y), Fixed_val(angle) );
    return Val_unit;
}


CAMLprim value ml_rotate_sprite_v_flip( value bmp, value sprite, value x, value y, value angle )
{
    rotate_sprite_v_flip(Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y), Fixed_val(angle) );
    return Val_unit;
}


CAMLprim value ml_rotate_scaled_sprite_native( value bmp, value sprite, value x, value y,
                                               value angle, value scale )
{
    rotate_scaled_sprite( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y),
                          Fixed_val(angle), Fixed_val(scale) );
    return Val_unit;
}
CAMLprim value ml_rotate_scaled_sprite_bytecode(value * argv, int argn)
{
    return ml_rotate_scaled_sprite_native(argv[0], argv[1], argv[2],
                                          argv[3], argv[4], argv[5]);
}


CAMLprim value ml_rotate_scaled_sprite_v_flip_native( value bmp, value sprite, value x, value y,
                                                      value angle, value scale )
{
    rotate_scaled_sprite_v_flip( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y),
                                 Fixed_val(angle), Fixed_val(scale) );
    return Val_unit;
}
CAMLprim value ml_rotate_scaled_sprite_v_flip_bytecode(value * argv, int argn)
{
    return ml_rotate_scaled_sprite_v_flip_native(argv[0], argv[1], argv[2],
                                                 argv[3], argv[4], argv[5]);
}


CAMLprim value ml_pivot_sprite_native( value bmp, value sprite, value x, value y,
                                       value cx, value cy, value angle )
{
    pivot_sprite( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y),
                  Int_val(cx), Int_val(cy), Fixed_val(angle) );
    return Val_unit;
}
CAMLprim value ml_pivot_sprite_bytecode(value * argv, int argn)
{
    return ml_pivot_sprite_native(argv[0], argv[1], argv[2], argv[3],
                                  argv[4], argv[5], argv[6]);
}


CAMLprim value ml_pivot_sprite_v_flip_native( value bmp, value sprite, value x, value y,
                                              value cx, value cy, value angle )
{
    pivot_sprite_v_flip( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y),
                         Int_val(cx), Int_val(cy), Fixed_val(angle) );
    return Val_unit;
}
CAMLprim value ml_pivot_sprite_v_flip_bytecode(value * argv, int argn)
{
    return ml_pivot_sprite_v_flip_native(argv[0], argv[1], argv[2], argv[3],
                                         argv[4], argv[5], argv[6]);
}


CAMLprim value ml_pivot_scaled_sprite_native( value bmp, value sprite, value x, value y,
                                              value cx, value cy, value angle, value scale )
{
    pivot_scaled_sprite( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y),
                         Int_val(cx), Int_val(cy), Fixed_val(angle), Fixed_val(scale) );
    return Val_unit;
}
CAMLprim value ml_pivot_scaled_sprite_bytecode(value * argv, int argn)
{
    return ml_pivot_scaled_sprite_native(argv[0], argv[1], argv[2], argv[3],
                                         argv[4], argv[5], argv[6], argv[7]);
}


CAMLprim value ml_pivot_scaled_sprite_v_flip_native( value bmp, value sprite, value x, value y,
                                                     value cx, value cy, value angle, value scale )
{
    pivot_scaled_sprite_v_flip( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y),
                                Int_val(cx), Int_val(cy), Fixed_val(angle), Fixed_val(scale) );
    return Val_unit;
}
CAMLprim value ml_pivot_scaled_sprite_v_flip_bytecode(value * argv, int argn)
{
    return ml_pivot_scaled_sprite_v_flip_native(argv[0], argv[1], argv[2], argv[3],
                                                argv[4], argv[5], argv[6], argv[7]);
}


CAMLprim value ml_stretch_sprite_native(value bmp, value sprite, value x, value y, value w, value h)
{
    stretch_sprite( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y), Int_val(w), Int_val(h) );
    return Val_unit;
}
CAMLprim value ml_stretch_sprite_bytecode(value * argv, int argn)
{
    return ml_stretch_sprite_native(argv[0], argv[1], argv[2],
                                    argv[3], argv[4], argv[5]);
}


CAMLprim value ml_draw_character_ex_native( value bmp, value sprite, value x, value y, value color, value bg )
{
    draw_character_ex(Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y), Int_val(color), Int_val(bg) );
    return Val_unit;
}
CAMLprim value ml_draw_character_ex_bytecode(value * argv, int argn)
{
    return ml_draw_character_ex_native(argv[0], argv[1], argv[2],
                                       argv[3], argv[4], argv[5]);
}


CAMLprim value ml_draw_lit_sprite( value bmp, value sprite, value x, value y, value color )
{
    draw_lit_sprite( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y), Int_val(color) );
    return Val_unit;
}


CAMLprim value ml_draw_trans_sprite( value bmp, value sprite, value x, value y )
{
    draw_trans_sprite( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y) );
    return Val_unit;
}


CAMLprim value ml_draw_gouraud_sprite_native( value bmp, value sprite, value x, value y,
                                              value c1, value c2, value c3, value c4 )
{
    draw_gouraud_sprite( Bitmap_val(bmp), Bitmap_val(sprite), Int_val(x), Int_val(y),
                         Int_val(c1), Int_val(c2), Int_val(c3), Int_val(c4) );
    return Val_unit;
}
CAMLprim value ml_draw_gouraud_sprite_bytecode(value * argv, int argn)
{
    return ml_draw_gouraud_sprite_native(argv[0], argv[1], argv[2], argv[3],
                                         argv[4], argv[5], argv[6], argv[7]);
}


/* RLE Sprites */

#define Val_rle(f) ((value)f)
#define Rle_val(v) ((RLE_SPRITE *)v)

CAMLprim value ml_get_rle_sprite( value bmp )
{
    RLE_SPRITE *rle;
    rle = get_rle_sprite(Bitmap_val(bmp));
    return Val_rle(rle);
}

CAMLprim value ml_destroy_rle_sprite( value sprite )
{
    destroy_rle_sprite(Rle_val(sprite));
    return Val_unit;
}

CAMLprim value ml_draw_rle_sprite( value bmp, value sprite, value x, value y )
{
    draw_rle_sprite( Bitmap_val(bmp), Rle_val(sprite), Int_val(x), Int_val(y) );
    return Val_unit;
}

CAMLprim value ml_draw_trans_rle_sprite( value bmp, value sprite, value x, value y )
{
    draw_trans_rle_sprite( Bitmap_val(bmp), Rle_val(sprite), Int_val(x), Int_val(y) );
    return Val_unit;
}

CAMLprim value ml_draw_lit_rle_sprite( value bmp, value sprite, value x, value y, value color )
{
    draw_lit_rle_sprite( Bitmap_val(bmp), Rle_val(sprite), Int_val(x), Int_val(y), Int_val(color) );
    return Val_unit;
}


#define Val_comp(f) ((value)f)
#define Comp_val(v) ((COMPILED_SPRITE *)v)

CAMLprim value ml_get_compiled_sprite( value bmp, value planar )
{
    COMPILED_SPRITE *spr;
    spr = get_compiled_sprite( Bitmap_val(bmp), Bool_val(planar) );
    return Val_comp(spr);
}

CAMLprim value ml_destroy_compiled_sprite( value sprite )
{
    destroy_compiled_sprite( Comp_val(sprite) );
    return Val_unit;
}

CAMLprim value ml_draw_compiled_sprite( value bmp, value sprite, value x, value y )
{
    draw_compiled_sprite( Bitmap_val(bmp), Comp_val(sprite), Int_val(x), Int_val(y) );
    return Val_unit;
}


#define Val_font(f) ((value)f)
#define Font_val(v) ((FONT *)v)

CAMLprim value get_font( value unit ) { return Val_font(font); }

CAMLprim value ml_allegro_404_char( value c ) { allegro_404_char = Int_val(c); return Val_unit; }

CAMLprim value ml_text_length( value f, value str )
{
    return Val_int( text_length( Font_val(f), String_val(str) ) );
}

CAMLprim value ml_text_height( value f )
{
    return Val_int( text_height( Font_val(f) ) );
}


CAMLprim value ml_textout_ex_native( value bmp, value f, value str,
                                     value x, value y, value color, value bg )
{
    textout_ex( Bitmap_val(bmp), Font_val(f), String_val(str),
                Int_val(x), Int_val(y), Int_val(color), Int_val(bg) );
    return Val_unit;
}
CAMLprim value ml_textout_ex_bytecode(value * argv, int argn)
{
    return ml_textout_ex_native(argv[0], argv[1], argv[2], argv[3],
                                argv[4], argv[5], argv[6]);
}


CAMLprim value ml_textout_centre_ex_native( value bmp, value f, value str,
                                     value x, value y, value color, value bg )
{
    textout_centre_ex( Bitmap_val(bmp), Font_val(f), String_val(str),
                       Int_val(x), Int_val(y), Int_val(color), Int_val(bg) );
    return Val_unit;
}
CAMLprim value ml_textout_centre_ex_bytecode(value * argv, int argn)
{
    return ml_textout_centre_ex_native(argv[0], argv[1], argv[2], argv[3],
                                       argv[4], argv[5], argv[6]);
}


CAMLprim value ml_textout_right_ex_native( value bmp, value f, value str,
                                     value x, value y, value color, value bg )
{
    textout_right_ex( Bitmap_val(bmp), Font_val(f), String_val(str),
                      Int_val(x), Int_val(y), Int_val(color), Int_val(bg) );
    return Val_unit;
}
CAMLprim value ml_textout_right_ex_bytecode(value * argv, int argn)
{
    return ml_textout_right_ex_native(argv[0], argv[1], argv[2], argv[3],
                                       argv[4], argv[5], argv[6]);
}


CAMLprim value ml_textout_justify_ex_native( value bmp, value f, value str,
                                             value x1, value x2, value y,
                                             value diff, value color, value bg )
{
    textout_justify_ex( Bitmap_val(bmp), Font_val(f), String_val(str),
                        Int_val(x1), Int_val(x2), Int_val(y),
                        Int_val(diff), Int_val(color), Int_val(bg) );
    return Val_unit;
}
CAMLprim value ml_textout_justify_ex_bytecode(value * argv, int argn)
{
    return ml_textout_justify_ex_native(argv[0], argv[1], argv[2], argv[3],
                                        argv[4], argv[5], argv[6], argv[7], argv[8]);
}


CAMLprim value ml_load_font( value filename )
{
    CAMLparam1( filename );
    CAMLlocal1( ml_font );

    FONT *myfont;
    PALETTE *palette;
    palette = malloc(sizeof(PALETTE));

    myfont = load_font( String_val(filename), (RGB *) &palette, NULL );


    ml_font = caml_alloc(2, 0);

    Store_field( ml_font, 0, Val_font(myfont) );
    Store_field( ml_font, 1, ((value) palette) );

    CAMLreturn( ml_font );
}


CAMLprim value ml_extract_font_range( value f, value begin, value end )
{
    FONT *font_range;

    font_range = extract_font_range( Font_val(f), Int_val(begin), Int_val(end) );

    if (font_range == NULL) caml_failwith("extract_font_range");

    return Val_font(font_range);
}


CAMLprim value ml_merge_fonts( value f1, value f2 )
{
    FONT *font_merge;

    font_merge = merge_fonts( Font_val(f1), Font_val(f2) );

    return Val_font(font_merge);
}


CAMLprim value ml_destroy_font( value f )
{
    destroy_font( Font_val(f) );
    return Val_unit;
}

/* {{{ Fonts TODO */

/*
font_has_alpha
get_font_range_begin
get_font_range_end
get_font_ranges
grab_font_from_bitmap
is_color_font
is_compatible_font
is_mono_font
is_trans_font
load_bios_font
load_bitmap_font
load_dat_font
load_grx_font
load_grx_or_bios_font
load_txt_font
make_trans_font
register_font_file_type
transpose_font
*/

/* }}} */


CAMLprim value ml_drawing_mode_pattern( value ml_mode, value pattern, value x_anchor, value y_anchor)
{
    int mode;
    switch (Int_val(ml_mode))
    {
        case 0: mode = DRAW_MODE_SOLID; break;
        case 1: mode = DRAW_MODE_XOR; break;
        case 2: mode = DRAW_MODE_COPY_PATTERN; break;
        case 3: mode = DRAW_MODE_SOLID_PATTERN; break;
        case 4: mode = DRAW_MODE_MASKED_PATTERN; break;
        case 5: mode = DRAW_MODE_TRANS; break;
    }
    drawing_mode( mode, Bitmap_val(pattern), Int_val(x_anchor), Int_val(y_anchor) );
    return Val_unit;
}


CAMLprim value ml_drawing_mode_1( value ml_mode )
{
    int mode;
    switch (Int_val(ml_mode))
    {
        case 0: mode = DRAW_MODE_SOLID; break;
        case 1: mode = DRAW_MODE_XOR; break;
        case 2: mode = DRAW_MODE_COPY_PATTERN; break;
        case 3: mode = DRAW_MODE_SOLID_PATTERN; break;
        case 4: mode = DRAW_MODE_MASKED_PATTERN; break;
        case 5: mode = DRAW_MODE_TRANS; break;
    }
    drawing_mode( mode, NULL, 0, 0 );
    return Val_unit;
}


CAMLprim value ml_xor_mode( value on )
{
    xor_mode( Bool_val(on) );
    return Val_unit;
}

CAMLprim value ml_solid_mode( value unit ) { solid_mode(); return Val_unit; }

CAMLprim value ml_set_trans_blender( value r, value g, value b, value a )
{
    set_trans_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_alpha_blender( value unit )
{
    set_alpha_blender();
    return Val_unit;
}

CAMLprim value ml_set_write_alpha_blender( value unit )
{
    set_write_alpha_blender();
    return Val_unit;
}

CAMLprim value ml_set_add_blender( value r, value g, value b, value a )
{
    set_add_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_burn_blender( value r, value g, value b, value a )
{
    set_burn_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_color_blender( value r, value g, value b, value a )
{
    set_color_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_difference_blender( value r, value g, value b, value a )
{
    set_difference_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_dissolve_blender( value r, value g, value b, value a )
{
    set_dissolve_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_dodge_blender( value r, value g, value b, value a )
{
    set_dodge_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_hue_blender( value r, value g, value b, value a )
{
    set_hue_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_invert_blender( value r, value g, value b, value a )
{
    set_invert_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_luminance_blender( value r, value g, value b, value a )
{
    set_luminance_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_multiply_blender( value r, value g, value b, value a )
{
    set_multiply_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_saturation_blender( value r, value g, value b, value a )
{
    set_saturation_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}

CAMLprim value ml_set_screen_blender( value r, value g, value b, value a )
{
    set_screen_blender(Int_val(r), Int_val(g), Int_val(b), Int_val(a));
    return Val_unit;
}



CAMLprim value ml_digi_driver_name( value unit )
{
    CAMLparam0();
    CAMLreturn( caml_copy_string( digi_driver->name ) );
}


CAMLprim value ml_install_sound( value _digi, value _midi )
{
    int digi, midi;
    switch (Int_val(_digi))
    {
        case 0: digi = DIGI_AUTODETECT; break;
        case 1: digi = DIGI_NONE; break;
    }
    switch (Int_val(_midi))
    {
        case 0: midi = MIDI_AUTODETECT; break;
        case 1: midi = MIDI_NONE; break;
    }

    if (install_sound( digi, midi, NULL) != 0) caml_failwith("install_sound");
    return Val_unit;
}

CAMLprim value ml_remove_sound( value unit )
{
    remove_sound();
    return Val_unit;
}


CAMLprim value ml_reserve_voices( value digi_voices, value midi_voices)
{
    reserve_voices( Int_val(digi_voices), Int_val(midi_voices) );
    return Val_unit;
}

CAMLprim value ml_set_volume_per_voice( value scale )
{
    set_volume_per_voice( Int_val(scale) );
    return Val_unit;
}

CAMLprim value ml_set_volume( value digi_volume, value midi_volume )
{
    set_volume( Int_val(digi_volume), Int_val(midi_volume) );
    return Val_unit;
}

CAMLprim value ml_set_hardware_volume( value digi_volume, value midi_volume )
{
    set_hardware_volume( Int_val(digi_volume), Int_val(midi_volume) );
    return Val_unit;
}

#if 0
CAMLprim value ml_get_volume( value unit )
{
    CAMLparam0();
    CAMLlocal1( volume );

    int digi_volume, midi_volume;

    get_volume( &digi_volume, &midi_volume );

    volume = caml_alloc(2, 0);

    Store_field( volume, 0, Val_int(digi_volume) );
    Store_field( volume, 1, Val_int(midi_volume) );

    CAMLreturn( volume );
}

CAMLprim value ml_get_hardware_volume( value unit )
{
    CAMLparam0();
    CAMLlocal1( volume );

    int digi_volume, midi_volume;

    get_hardware_volume( &digi_volume, &midi_volume );

    volume = caml_alloc(2, 0);

    Store_field( volume, 0, Val_int(digi_volume) );
    Store_field( volume, 1, Val_int(midi_volume) );

    CAMLreturn( volume );
}
#endif

/*
int detect_digi_driver(int driver_id);
int detect_midi_driver(int driver_id);
*/


/* {{{ Digital Sample Routines */

#define Sample_val(spl) ((SAMPLE *) spl)
#define Val_sample(spl) ((value) spl)

CAMLprim value ml_load_sample( value filename )
{
    SAMPLE *spl;
    spl = load_sample( String_val(filename) );
    return Val_sample(spl);
}


CAMLprim value ml_destroy_sample( value spl )
{
    destroy_sample( Sample_val(spl) );
    return Val_unit;
}


CAMLprim value ml_adjust_sample( value spl, value vol, value pan, value freq, value loop )
{
    adjust_sample( Sample_val(spl), Int_val(vol), Int_val(pan), Int_val(freq), Bool_val(loop) );
    return Val_unit;
}


CAMLprim value ml_play_sample( value spl, value vol, value pan, value freq, value loop )
{
    return Val_int( play_sample( Sample_val(spl), Int_val(vol), Int_val(pan), Int_val(freq), Bool_val(loop) ) );
}


CAMLprim value ml_stop_sample( value spl )
{
    stop_sample( Sample_val(spl) );
    return Val_unit;
}

/* }}} */

CAMLprim value ml_replace_filename( value path, value filename )
{
    CAMLparam0();
    char dest[256];
    replace_filename(dest, String_val(path), String_val(filename), sizeof(dest));
    CAMLreturn( caml_copy_string(dest) );
}

/* {{{ Datafile routines */

#define Val_datafile(d) ((value)d)
#define Datafile_val(v) ((DATAFILE *)v)

CAMLprim value ml_load_datafile( value filename )
{
    DATAFILE *dat;
    dat = load_datafile( String_val(filename) );
    if (!dat) caml_failwith("load_datafile");
    return Val_datafile(dat);
}

CAMLprim value ml_unload_datafile( value dat )
{
    unload_datafile( Datafile_val(dat) );
    return Val_unit;
}

CAMLprim value datafile_index( value dat, value i )
{
    return (value) (Datafile_val(dat))[Long_val(i)].dat;
}

CAMLprim value ml_fixup_datafile( value dat )
{
    fixup_datafile( Datafile_val(dat) );
    return Val_unit;
}

/*
TODO:
const char *get_datafile_property(const DATAFILE *dat, int type);
DATAFILE_INDEX *create_datafile_index(const char *filename);
DATAFILE *load_datafile_object_indexed(const DATAFILE_INDEX *index, int item);
DATAFILE *load_datafile_object(const char *filename, const char *objectname);
void unload_datafile_object(DATAFILE *dat);
*/

/* }}} */



/* GUI Routines */

CAMLprim value ml_gfx_mode_select_ex( value unit )
{
    CAMLparam1( unit );
    CAMLlocal1( out );

    int w, h, bpp;
    int gfx_mode = GFX_AUTODETECT;
    w = SCREEN_W;
    h = SCREEN_H;
    bpp = bitmap_color_depth(screen);

    if (!gfx_mode_select_ex( &gfx_mode, &w, &h, &bpp ))
    {
        caml_failwith("gfx_mode_select_ex");
    }

    out = caml_alloc(4, 0);

    switch (gfx_mode)
    {
        case GFX_AUTODETECT:             Store_field( out, 0, Val_int(0) ); break;
        case GFX_AUTODETECT_FULLSCREEN:  Store_field( out, 0, Val_int(1) ); break;
        case GFX_AUTODETECT_WINDOWED:    Store_field( out, 0, Val_int(2) ); break;
        case GFX_SAFE:                   Store_field( out, 0, Val_int(3) ); break;
        case GFX_TEXT:                   Store_field( out, 0, Val_int(4) ); break;
        default:  caml_failwith("gfx_mode_select_ex");
    }

    Store_field( out, 1, Val_int(w) );
    Store_field( out, 2, Val_int(h) );
    Store_field( out, 3, Val_int(bpp) );

    CAMLreturn( out );
}



/* Polygon rendering */

/* {{{ triangle3d_f */

CAMLprim value ml_triangle3d_f_native( value bmp, value _type, value tex, value _v1, value _v2, value _v3 )
{
    V3D_f v1, v2, v3;

    v1.x = Double_val(Field(_v1, 0));
    v1.y = Double_val(Field(_v1, 1));
    v1.z = Double_val(Field(_v1, 2));
    v1.u = Double_val(Field(_v1, 3));
    v1.v = Double_val(Field(_v1, 4));
    v1.c = Int_val(Field(_v1, 5));

    v2.x = Double_val(Field(_v2, 0));
    v2.y = Double_val(Field(_v2, 1));
    v2.z = Double_val(Field(_v2, 2));
    v2.u = Double_val(Field(_v2, 3));
    v2.v = Double_val(Field(_v2, 4));
    v2.c = Int_val(Field(_v2, 5));

    v3.x = Double_val(Field(_v3, 0));
    v3.y = Double_val(Field(_v3, 1));
    v3.z = Double_val(Field(_v3, 2));
    v3.u = Double_val(Field(_v3, 3));
    v3.v = Double_val(Field(_v3, 4));
    v3.c = Int_val(Field(_v3, 5));

    int type;
    switch (Int_val(_type))
    {
        case  0: type = POLYTYPE_ATEX; break;
        case  1: type = POLYTYPE_ATEX_LIT; break;
        case  2: type = POLYTYPE_ATEX_MASK; break;
        case  3: type = POLYTYPE_ATEX_MASK_LIT; break;
        case  4: type = POLYTYPE_ATEX_MASK_TRANS; break;
        case  5: type = POLYTYPE_ATEX_TRANS; break;
        case  6: type = POLYTYPE_FLAT; break;
        case  7: type = POLYTYPE_GCOL; break;
        case  8: type = POLYTYPE_GRGB; break;
        case  9: type = POLYTYPE_PTEX; break;
        case 10: type = POLYTYPE_PTEX_LIT; break;
        case 11: type = POLYTYPE_PTEX_MASK; break;
        case 12: type = POLYTYPE_PTEX_MASK_LIT; break;
        case 13: type = POLYTYPE_PTEX_MASK_TRANS; break;
        case 14: type = POLYTYPE_PTEX_TRANS; break;
        default: caml_invalid_argument("triangle3d_f");
    }
    triangle3d_f( Bitmap_val(bmp), type, Bitmap_val(tex), &v1, &v2, &v3 );
    return Val_unit;
}
CAMLprim value ml_triangle3d_f_bytecode(value * argv, int argn)
{
    return ml_triangle3d_f_native(argv[0], argv[1], argv[2],
                                  argv[3], argv[4], argv[5]);
}

/* }}} */
/* {{{ triangle3d */

CAMLprim value ml_triangle3d_native( value bmp, value _type, value tex, value _v1, value _v2, value _v3 )
{
    V3D v1, v2, v3;

    v1.x = Fixed_val(Field(_v1, 0));
    v1.y = Fixed_val(Field(_v1, 1));
    v1.z = Fixed_val(Field(_v1, 2));
    v1.u = Fixed_val(Field(_v1, 3));
    v1.v = Fixed_val(Field(_v1, 4));
    v1.c = Int_val(Field(_v1, 5));

    v2.x = Fixed_val(Field(_v2, 0));
    v2.y = Fixed_val(Field(_v2, 1));
    v2.z = Fixed_val(Field(_v2, 2));
    v2.u = Fixed_val(Field(_v2, 3));
    v2.v = Fixed_val(Field(_v2, 4));
    v2.c = Int_val(Field(_v2, 5));

    v3.x = Fixed_val(Field(_v3, 0));
    v3.y = Fixed_val(Field(_v3, 1));
    v3.z = Fixed_val(Field(_v3, 2));
    v3.u = Fixed_val(Field(_v3, 3));
    v3.v = Fixed_val(Field(_v3, 4));
    v3.c = Int_val(Field(_v3, 5));

    int type;
    switch (Int_val(_type))
    {
        case  0: type = POLYTYPE_ATEX; break;
        case  1: type = POLYTYPE_ATEX_LIT; break;
        case  2: type = POLYTYPE_ATEX_MASK; break;
        case  3: type = POLYTYPE_ATEX_MASK_LIT; break;
        case  4: type = POLYTYPE_ATEX_MASK_TRANS; break;
        case  5: type = POLYTYPE_ATEX_TRANS; break;
        case  6: type = POLYTYPE_FLAT; break;
        case  7: type = POLYTYPE_GCOL; break;
        case  8: type = POLYTYPE_GRGB; break;
        case  9: type = POLYTYPE_PTEX; break;
        case 10: type = POLYTYPE_PTEX_LIT; break;
        case 11: type = POLYTYPE_PTEX_MASK; break;
        case 12: type = POLYTYPE_PTEX_MASK_LIT; break;
        case 13: type = POLYTYPE_PTEX_MASK_TRANS; break;
        case 14: type = POLYTYPE_PTEX_TRANS; break;
        default: caml_invalid_argument("triangle3d");
    }
    triangle3d( Bitmap_val(bmp), type, Bitmap_val(tex), &v1, &v2, &v3 );
    return Val_unit;
}
CAMLprim value ml_triangle3d_bytecode(value * argv, int argn)
{
    return ml_triangle3d_native(argv[0], argv[1], argv[2],
                                argv[3], argv[4], argv[5]);
}

/* }}} */

/* {{{ quad3d_f */

CAMLprim value ml_quad3d_f_native( value bmp, value _type, value tex,
                                   value _v1, value _v2, value _v3, value _v4 )
{
    V3D_f v1, v2, v3, v4;

    v1.x = Double_val(Field(_v1, 0));
    v1.y = Double_val(Field(_v1, 1));
    v1.z = Double_val(Field(_v1, 2));
    v1.u = Double_val(Field(_v1, 3));
    v1.v = Double_val(Field(_v1, 4));
    v1.c = Int_val(Field(_v1, 5));

    v2.x = Double_val(Field(_v2, 0));
    v2.y = Double_val(Field(_v2, 1));
    v2.z = Double_val(Field(_v2, 2));
    v2.u = Double_val(Field(_v2, 3));
    v2.v = Double_val(Field(_v2, 4));
    v2.c = Int_val(Field(_v2, 5));

    v3.x = Double_val(Field(_v3, 0));
    v3.y = Double_val(Field(_v3, 1));
    v3.z = Double_val(Field(_v3, 2));
    v3.u = Double_val(Field(_v3, 3));
    v3.v = Double_val(Field(_v3, 4));
    v3.c = Int_val(Field(_v3, 5));

    v4.x = Double_val(Field(_v4, 0));
    v4.y = Double_val(Field(_v4, 1));
    v4.z = Double_val(Field(_v4, 2));
    v4.u = Double_val(Field(_v4, 3));
    v4.v = Double_val(Field(_v4, 4));
    v4.c = Int_val(Field(_v4, 5));

    int type;
    switch (Int_val(_type))
    {
        case  0: type = POLYTYPE_ATEX; break;
        case  1: type = POLYTYPE_ATEX_LIT; break;
        case  2: type = POLYTYPE_ATEX_MASK; break;
        case  3: type = POLYTYPE_ATEX_MASK_LIT; break;
        case  4: type = POLYTYPE_ATEX_MASK_TRANS; break;
        case  5: type = POLYTYPE_ATEX_TRANS; break;
        case  6: type = POLYTYPE_FLAT; break;
        case  7: type = POLYTYPE_GCOL; break;
        case  8: type = POLYTYPE_GRGB; break;
        case  9: type = POLYTYPE_PTEX; break;
        case 10: type = POLYTYPE_PTEX_LIT; break;
        case 11: type = POLYTYPE_PTEX_MASK; break;
        case 12: type = POLYTYPE_PTEX_MASK_LIT; break;
        case 13: type = POLYTYPE_PTEX_MASK_TRANS; break;
        case 14: type = POLYTYPE_PTEX_TRANS; break;
        default: caml_invalid_argument("quad3d_f");
    }
    quad3d_f( Bitmap_val(bmp), type, Bitmap_val(tex), &v1, &v2, &v3, &v4 );
    return Val_unit;
}
CAMLprim value ml_quad3d_f_bytecode(value * argv, int argn)
{
    return ml_quad3d_f_native(argv[0], argv[1], argv[2],
                              argv[3], argv[4], argv[5], argv[6]);
}

/* }}} */
/* {{{ quad3d */

CAMLprim value ml_quad3d_native( value bmp, value _type, value tex,
                                 value _v1, value _v2, value _v3, value _v4 )
{
    V3D v1, v2, v3, v4;

    v1.x = Fixed_val(Field(_v1, 0));
    v1.y = Fixed_val(Field(_v1, 1));
    v1.z = Fixed_val(Field(_v1, 2));
    v1.u = Fixed_val(Field(_v1, 3));
    v1.v = Fixed_val(Field(_v1, 4));
    v1.c = Int_val(Field(_v1, 5));

    v2.x = Fixed_val(Field(_v2, 0));
    v2.y = Fixed_val(Field(_v2, 1));
    v2.z = Fixed_val(Field(_v2, 2));
    v2.u = Fixed_val(Field(_v2, 3));
    v2.v = Fixed_val(Field(_v2, 4));
    v2.c = Int_val(Field(_v2, 5));

    v3.x = Fixed_val(Field(_v3, 0));
    v3.y = Fixed_val(Field(_v3, 1));
    v3.z = Fixed_val(Field(_v3, 2));
    v3.u = Fixed_val(Field(_v3, 3));
    v3.v = Fixed_val(Field(_v3, 4));
    v3.c = Int_val(Field(_v3, 5));

    v4.x = Fixed_val(Field(_v4, 0));
    v4.y = Fixed_val(Field(_v4, 1));
    v4.z = Fixed_val(Field(_v4, 2));
    v4.u = Fixed_val(Field(_v4, 3));
    v4.v = Fixed_val(Field(_v4, 4));
    v4.c = Int_val(Field(_v4, 5));

    int type;
    switch (Int_val(_type))
    {
        case  0: type = POLYTYPE_ATEX; break;
        case  1: type = POLYTYPE_ATEX_LIT; break;
        case  2: type = POLYTYPE_ATEX_MASK; break;
        case  3: type = POLYTYPE_ATEX_MASK_LIT; break;
        case  4: type = POLYTYPE_ATEX_MASK_TRANS; break;
        case  5: type = POLYTYPE_ATEX_TRANS; break;
        case  6: type = POLYTYPE_FLAT; break;
        case  7: type = POLYTYPE_GCOL; break;
        case  8: type = POLYTYPE_GRGB; break;
        case  9: type = POLYTYPE_PTEX; break;
        case 10: type = POLYTYPE_PTEX_LIT; break;
        case 11: type = POLYTYPE_PTEX_MASK; break;
        case 12: type = POLYTYPE_PTEX_MASK_LIT; break;
        case 13: type = POLYTYPE_PTEX_MASK_TRANS; break;
        case 14: type = POLYTYPE_PTEX_TRANS; break;
        default: caml_invalid_argument("quad3d");
    }
    quad3d( Bitmap_val(bmp), type, Bitmap_val(tex), &v1, &v2, &v3, &v4 );
    return Val_unit;
}
CAMLprim value ml_quad3d_bytecode(value * argv, int argn)
{
    return ml_quad3d_native(argv[0], argv[1], argv[2],
                            argv[3], argv[4], argv[5], argv[6]);
}

/* }}} */


CAMLprim value ml_clear_scene( value bmp )
{
    clear_scene( Bitmap_val(bmp) );
    return Val_unit;
}

CAMLprim value ml_render_scene( value unit )
{
    render_scene();
    return Val_unit;
}

CAMLprim value ml_create_scene( value nedge, value npoly )
{
    return Val_int( create_scene( Int_val(nedge), Int_val(npoly) ) );
}

CAMLprim value ml_destroy_scene( value unit )
{
    destroy_scene();
    return Val_unit;
}



#define Val_zbuffer(b) ((value) b)
#define Zbuffer_val(v) ((ZBUFFER *) v)


CAMLprim value ml_create_zbuffer( value bmp )
{
    return Val_zbuffer( create_zbuffer( Bitmap_val(bmp) ) );
}

CAMLprim value ml_create_sub_zbuffer( value parent, value x, value y, value width, value height )
{
    return Val_zbuffer( create_sub_zbuffer( Zbuffer_val(parent), Int_val(x), Int_val(y),
                                            Int_val(width), Int_val(height) ) );
}

CAMLprim value ml_set_zbuffer( value zbuf )
{
    set_zbuffer( Zbuffer_val(zbuf) );
    return Val_unit;
}

CAMLprim value ml_clear_zbuffer( value zbuf, value z )
{
    clear_zbuffer( Zbuffer_val(zbuf), Double_val(z) );
    return Val_unit;
}

CAMLprim value ml_destroy_zbuffer( value zbuf)
{
    destroy_zbuffer( Zbuffer_val(zbuf) );
    return Val_unit;
}


/* 3D math routines */

#define Val_matrix(m) ((value) m)
#define Matrix_val(v) ((MATRIX *) v)

#define Val_matrix_f(m) ((value) m)
#define Matrix_f_val(v) ((MATRIX_f *) v)


CAMLprim value ml_get_identity_matrix( value unit )
{
    MATRIX *m;
    m = malloc(sizeof(MATRIX));
    memcpy( (void*) m, (void*) &identity_matrix, sizeof(MATRIX) );
    return Val_matrix(m);
}

CAMLprim value ml_get_identity_matrix_f( value unit )
{
    MATRIX_f *m;
    m = malloc(sizeof(MATRIX_f));
    memcpy( (void*) m, (void*) &identity_matrix_f, sizeof(MATRIX_f) );
    return Val_matrix_f(m);
}

CAMLprim value ml_free_matrix( value m )
{
    free((void *)m);
    return Val_unit;
}


CAMLprim value ml_new_matrix( value unit )
{
    MATRIX *m;
    m = malloc(sizeof(MATRIX));
    return Val_matrix(m);
}

CAMLprim value ml_new_matrix_f( value unit )
{
    MATRIX_f *m;
    m = malloc(sizeof(MATRIX_f));
    return Val_matrix_f(m);
}

CAMLprim value ml_make_matrix_f( value v, value t )
{
    MATRIX_f *m;
    m = malloc(sizeof(MATRIX_f));

    m->v[0][0] = Double_val(Field(Field(v, 0), 0));
    m->v[0][1] = Double_val(Field(Field(v, 1), 0));
    m->v[0][2] = Double_val(Field(Field(v, 2), 0));

    m->v[1][0] = Double_val(Field(Field(v, 0), 1));
    m->v[1][1] = Double_val(Field(Field(v, 1), 1));
    m->v[1][2] = Double_val(Field(Field(v, 2), 1));

    m->v[2][0] = Double_val(Field(Field(v, 0), 2));
    m->v[2][1] = Double_val(Field(Field(v, 1), 2));
    m->v[2][2] = Double_val(Field(Field(v, 2), 2));


    m->t[0] = Double_val(Field(t, 0));
    m->t[1] = Double_val(Field(t, 1));
    m->t[2] = Double_val(Field(t, 2));

    return Val_matrix_f(m);
}

CAMLprim value ml_make_matrix( value v, value t )
{
    MATRIX *m;
    m = malloc(sizeof(MATRIX));

    m->v[0][0] = Fixed_val(Field(Field(v, 0), 0));
    m->v[0][1] = Fixed_val(Field(Field(v, 1), 0));
    m->v[0][2] = Fixed_val(Field(Field(v, 2), 0));

    m->v[1][0] = Fixed_val(Field(Field(v, 0), 1));
    m->v[1][1] = Fixed_val(Field(Field(v, 1), 1));
    m->v[1][2] = Fixed_val(Field(Field(v, 2), 1));

    m->v[2][0] = Fixed_val(Field(Field(v, 0), 2));
    m->v[2][1] = Fixed_val(Field(Field(v, 1), 2));
    m->v[2][2] = Fixed_val(Field(Field(v, 2), 2));


    m->t[0] = Fixed_val(Field(t, 0));
    m->t[1] = Fixed_val(Field(t, 1));
    m->t[2] = Fixed_val(Field(t, 2));

    return Val_matrix(m);
}


CAMLprim value ml_get_translation_matrix( value m, value x, value y, value z )
{
    get_translation_matrix( Matrix_val(m), Fixed_val(x), Fixed_val(y), Fixed_val(z) );
    return Val_unit;
}
CAMLprim value ml_get_translation_matrix_f( value m, value x, value y, value z )
{
    get_translation_matrix_f( Matrix_f_val(m), Double_val(x), Double_val(y), Double_val(z) );
    return Val_unit;
}

CAMLprim value ml_get_scaling_matrix( value m, value x, value y, value z )
{
    get_scaling_matrix( Matrix_val(m), Fixed_val(x), Fixed_val(y), Fixed_val(z) );
    return Val_unit;
}
CAMLprim value ml_get_scaling_matrix_f( value m, value x, value y, value z )
{
    get_scaling_matrix_f( Matrix_f_val(m), Double_val(x), Double_val(y), Double_val(z) );
    return Val_unit;
}

CAMLprim value ml_get_x_rotate_matrix( value m, value r )
{
    get_x_rotate_matrix( Matrix_val(m), Fixed_val(r) );
    return Val_unit;
}
CAMLprim value ml_get_x_rotate_matrix_f( value m, value r )
{
    get_x_rotate_matrix_f( Matrix_f_val(m), Double_val(r) );
    return Val_unit;
}

CAMLprim value ml_get_y_rotate_matrix( value m, value r )
{
    get_y_rotate_matrix( Matrix_val(m), Fixed_val(r) );
    return Val_unit;
}
CAMLprim value ml_get_y_rotate_matrix_f( value m, value r )
{
    get_y_rotate_matrix_f( Matrix_f_val(m), Double_val(r) );
    return Val_unit;
}

CAMLprim value ml_get_z_rotate_matrix( value m, value r )
{
    get_z_rotate_matrix( Matrix_val(m), Fixed_val(r) );
    return Val_unit;
}
CAMLprim value ml_get_z_rotate_matrix_f( value m, value r )
{
    get_z_rotate_matrix_f( Matrix_f_val(m), Double_val(r) );
    return Val_unit;
}

CAMLprim value ml_get_rotation_matrix( value m, value x, value y, value z )
{
    get_rotation_matrix( Matrix_val(m), Fixed_val(x), Fixed_val(y), Fixed_val(z) );
    return Val_unit;
}
CAMLprim value ml_get_rotation_matrix_f( value m, value x, value y, value z )
{
    get_rotation_matrix_f( Matrix_f_val(m), Double_val(x), Double_val(y), Double_val(z) );
    return Val_unit;
}



CAMLprim value ml_get_align_matrix_f_native( value m, value xfront, value yfront, value zfront, value xup, value yup, value zup)
{
    get_align_matrix_f( Matrix_f_val(m), Double_val(xfront), Double_val(yfront), Double_val(zfront), Double_val(xup), Double_val(yup), Double_val(zup) );
    return Val_unit;
}
CAMLprim value ml_get_align_matrix_f_bytecode(value * argv, int argn)
{
    return ml_get_align_matrix_f_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5], argv[6]);
}

CAMLprim value ml_get_vector_rotation_matrix_f( value m, value x, value y, value z, value a )
{
    get_vector_rotation_matrix_f( Matrix_f_val(m), Double_val(x), Double_val(y), Double_val(z), Double_val(a) );
    return Val_unit;
}

CAMLprim value ml_get_transformation_matrix_f_native( value m, value scale, value xrot, value yrot, value zrot, value x, value y, value z )
{
    get_transformation_matrix_f( Matrix_f_val(m), Double_val(scale), Double_val(xrot), Double_val(yrot), Double_val(zrot), Double_val(x), Double_val(y), Double_val(z) );
    return Val_unit;
}
CAMLprim value ml_get_transformation_matrix_f_bytecode(value * argv, int argn)
{
    return ml_get_transformation_matrix_f_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5], argv[6], argv[7]);
}

CAMLprim value ml_get_camera_matrix_f_native( value m, value x, value y, value z,
                         value xfront, value yfront, value zfront, value xup, value yup, value zup, value fov, value aspect )
{
    get_camera_matrix_f( Matrix_f_val(m), Double_val(x), Double_val(y), Double_val(z), Double_val(xfront), Double_val(yfront), Double_val(zfront), Double_val(xup), Double_val(yup), Double_val(zup), Double_val(fov), Double_val(aspect) );
    return Val_unit;
}
CAMLprim value ml_get_camera_matrix_f_bytecode(value * argv, int argn)
{
    return ml_get_camera_matrix_f_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5], argv[6], argv[7], argv[8], argv[9], argv[10], argv[11]);
}

CAMLprim value ml_qtranslate_matrix_f( value m, value x, value y, value z )
{
    qtranslate_matrix_f( Matrix_f_val(m), Double_val(x), Double_val(y), Double_val(z) );
    return Val_unit;
}

CAMLprim value ml_qscale_matrix_f( value m, value scale )
{
    qscale_matrix_f( Matrix_f_val(m), Double_val(scale) );
    return Val_unit;
}



CAMLprim value ml_get_align_matrix_native( value m, value xfront, value yfront, value zfront, value xup, value yup, value zup)
{
    get_align_matrix( Matrix_val(m), Fixed_val(xfront), Fixed_val(yfront), Fixed_val(zfront), Fixed_val(xup), Fixed_val(yup), Fixed_val(zup) );
    return Val_unit;
}
CAMLprim value ml_get_align_matrix_bytecode(value * argv, int argn)
{
    return ml_get_align_matrix_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5], argv[6]);
}

CAMLprim value ml_get_vector_rotation_matrix( value m, value x, value y, value z, value a )
{
    get_vector_rotation_matrix( Matrix_val(m), Fixed_val(x), Fixed_val(y), Fixed_val(z), Fixed_val(a) );
    return Val_unit;
}

CAMLprim value ml_get_transformation_matrix_native( value m, value scale, value xrot, value yrot, value zrot, value x, value y, value z )
{
    get_transformation_matrix( Matrix_val(m), Fixed_val(scale), Fixed_val(xrot), Fixed_val(yrot), Fixed_val(zrot), Fixed_val(x), Fixed_val(y), Fixed_val(z) );
    return Val_unit;
}
CAMLprim value ml_get_transformation_matrix_bytecode(value * argv, int argn)
{
    return ml_get_transformation_matrix_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5], argv[6], argv[7]);
}

CAMLprim value ml_get_camera_matrix_native( value m, value x, value y, value z,
                         value xfront, value yfront, value zfront, value xup, value yup, value zup, value fov, value aspect )
{
    get_camera_matrix( Matrix_val(m), Fixed_val(x), Fixed_val(y), Fixed_val(z), Fixed_val(xfront), Fixed_val(yfront), Fixed_val(zfront), Fixed_val(xup), Fixed_val(yup), Fixed_val(zup), Fixed_val(fov), Fixed_val(aspect) );
    return Val_unit;
}
CAMLprim value ml_get_camera_matrix_bytecode(value * argv, int argn)
{
    return ml_get_camera_matrix_native(argv[0], argv[1], argv[2], argv[3], argv[4], argv[5], argv[6], argv[7], argv[8], argv[9], argv[10], argv[11]);
}

CAMLprim value ml_qtranslate_matrix( value m, value x, value y, value z )
{
    qtranslate_matrix( Matrix_val(m), Fixed_val(x), Fixed_val(y), Fixed_val(z) );
    return Val_unit;
}

CAMLprim value ml_qscale_matrix( value m, value scale )
{
    qscale_matrix( Matrix_val(m), Fixed_val(scale) );
    return Val_unit;
}



CAMLprim value ml_matrix_mul( value m1, value m2, value out )
{
    matrix_mul( Matrix_val(m1), Matrix_val(m2), Matrix_val(out) );
    return Val_unit;
}
CAMLprim value ml_matrix_mul_f( value m1, value m2, value out )
{
    matrix_mul_f( Matrix_f_val(m1), Matrix_f_val(m2), Matrix_f_val(out) );
    return Val_unit;
}


CAMLprim value ml_vector_length( value x, value y, value z )
{
    return Val_fixed( vector_length( Fixed_val(x), Fixed_val(y), Fixed_val(z)) );
}
CAMLprim value ml_vector_length_f( value x, value y, value z )
{
    return caml_copy_double(
            vector_length_f( Double_val(x), Double_val(y), Double_val(z) ));
}


CAMLprim value ml_apply_matrix_f( value m, value x, value y, value z )
{
    CAMLparam4( m, x, y, z );
    CAMLlocal1( out );

    float xout, yout, zout;
    apply_matrix_f( Matrix_f_val(m), Double_val(x), Double_val(y), Double_val(z), &xout, &yout, &zout );

    out = caml_alloc(3, 0);

    Store_field( out, 0, caml_copy_double(xout) );
    Store_field( out, 1, caml_copy_double(yout) );
    Store_field( out, 2, caml_copy_double(zout) );

    CAMLreturn( out );
}

CAMLprim value ml_apply_matrix( value m, value x, value y, value z )
{
    CAMLparam4( m, x, y, z );
    CAMLlocal1( out );

    fixed xout, yout, zout;
    apply_matrix( Matrix_val(m), Fixed_val(x), Fixed_val(y), Fixed_val(z), &xout, &yout, &zout );

    out = caml_alloc(3, 0);

    Store_field( out, 0, Val_fixed(xout) );
    Store_field( out, 1, Val_fixed(yout) );
    Store_field( out, 2, Val_fixed(zout) );

    CAMLreturn( out );
}



CAMLprim value ml_set_projection_viewport( value x, value y, value w, value h )
{
    set_projection_viewport( Int_val(x), Int_val(y), Int_val(w), Int_val(h) );
    return Val_unit;
}


CAMLprim value ml_persp_project_f( value x, value y, value z )
{
    CAMLparam3( x, y, z );
    CAMLlocal1( out );

    float xout, yout;
    persp_project_f( Double_val(x), Double_val(y), Double_val(z), &xout, &yout );

    out = caml_alloc(2, 0);

    Store_field( out, 0, caml_copy_double(xout) );
    Store_field( out, 1, caml_copy_double(yout) );

    CAMLreturn( out );
}

CAMLprim value ml_persp_project( value x, value y, value z )
{
    CAMLparam3( x, y, z );
    CAMLlocal1( out );

    fixed xout, yout;
    persp_project( Fixed_val(x), Fixed_val(y), Fixed_val(z), &xout, &yout );

    out = caml_alloc(2, 0);

    Store_field( out, 0, Val_fixed(xout) );
    Store_field( out, 1, Val_fixed(yout) );

    CAMLreturn( out );
}




/* Quaternion math routines */

#define Val_quat(q) ((value) q)
#define Quat_val(v) ((QUAT *) v)

CAMLprim value ml_make_quat( value w, value x, value y, value z )
{
    QUAT *q;
    q = malloc(sizeof(QUAT));

    q->w = Double_val(w);
    q->x = Double_val(x);
    q->y = Double_val(y);
    q->z = Double_val(z);

    return Val_quat(q);
}

CAMLprim value ml_free_quat( value q )
{
    free((void *)q);
    return Val_unit;
}

CAMLprim value ml_get_identity_quat( value unit )
{
    QUAT *q;
    q = malloc(sizeof(QUAT));

    q->w = identity_quat.w;
    q->x = identity_quat.x;
    q->y = identity_quat.y;
    q->z = identity_quat.z;

    return Val_quat(q);
}

CAMLprim value ml_get_x_rotate_quat( value q, value r ) { get_x_rotate_quat( Quat_val(q), Double_val(r) ); return Val_unit; }
CAMLprim value ml_get_y_rotate_quat( value q, value r ) { get_y_rotate_quat( Quat_val(q), Double_val(r) ); return Val_unit; }
CAMLprim value ml_get_z_rotate_quat( value q, value r ) { get_z_rotate_quat( Quat_val(q), Double_val(r) ); return Val_unit; }


CAMLprim value ml_get_rotation_quat( q, x, y, z )
{
    get_rotation_quat( Quat_val(q), Double_val(x), Double_val(y), Double_val(z) );
    return Val_unit;
}

CAMLprim value ml_get_vector_rotation_quat( q, x, y, z, a )
{
    get_vector_rotation_quat( Quat_val(q), Double_val(x), Double_val(y), Double_val(z), Double_val(a) );
    return Val_unit;
}

 
CAMLprim value ml_quat_to_matrix( value q )
{
    MATRIX_f *m;
    m = malloc(sizeof(MATRIX_f));
    quat_to_matrix( Quat_val(q), m );
    return Val_matrix_f(m);
}

CAMLprim value ml_matrix_to_quat( value m )
{
    QUAT *q;
    q = malloc(sizeof(QUAT));
    matrix_to_quat( Matrix_f_val(m), q );
    return Val_quat(q);
}


CAMLprim value ml_quat_mul( value p, value q )
{
    QUAT *out;
    out = malloc(sizeof(QUAT));
    quat_mul( Quat_val(p), Quat_val(q), out );
    return Val_quat(out);
}

CAMLprim value ml_apply_quat( value q, value x, value y, value z )
{
    CAMLparam4( q, x, y, z );
    CAMLlocal1( out );

    float xout, yout, zout;
    apply_quat( Quat_val(q), Double_val(x), Double_val(y), Double_val(z), &xout, &yout, &zout);


    out = caml_alloc(3, 0);

    Store_field( out, 0, caml_copy_double(xout) );
    Store_field( out, 1, caml_copy_double(yout) );
    Store_field( out, 2, caml_copy_double(zout) );

    CAMLreturn( out );
}

CAMLprim value ml_quat_interpolate( value from, value to, value t )
{
    QUAT *out;
    out = malloc(sizeof(QUAT));

    quat_interpolate( Quat_val(from), Quat_val(to), Double_val(t), out );

    return Val_quat(out);
}


CAMLprim value ml_quat_slerp( value from, value to, value t, value _how )
{
    QUAT *out;
    int how;
    out = malloc(sizeof(QUAT));

    switch (Int_val(_how))
    {
        case 0: how = QUAT_SHORT; break;
        case 1: how = QUAT_LONG; break;
        case 2: how = QUAT_CW; break;
        case 3: how = QUAT_CCW; break;
        case 4: how = QUAT_USER; break;
    }

    quat_slerp( Quat_val(from), Quat_val(to), Double_val(t), out, how );

    return Val_quat(out);
}


int main(int argc, char **argv)
{
    caml_main(argv);
    return 0;
}
END_OF_MAIN()


/* vim: sw=4 sts=4 ts=4 et fdm=marker
 */
