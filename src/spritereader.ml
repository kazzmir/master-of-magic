type graphics_header = {
  width : int;
  height : int;
  unknown1 : int;
  bitmap_count : int;
  unknown2 : int;
  unknown3 : int;
  unknown4 : int;
  palette_info_offset : int;
  unknown5 : int
}

type graphics_palette = {
  palette_offset : int;
  first_palette_color_index : int;
  palette_color_count : int;
  unknown : int
}

type palette_entry = {
  red : int;
  green : int;
  blue : int
}

(*
 * TGfxPalette = Array [0..255] of TColor;
 *)
type palette = palette_entry list;;

let default_palette : palette =
 let make r g b = {red = r; green = g; blue = b;} in
 [(make 0x0  0x0  0x0);
 (make 0x8  0x4  0x4);
 (make 0x24 0x1c 0x18);
 (make 0x38 0x30 0x2c);
 (make 0x48 0x40 0x3c);
 (make 0x58 0x50 0x4c);
 (make 0x68 0x60 0x5c);
 (make 0x7c 0x74 70);
 (make 0x8c 0x84 0x80);
 (make 0x9c 0x94 0x90);
 (make 0xac 0xa4 0xa0);
 (make 0xc0 0xb8 0xb4);
 (make 0xd0 0xc8 0xc4);
 (make 0xe0 0xd8 0xd4);
 (make 0xf0 0xe8 0xe4);
 (make 0xfc 0xfc 0xfc);
 (make 0x38 0x20 0x1c);
 (make 0x40 0x2c 0x24);
 (make 0x48 0x34 0x2c);
 (make 0x50 0x3c 0x30);
 (make 0x58 0x40 0x34);
 (make 0x5c 0x44 0x38);
 (make 0x60 0x48 0x3c);
 (make 0x64 0x4c 0x3c);
 (make 0x68 0x50 0x40);
 (make 0x70 0x54 0x44);
 (make 0x78 0x5c 0x4c);
 (make 0x80 0x64 0x50);
 (make 0x8c 0x70 0x58);
 (make 0x94 0x74 0x5c);
 (make 0x9c 0x7c 0x64);
 (make 0xa4 0x84 0x68);
 (make 0xec 0xc0 0xd4);
 (make 0xd4 0x98 0xb4);
 (make 0xbc 0x74 0x94);
 (make 0xa4 0x54 0x7c);
 (make 0x8c 0x38 0x60);
 (make 0x74 0x24 0x4c);
 (make 0x5c 0x10 0x34);
 (make 0x44 0x4 0x20);
 (make 0xec 0xc0 0xc0);
 (make 0xd4 0x94 0x94);
 (make 0xbc 0x74 0x74);
 (make 0xa4 0x54 0x54);
 (make 0x8c 0x38 0x38);
 (make 0x74 0x24 0x24);
 (make 0x5c 0x10 0x10);
 (make 0x44 0x4 0x4);
 (make 0xec 0xd4 0xc0);
 (make 0xd4 0xb4 0x98);
 (make 0xbc 0x98 0x74);
 (make 0xa4 0x7c 0x54);
 (make 0x8c 0x60 0x38);
 (make 0x74 0x4c 0x24);
 (make 0x5c 0x34 0x10);
 (make 0x44 0x24 0x4);
 (make 0xec 0xec 0xc0);
 (make 0xd4 0xd4 0x94);
 (make 0xbc 0xbc 0x74);
 (make 0xa4 0xa4 0x54);
 (make 0x8c 0x8c 0x38);
 (make 0x74 0x74 0x24);
 (make 0x5c 0x5c 0x10);
 (make 0x44 0x44 0x4);
 (make 0xd4 0xec 0xbc);
 (make 0xb8 0xd4 0x98);
 (make 0x98 0xbc 0x74);
 (make 0x7c 0xa4 0x54);
 (make 0x60 0x8c 0x38);
 (make 0x4c 0x74 0x24);
 (make 0x38 0x5c 0x10);
 (make 0x24 0x44 0x4);
 (make 0xc0 0xec 0xc0);
 (make 0x98 0xd4 0x98);
 (make 0x74 0xbc 0x74);
 (make 0x54 0xa4 0x54);
 (make 0x38 0x8c 0x38);
 (make 0x24 0x74 0x24);
 (make 0x10 0x5c 0x10);
 (make 0x4 0x44 0x4);
 (make 0xc0 0xec 0xd8);
 (make 0x98 0xd4 0xb8);
 (make 0x74 0xbc 0x98);
 (make 0x54 0xa4 0x7c);
 (make 0x38 0x8c 0x60);
 (make 0x24 0x74 0x4c);
 (make 0x10 0x5c 0x38);
 (make 0x4 0x44 0x24);
 (make 0xf4 0xc0 0xa0);
 (make 0xe0 0xa0 0x84);
 (make 0xcc 0x84 0x6c);
 (make 0xc8 0x8c 0x68);
 (make 0xa8 0x78 0x54);
 (make 0x98 0x68 0x4c);
 (make 0x8c 0x60 0x44);
 (make 0x7c 0x50 0x3c);
 (make 0xc0 0xd8 0xec);
 (make 0x94 0xb4 0xd4);
 (make 0x70 0x98 0xbc);
 (make 0x54 0x7c 0xa4);
 (make 0x38 0x64 0x8c);
 (make 0x24 0x4c 0x74);
 (make 0x10 0x38 0x5c);
 (make 0x4 0x24 0x44);
 (make 0xc0 0xc0 0xec);
 (make 0x98 0x98 0xd4);
 (make 0x74 0x74 0xbc);
 (make 0x54 0x54 0xa4);
 (make 0x3c 0x38 0x8c);
 (make 0x24 0x24 0x74);
 (make 0x10 0x10 0x5c);
 (make 0x4 0x4 0x44);
 (make 0xd8 0xc0 0xec);
 (make 0xb8 0x98 0xd4);
 (make 0x98 0x74 0xbc);
 (make 0x7c 0x54 0xa4);
 (make 0x60 0x38 0x8c);
 (make 0x4c 0x24 0x74);
 (make 0x38 0x10 0x5c);
 (make 0x24 0x4 0x44);
 (make 0xec 0xc0 0xec);
 (make 0xd4 0x98 0xd4);
 (make 0xbc 0x74 0xbc);
 (make 0xa4 0x54 0xa4);
 (make 0x8c 0x38 0x8c);
 (make 0x74 0x24 0x74);
 (make 0x5c 0x10 0x5c);
 (make 0x44 0x4 0x44);
 (make 0xd8 0xd0 0xd0);
 (make 0xc0 0xb0 0xb0);
 (make 0xa4 0x90 0x90);
 (make 0x8c 0x74 0x74);
 (make 0x78 0x5c 0x5c);
 (make 0x68 0x4c 0x4c);
 (make 0x5c 0x3c 0x3c);
 (make 0x48 0x2c 0x2c);
 (make 0xd0 0xd8 0xd0);
 (make 0xb0 0xc0 0xb0);
 (make 0x90 0xa4 0x90);
 (make 0x74 0x8c 0x74);
 (make 0x5c 0x78 0x5c);
 (make 0x4c 0x68 0x4c);
 (make 0x3c 0x5c 0x3c);
 (make 0x2c 0x48 0x2c);
 (make 0xc8 0xc8 0xd4);
 (make 0xb0 0xb0 0xc0);
 (make 0x90 0x90 0xa4);
 (make 0x74 0x74 0x8c);
 (make 0x5c 0x5c 0x78);
 (make 0x4c 0x4c 0x68);
 (make 0x3c 0x3c 0x5c);
 (make 0x2c 0x2c 0x48);
 (make 0xd8 0xdc 0xec);
 (make 0xc8 0xcc 0xdc);
 (make 0xb8 0xc0 0xd4);
 (make 0xa8 0xb8 0xcc);
 (make 0x9c 0xb0 0xcc);
 (make 0x94 0xac 0xcc);
 (make 0x88 0xa4 0xcc);
 (make 0x88 0x94 0xdc);
 (make 0xfc 0xf0 0x90);
 (make 0xfc 0xe4 0x60);
 (make 0xfc 0xc8 0x24);
 (make 0xfc 0xac 0xc);
 (make 0xfc 0x78 0x10);
 (make 0xd0 0x1c 0x0);
 (make 0x98 0x0 0x0);
 (make 0x58 0x0 0x0);
 (make 0x90 0xf0 0xfc);
 (make 0x60 0xe4 0xfc);
 (make 0x24 0xc8 0xfc);
 (make 0xc 0xac 0xfc);
 (make 0x10 0x78 0xfc);
 (make 0x0 0x1c 0xd0);
 (make 0x0 0x0 0x98);
 (make 0x0 0x0 0x58);
 (make 0xfc 0xc8 0x64);
 (make 0xfc 0xb4 0x2c);
 (make 0xec 0xa4 0x24);
 (make 0xdc 0x94 0x1c);
 (make 0xcc 0x88 0x18);
 (make 0xbc 0x7c 0x14);
 (make 0xa4 0x6c 0x1c);
 (make 0x8c 0x60 0x24);
 (make 0x78 0x54 0x24);
 (make 0x60 0x44 0x24);
 (make 0x48 0x38 0x24);
 (make 0x34 0x28 0x1c);
 (make 0x90 0x68 0x34);
 (make 0x90 0x64 0x2c);
 (make 0x94 0x6c 0x34);
 (make 0x94 0x70 0x40);
 (make 0x8c 0x5c 0x24);
 (make 0x90 0x64 0x2c);
 (make 0x90 0x68 0x30);
 (make 0x98 0x78 0x4c);
 (make 0x60 0x3c 0x2c);
 (make 0x54 0xa4 0xa4);
 (make 0xc0 0x0 0x0);
 (make 0xfc 0x88 0xe0);
 (make 0xfc 0x58 0x84);
 (make 0xf4 0x0 0xc);
 (make 0xd4 0x0 0x0);
 (make 0xac 0x0 0x0);
 (make 0xe8 0xa8 0xfc);
 (make 0xe0 0x7c 0xfc);
 (make 0xd0 0x3c 0xfc);
 (make 0xc4 0x0 0xfc);
 (make 0x90 0x0 0xbc);
 (make 0xfc 0xf4 0x7c);
 (make 0xfc 0xe4 0x0);
 (make 0xe4 0xd0 0x0);
 (make 0xa4 0x98 0x0);
 (make 0x64 0x58 0x0);
 (make 0xac 0xfc 0xa8);
 (make 0x74 0xe4 0x70);
 (make 0x0 0xbc 0x0);
 (make 0x0 0xa4 0x0);
 (make 0x0 0x7c 0x0);
 (make 0xac 0xa8 0xfc);
 (make 0x80 0x7c 0xfc);
 (make 0x0 0x0 0xfc);
 (make 0x0 0x0 0xbc);
 (make 0x0 0x0 0x7c);
 (make 0x30 0x30 0x50);
 (make 0x28 0x28 0x48);
 (make 0x24 0x24 0x40);
 (make 0x20 0x1c 0x38);
 (make 0x1c 0x18 0x34);
 (make 0x18 0x14 0x2c);
 (make 0x14 0x10 0x24);
 (make 0x10 0xc 0x20);
 (make 0xa0 0xa0 0xb4);
 (make 0x88 0x88 0xa4);
 (make 0x74 0x74 0x90);
 (make 0x60 0x60 0x80);
 (make 0x50 0x4c 0x70);
 (make 0x40 0x3c 0x60);
 (make 0x30 0x2c 0x50);
 (make 0x24 0x20 0x40);
 (make 0x18 0x14 0x30);
 (make 0x10 0xc 0x20);
 (make 0x14 0xc 0x8);
 (make 0x18 0x10 0xc);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0);
 (make 0x0 0x0 0x0)];;


