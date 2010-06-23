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


let lbxToSprite (lbx : Lbxreader.lbxfile) =
  let counter = ref 0 in
  let read n = begin
    let t = (List.nth lbx.Lbxreader.data n) in
    counter := !counter + 1;
    t
  end in
  0
;;

(*

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
