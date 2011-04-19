(* Reads the images out of terrain.lbx.
 *
 * The terrain graphics in TERRAIN.LBX aren't in the same graphics format as the rest of the MoM graphics. The format is much simpler. 
 *
 * TERRAIN.LBX subfile 0 has a 192 byte header on it - I've no idea what this is for so skip it. 
 *
 * Each terrain tile is then made up of 
 * An 8 byte header 
 * The image data 
 * A 4 byte footer 
 *
 * The first byte of the header is the width - this is always 20. The second byte is the height - this is always 18. I've no idea what the remaining 6 bytes in the header or the 4 bytes in the footer are for. 
 *
 * The image data is therefore always 20 * 18 = 360 bytes long, so each image including the header and footer takes up 372 bytes. Each byte of image data is an index into the standard MoM palette (see code in post about standard graphics format for this palette). There are 1,761 images. 
 *)

let read_bytes bytes input =
  let rec loop all left =
    match left with
    | 0 -> List.rev all
    | n -> let item = List.hd !input in
           input := List.tl !input;
           loop (item :: all) (n - 1)
  in
  loop [] bytes
;;

let skip input bytes : unit =
  ignore(read_bytes bytes input)
;;

let join separator stuff =
  let rec loop all rest =
    match rest with
    | [] -> all
    | item :: more -> loop (all ^ separator ^ item) more
  in
  loop "" stuff
;;

let read_word words = read_bytes (2 * words);;

let read_image input : int list =
  let width = read_word 1 input in
  let height = read_word 1 input in
  let rest_header = read_word 6 input in
  Printf.printf "Read image %dx%d\n" (List.nth width 0) (List.nth height 0);
  (* Printf.printf " %s\n" (join " " (List.map (fun i -> Format.sprintf "%d" i) (List.append width (List.append height rest_header)))); *)
  let data = read_bytes ((List.nth width 0) * (List.nth height 0)) input in
  (* let data = read_bytes (20 * 18) input in *)
  let footer = read_word 4 input in
  data
;;

(* returns a list of images *)
let read bytes =
  let input = ref bytes in
  skip input 192; 
  (* skip first 192 bytes *)
  let rec loop images =
    match !input with
    | [] -> images
    | stuff -> loop ((read_image input) :: images)
  in
  loop []
;;

let main () =
  let file = "data/terrain.lbx" in
  let lbx_files = Lbxreader.read_lbx file in
  let images = read (List.nth lbx_files 0).Lbxreader.data in
  Printf.printf "Read %d images\n" (List.length images)
  (*
  Printf.printf "Read %d lbx files\n" (List.length lbx_files);
  List.iter (fun lbx -> Printf.printf "Lbx %d length %d\n" lbx.Lbxreader.id (List.length lbx.Lbxreader.data)) lbx_files
  *)
  (*
  let images = read file in
  *)
;;

main ()