(*

const
     {==========================================================================
      |
      | Constant    : MOM_PALETTE
      | Description : Default Master of Magic palette
      |
      =========================================================================}
     MOM_PALETTE: Array [0..255] of TGfxPaletteEntry =
       ((r: $0;  g: $0;  b: $0),
        (r: $8;  g: $4;  b: $4),
        (r: $24; g: $1c; b: $18),
        (r: $38; g: $30; b: $2c),
        (r: $48; g: $40; b: $3c),
        (r: $58; g: $50; b: $4c),
        (r: $68; g: $60; b: $5c),
        (r: $7c; g: $74; b: $70),
        (r: $8c; g: $84; b: $80),
        (r: $9c; g: $94; b: $90),
        (r: $ac; g: $a4; b: $a0),
        (r: $c0; g: $b8; b: $b4),
        (r: $d0; g: $c8; b: $c4),
        (r: $e0; g: $d8; b: $d4),
        (r: $f0; g: $e8; b: $e4),
        (r: $fc; g: $fc; b: $fc),
        (r: $38; g: $20; b: $1c),
        (r: $40; g: $2c; b: $24),
        (r: $48; g: $34; b: $2c),
        (r: $50; g: $3c; b: $30),
        (r: $58; g: $40; b: $34),
        (r: $5c; g: $44; b: $38),
        (r: $60; g: $48; b: $3c),
        (r: $64; g: $4c; b: $3c),
        (r: $68; g: $50; b: $40),
        (r: $70; g: $54; b: $44),
        (r: $78; g: $5c; b: $4c),
        (r: $80; g: $64; b: $50),
        (r: $8c; g: $70; b: $58),
        (r: $94; g: $74; b: $5c),
        (r: $9c; g: $7c; b: $64),
        (r: $a4; g: $84; b: $68),
        (r: $ec; g: $c0; b: $d4),
        (r: $d4; g: $98; b: $b4),
        (r: $bc; g: $74; b: $94),
        (r: $a4; g: $54; b: $7c),
        (r: $8c; g: $38; b: $60),
        (r: $74; g: $24; b: $4c),
        (r: $5c; g: $10; b: $34),
        (r: $44; g: $4;  b: $20),
        (r: $ec; g: $c0; b: $c0),
        (r: $d4; g: $94; b: $94),
        (r: $bc; g: $74; b: $74),
        (r: $a4; g: $54; b: $54),
        (r: $8c; g: $38; b: $38),
        (r: $74; g: $24; b: $24),
        (r: $5c; g: $10; b: $10),
        (r: $44; g: $4;  b: $4),
        (r: $ec; g: $d4; b: $c0),
        (r: $d4; g: $b4; b: $98),
        (r: $bc; g: $98; b: $74),
        (r: $a4; g: $7c; b: $54),
        (r: $8c; g: $60; b: $38),
        (r: $74; g: $4c; b: $24),
        (r: $5c; g: $34; b: $10),
        (r: $44; g: $24; b: $4),
        (r: $ec; g: $ec; b: $c0),
        (r: $d4; g: $d4; b: $94),
        (r: $bc; g: $bc; b: $74),
        (r: $a4; g: $a4; b: $54),
        (r: $8c; g: $8c; b: $38),
        (r: $74; g: $74; b: $24),
        (r: $5c; g: $5c; b: $10),
        (r: $44; g: $44; b: $4),
        (r: $d4; g: $ec; b: $bc),
        (r: $b8; g: $d4; b: $98),
        (r: $98; g: $bc; b: $74),
        (r: $7c; g: $a4; b: $54),
        (r: $60; g: $8c; b: $38),
        (r: $4c; g: $74; b: $24),
        (r: $38; g: $5c; b: $10),
        (r: $24; g: $44; b: $4),
        (r: $c0; g: $ec; b: $c0),
        (r: $98; g: $d4; b: $98),
        (r: $74; g: $bc; b: $74),
        (r: $54; g: $a4; b: $54),
        (r: $38; g: $8c; b: $38),
        (r: $24; g: $74; b: $24),
        (r: $10; g: $5c; b: $10),
        (r: $4;  g: $44; b: $4),
        (r: $c0; g: $ec; b: $d8),
        (r: $98; g: $d4; b: $b8),
        (r: $74; g: $bc; b: $98),
        (r: $54; g: $a4; b: $7c),
        (r: $38; g: $8c; b: $60),
        (r: $24; g: $74; b: $4c),
        (r: $10; g: $5c; b: $38),
        (r: $4;  g: $44; b: $24),
        (r: $f4; g: $c0; b: $a0),
        (r: $e0; g: $a0; b: $84),
        (r: $cc; g: $84; b: $6c),
        (r: $c8; g: $8c; b: $68),
        (r: $a8; g: $78; b: $54),
        (r: $98; g: $68; b: $4c),
        (r: $8c; g: $60; b: $44),
        (r: $7c; g: $50; b: $3c),
        (r: $c0; g: $d8; b: $ec),
        (r: $94; g: $b4; b: $d4),
        (r: $70; g: $98; b: $bc),
        (r: $54; g: $7c; b: $a4),
        (r: $38; g: $64; b: $8c),
        (r: $24; g: $4c; b: $74),
        (r: $10; g: $38; b: $5c),
        (r: $4;  g: $24; b: $44),
        (r: $c0; g: $c0; b: $ec),
        (r: $98; g: $98; b: $d4),
        (r: $74; g: $74; b: $bc),
        (r: $54; g: $54; b: $a4),
        (r: $3c; g: $38; b: $8c),
        (r: $24; g: $24; b: $74),
        (r: $10; g: $10; b: $5c),
        (r: $4;  g: $4;  b: $44),
        (r: $d8; g: $c0; b: $ec),
        (r: $b8; g: $98; b: $d4),
        (r: $98; g: $74; b: $bc),
        (r: $7c; g: $54; b: $a4),
        (r: $60; g: $38; b: $8c),
        (r: $4c; g: $24; b: $74),
        (r: $38; g: $10; b: $5c),
        (r: $24; g: $4;  b: $44),
        (r: $ec; g: $c0; b: $ec),
        (r: $d4; g: $98; b: $d4),
        (r: $bc; g: $74; b: $bc),
        (r: $a4; g: $54; b: $a4),
        (r: $8c; g: $38; b: $8c),
        (r: $74; g: $24; b: $74),
        (r: $5c; g: $10; b: $5c),
        (r: $44; g: $4;  b: $44),
        (r: $d8; g: $d0; b: $d0),
        (r: $c0; g: $b0; b: $b0),
        (r: $a4; g: $90; b: $90),
        (r: $8c; g: $74; b: $74),
        (r: $78; g: $5c; b: $5c),
        (r: $68; g: $4c; b: $4c),
        (r: $5c; g: $3c; b: $3c),
        (r: $48; g: $2c; b: $2c),
        (r: $d0; g: $d8; b: $d0),
        (r: $b0; g: $c0; b: $b0),
        (r: $90; g: $a4; b: $90),
        (r: $74; g: $8c; b: $74),
        (r: $5c; g: $78; b: $5c),
        (r: $4c; g: $68; b: $4c),
        (r: $3c; g: $5c; b: $3c),
        (r: $2c; g: $48; b: $2c),
        (r: $c8; g: $c8; b: $d4),
        (r: $b0; g: $b0; b: $c0),
        (r: $90; g: $90; b: $a4),
        (r: $74; g: $74; b: $8c),
        (r: $5c; g: $5c; b: $78),
        (r: $4c; g: $4c; b: $68),
        (r: $3c; g: $3c; b: $5c),
        (r: $2c; g: $2c; b: $48),
        (r: $d8; g: $dc; b: $ec),
        (r: $c8; g: $cc; b: $dc),
        (r: $b8; g: $c0; b: $d4),
        (r: $a8; g: $b8; b: $cc),
        (r: $9c; g: $b0; b: $cc),
        (r: $94; g: $ac; b: $cc),
        (r: $88; g: $a4; b: $cc),
        (r: $88; g: $94; b: $dc),
        (r: $fc; g: $f0; b: $90),
        (r: $fc; g: $e4; b: $60),
        (r: $fc; g: $c8; b: $24),
        (r: $fc; g: $ac; b: $c),
        (r: $fc; g: $78; b: $10),
        (r: $d0; g: $1c; b: $0),
        (r: $98; g: $0;  b: $0),
        (r: $58; g: $0;  b: $0),
        (r: $90; g: $f0; b: $fc),
        (r: $60; g: $e4; b: $fc),
        (r: $24; g: $c8; b: $fc),
        (r: $c;  g: $ac; b: $fc),
        (r: $10; g: $78; b: $fc),
        (r: $0;  g: $1c; b: $d0),
        (r: $0;  g: $0;  b: $98),
        (r: $0;  g: $0;  b: $58),
        (r: $fc; g: $c8; b: $64),
        (r: $fc; g: $b4; b: $2c),
        (r: $ec; g: $a4; b: $24),
        (r: $dc; g: $94; b: $1c),
        (r: $cc; g: $88; b: $18),
        (r: $bc; g: $7c; b: $14),
        (r: $a4; g: $6c; b: $1c),
        (r: $8c; g: $60; b: $24),
        (r: $78; g: $54; b: $24),
        (r: $60; g: $44; b: $24),
        (r: $48; g: $38; b: $24),
        (r: $34; g: $28; b: $1c),
        (r: $90; g: $68; b: $34),
        (r: $90; g: $64; b: $2c),
        (r: $94; g: $6c; b: $34),
        (r: $94; g: $70; b: $40),
        (r: $8c; g: $5c; b: $24),
        (r: $90; g: $64; b: $2c),
        (r: $90; g: $68; b: $30),
        (r: $98; g: $78; b: $4c),
        (r: $60; g: $3c; b: $2c),
        (r: $54; g: $a4; b: $a4),
        (r: $c0; g: $0;  b: $0),
        (r: $fc; g: $88; b: $e0),
        (r: $fc; g: $58; b: $84),
        (r: $f4; g: $0;  b: $c),
        (r: $d4; g: $0;  b: $0),
        (r: $ac; g: $0;  b: $0),
        (r: $e8; g: $a8; b: $fc),
        (r: $e0; g: $7c; b: $fc),
        (r: $d0; g: $3c; b: $fc),
        (r: $c4; g: $0;  b: $fc),
        (r: $90; g: $0;  b: $bc),
        (r: $fc; g: $f4; b: $7c),
        (r: $fc; g: $e4; b: $0),
        (r: $e4; g: $d0; b: $0),
        (r: $a4; g: $98; b: $0),
        (r: $64; g: $58; b: $0),
        (r: $ac; g: $fc; b: $a8),
        (r: $74; g: $e4; b: $70),
        (r: $0;  g: $bc; b: $0),
        (r: $0;  g: $a4; b: $0),
        (r: $0;  g: $7c; b: $0),
        (r: $ac; g: $a8; b: $fc),
        (r: $80; g: $7c; b: $fc),
        (r: $0;  g: $0;  b: $fc),
        (r: $0;  g: $0;  b: $bc),
        (r: $0;  g: $0;  b: $7c),
        (r: $30; g: $30; b: $50),
        (r: $28; g: $28; b: $48),
        (r: $24; g: $24; b: $40),
        (r: $20; g: $1c; b: $38),
        (r: $1c; g: $18; b: $34),
        (r: $18; g: $14; b: $2c),
        (r: $14; g: $10; b: $24),
        (r: $10; g: $c;  b: $20),
        (r: $a0; g: $a0; b: $b4),
        (r: $88; g: $88; b: $a4),
        (r: $74; g: $74; b: $90),
        (r: $60; g: $60; b: $80),
        (r: $50; g: $4c; b: $70),
        (r: $40; g: $3c; b: $60),
        (r: $30; g: $2c; b: $50),
        (r: $24; g: $20; b: $40),
        (r: $18; g: $14; b: $30),
        (r: $10; g: $c;  b: $20),
        (r: $14; g: $c;  b: $8),
        (r: $18; g: $10; b: $c),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0),
        (r: $0;  g: $0;  b: $0));

{===============================================================================
|
| Method      : SaveGfxAsBmps
| Description : Converts this graphics file inside the LBX file to multiple
|               BMPs
|
==============================================================================}
procedure TLBXFileContents.SaveGfxAsBmps (FileName: String);
var
   fin: TFileStream;
   BitmapNo, BitmapOffset, RLE_val, ColourNo, BitmapStart, BitmapEnd, x, y,
      BitmapSize, BitmapIndex, next_ctl, long_data, n_r, last_pos, RleLength,
      RleCounter: Integer;
   BitmapNumberString: String;
   GfxHeader: TGfxHeader;
   GfxPaletteInfo: TGfxPaletteInfo;
   GfxPaletteEntry: TGfxPaletteEntry;
   BitmapOffsets: TList;
   Palette: TGfxPalette;
   b: TBitmap;
   ImageBuffer: Array [0..65499] of Byte;
   ColourValue: TColor;
begin
     { Open LBX file }
     fin := TFileStream.Create (LBXFile.FileName, fmOpenRead or fmShareDenyWrite);
     try
        { Find start of graphics file }
        fin.Seek (FileOffset, soFromBeginning);

        { Read graphics header }
        fin.ReadBuffer (GfxHeader, SizeOf (TGfxHeader));

        { Read file offsets of each bitmap }
        BitmapOffsets := TList.Create;
        try
           { Not -1 since there is an extra offset specifying the end of the
             last image }
           for BitmapNo := 0 to GfxHeader.BitmapCount do
           begin
                fin.ReadBuffer (BitmapOffset, 4);
                BitmapOffsets.Add (Pointer (BitmapOffset));
           end;

           { Default palette }
           for ColourNo := 0 to 255 do
           begin
                GfxPaletteEntry := MOM_PALETTE [ColourNo];

                Palette [ColourNo] :=
                   (GfxPaletteEntry.b shl 16) +
                   (GfxPaletteEntry.g shl 8) +
                    GfxPaletteEntry.r;
           end;

           { Read palette info if present }
           if GfxHeader.PaletteInfoOffset > 0 then
           begin
                fin.Seek (FileOffset + GfxHeader.PaletteInfoOffset, soFromBeginning);
                fin.ReadBuffer (GfxPaletteInfo, SizeOf (TGfxPaletteInfo));

                { Read palette }
                fin.Seek (FileOffset + GfxPaletteInfo.PaletteOffset, soFromBeginning);
                for ColourNo := 0 to GfxPaletteInfo.PaletteColourCount - 1 do
                begin
                     fin.ReadBuffer (GfxPaletteEntry, SizeOf (TGfxPaletteEntry));

                     { Multiply colour values up by 4 }
                     Palette [GfxPaletteInfo.FirstPaletteColourIndex + ColourNo] :=
                        (GfxPaletteEntry.b shl 18) +
                        (GfxPaletteEntry.g shl 10) +
                        (GfxPaletteEntry.r shl 2);
                end;
           end else
           begin
                { No palette info, use defaults }
                GfxPaletteInfo.FirstPaletteColourIndex := 0;
                GfxPaletteInfo.PaletteColourCount      := 255;
           end;

           { Reuse the same bitmap for each image }
           b := TBitmap.Create;
           try
              b.Width  := GfxHeader.Width;
              b.Height := GfxHeader.Height;

              { Set background colour }
              for x := 0 to GfxHeader.Width - 1 do
                  for y := 0 to GfxHeader.Height - 1 do
                      b.Canvas.Pixels [x, y] := $FF00FF;

              { Values of at least this indicate run length values }
              RLE_val := GfxPaletteInfo.FirstPaletteColourIndex +
                         GfxPaletteInfo.PaletteColourCount;

              { Convert each bitmap }
              for BitmapNo := 0 to GfxHeader.BitmapCount - 1 do
              begin
                   BitmapStart := Integer (BitmapOffsets.Items [BitmapNo]);
                   BitmapEnd   := Integer (BitmapOffsets.Items [BitmapNo + 1]);
                   BitmapSize  := BitmapEnd - BitmapStart;

                   if BitmapSize > 65500 then
                      raise ELBXException.Create
                            ('Does not support encoded images larger than 65500 bytes, found image of size ' +
                             IntToStr (BitmapSize));

                   { Read in entire bitmap }
                   fin.Seek (FileOffset + BitmapStart, soFromBeginning);
                   fin.ReadBuffer (ImageBuffer, BitmapSize);

                   { Byte 0 tells us whether to reset the image half way
                     through an animation }
                   if (ImageBuffer [0] = 1) and (BitmapNo > 0) then
                      for x := 0 to GfxHeader.Width - 1 do
                          for y := 0 to GfxHeader.Height - 1 do
                              b.Canvas.Pixels [x, y] := $FF00FF;

                   { Decode bitmap }
                   BitmapIndex := 1; { Current index into the image buffer }
                   x := 0;
                   y := GfxHeader.Height;
                   next_ctl  := 0;
                   long_data := 0;
                   n_r       := 0;
                   last_pos  := 0;

                   while (x < GfxHeader.Width) and (BitmapIndex < BitmapSize) do
                   begin
                        y := 0;
                        if (ImageBuffer [BitmapIndex] = $FF) then
                        begin
                             inc (BitmapIndex);
                             RLE_val := GfxPaletteInfo.FirstPaletteColourIndex +
                                        GfxPaletteInfo.PaletteColourCount;
                        end else
                        begin
                             long_data := ImageBuffer [BitmapIndex + 2];
                             next_ctl  := BitmapIndex + ImageBuffer [BitmapIndex + 1] + 2;

                             case ImageBuffer [BitmapIndex] of
                                  $00: RLE_val := GfxPaletteInfo.FirstPaletteColourIndex +
                                                  GfxPaletteInfo.PaletteColourCount;
                                  $80: RLE_val := $E0;
                             else
                                 raise ELBXException.Create ('Unrecognized RLE value');
                             end;

                             y := ImageBuffer [BitmapIndex + 3];
                             inc (BitmapIndex, 4);

                             n_r := BitmapIndex;
                             while n_r < next_ctl do
                             begin
                                  while (n_r < BitmapIndex + long_data) and (x < GfxHeader.Width) do
                                  begin
                                       if (ImageBuffer [n_r] >= RLE_val) then
                                       begin
                                            { This value is an run length, the
                                              next value is the value to repeat }
                                            last_pos := n_r + 1;
                                            RleLength := ImageBuffer [n_r] - RLE_val + 1;
{                                               if (RleLength + y > GfxHeader.Height) then
                                               raise ELBXException.Create ('RLE length overrun on y');}

                                            RleCounter := 0;
                                            while (RleCounter < RleLength) and (y < GfxHeader.Height) do
                                            begin
                                                 if (x < GfxHeader.Width) and (y < GfxHeader.Height) and
                                                    (x >= 0) and (y >= 0) then
                                                 begin
                                                      ColourValue := Palette [ImageBuffer [last_pos]];
                                                      if ColourValue = $B4A0A0 then
                                                          b.Canvas.Pixels [x, y] := $00FF00
                                                      else
                                                          b.Canvas.Pixels [x, y] := ColourValue;
                                                 end else
                                                      raise ELBXException.Create ('RLE length overrun on output');

                                                 inc (y);
                                                 inc (RleCounter);
                                            end;
                                            inc (n_r, 2);
                                       end else
                                       begin
                                            { Regular single pixel }
                                            if (x < GfxHeader.Width) and (y < GfxHeader.Height) and
                                               (x >= 0) and (y >= 0) then
                                            begin
                                                 ColourValue := Palette [ImageBuffer [n_r]];
                                                 if ColourValue = $B4A0A0 then
                                                     b.Canvas.Pixels [x, y] := $00FF00
                                                 else
                                                     b.Canvas.Pixels [x, y] := ColourValue;
                                            end;
{                                               else
                                                raise ELBXException.Create ('Buffer overrun');}

                                            inc (n_r);
                                            inc (y);
                                       end;
                                  end;

                                  if n_r < next_ctl then
                                  begin
                                       { On se trouve sur un autre RLE sur la ligne
                                         Some others data are here }
                                       inc (y, ImageBuffer [n_r + 1]); { next pos Y to write pixels }
                                       BitmapIndex := n_r + 2;
                                       long_data := ImageBuffer [n_r]; { number of data to put }
                                       inc (n_r, 2);

{                                          if n_r >= next_ctl then
                                          raise ELBXException.Create ('More RLE but lines too short');}
                                  end;
                             end;

                             BitmapIndex := next_ctl; { jump to next line }
                        end;

                        inc (x);
                   end;

                   { Save bitmap }
                   BitmapNumberString := IntToStr (BitmapNo);
                   while Length (BitmapNumberString) < 3 do
                         BitmapNumberString := '0' + BitmapNumberString;

                   b.SaveToFile (ChangeFileExt (FileName, '') + '_' + BitmapNumberString + '.bmp');
              end;

           finally
              b.Free;
           end;

        finally
           BitmapOffsets.Free;
        end;

     finally
        fin.Free;
     end;
end;
*)
