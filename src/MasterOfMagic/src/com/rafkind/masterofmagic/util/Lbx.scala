package com.rafkind.masterofmagic.util

import scala.collection.immutable.TreeMap
import java.io._

/*
 * This file should implement some file i/o to read .lbx files
 * LBX Format:
 * LBX seems to use a little endian encoding. So if '2003' is read as a 16-bit
 * number then byte 1 is '20' and byte 2 is '03' but its reversed so the real
 * number is 0320.
 *
 * byte = 8 bits
 * word = 16 bits
 * dword = 32 bits
 *
 * The number next to the size is the number of bytes in the file.
 *
 * Header:
 *  word(0) = number of files in the lbx archive
 *  dword(2-5) = signature, usually 'adfe0000'
 *  word(6-7) = version?
 *  dword(8-11) = offset of first file.
 *  dword(12+n..12+n+4) where n is the index of the file (-1 since I already
 *    accounted for the first file at offset 8)
 *
 * File contents:
 *   ???
 *
 * Example of armylist.lbx: $ xxd armylist.lbx
 * 0000000: 0900 adfe 0000 0000 2003 0000 70fc 0000 
 * 0000010: 7cfd 0000 88fe 0000 8204 0100 720a 0100
 * 0000020: 56df 0100 b3e0 0100 83e9 0100 4cf2 0100
 *
 * Number of files: 9
 * Signature: adfe0000
 * Version: 0000
 * File 1: 0x320
 * File 2: 0xfc70
 * File 3: 0xfd7c
 * File 4: 0xfe88
 * File 5: 0x10482
 * File 6: 0x10a72
 * File 7: 0x1df56
 * File 8: 0x1e0b3
 * File 9: 0x1e983
 */

/*
(*
open ExtLib
*)

type offsets = Offset of int * int;; 

type lbxheader = {
  offsets : offsets list;
};;

(* just a list of bytes *)
type lbxfile = {
  id : int;
  data : int list;
};;

(* Reads bytes in little-endian format *)
let read_bytes bytes input =
  let read byte = input_byte input in
  (* little endian, so reverse the numbers *)
  let bytes = List.rev (ExtList.List.of_enum (Enum.map read (ExtList.List.enum
  (ExtList.List.make
  bytes 0)))) in
  (* List.iter (fun byte -> Printf.printf "Got byte %d\n" byte) bytes; *)
  (* Add all the numbers up. For a 4 byte number it would be:
   *   n4 * 256^3 + n3 * 256^2 + n2 * 256^1 + n1 * 256^0
   *)
  List.fold_left (fun total now -> total * 256 + now) 0 bytes
;;

(* curried functions *)
let read_bytes_2 : in_channel -> int = read_bytes 2;;
let read_bytes_4 : in_channel -> int = read_bytes 4;;

(*
let inject func times =
  ExtList.List.of_enum (Enum.map func (ExtList.List.enum
  (ExtList.List.make times 0)))
;;
*)

let read_lbx_file input id offset : lbxfile =
  match offset with
  | Offset (start_offset, end_offset) -> begin
    (*
    Printf.printf "Read lbx file from %x to %x\n" start_offset end_offset;
    *)
    seek_in input start_offset;
    let bytes = Utils.inject (fun _ -> input_byte input) (end_offset - start_offset) in
    (*
    Printf.printf " Read %d bytes\n" (List.length bytes);
    *)
    {id = id; data = bytes}
  end
;;

(* this seems to process the list backwards, but im not sure why exactly *)
let rec do_successive_pairs (things : int list) (doer : int -> int -> 'a) : 'a list =
  match things with
  (* base case *)
  | a :: b :: [] -> [(doer a b)]
  (* recursive cases *)
  | a :: b :: rest -> (doer a b) :: (do_successive_pairs (b :: rest) doer)
  (* failure case *)
  | _ :: [] | [] -> raise (Failure "Need more than 1 pair")
;;

(* Read the LBX header. See above *)
let read_header input =
  let f : int = read_bytes_2 input in
  (*
  Printf.printf "Number of files: %d\n" f;
  Printf.printf "Signature: %d\n" (read_bytes_4 input);
  Printf.printf "Version: %d\n" (read_bytes_2 input);
  *)
  let signature = read_bytes_4 input in
  let version = read_bytes_2 input in
  let offsets = Utils.inject (fun _ -> (read_bytes_4 input)) f in
  (*
  for i = 0 to f do
    Printf.printf "Next file is: %x\n" (read_bytes_4 input);
  done;
  List.iter (fun offset -> Printf.printf "Next file is %x\n" offset) offsets;
  *)
  (*
  let lbxfiles = do_successive_pairs offsets (fun from xto -> read_lbx_file input from xto) in
  *)
  {offsets = (do_successive_pairs offsets (fun from xto -> Offset (from, xto)))}
;;

(* Read an LBX archive. `file' is the filename *)
let read_lbx file : lbxfile list =
  (*
  Printf.printf "Reading file %s\n" file;
  *)
  let input = open_in_bin file in
  (*
  Printf.printf "Opened file\n";
  *)
  let header = read_header input in
  let id = ref 0 in
  let lbxfiles = List.map (fun offset -> id := !id + 1; read_lbx_file input !id offset) header.offsets in
  close_in input;
  lbxfiles
;;

(*
Printf.printf "Lbx reader\n";;
read_lbx Sys.argv.(1);
*)
*/

class Data{
}

class LbxReader(path:String){
  val files:Map[Int, Data] = readLbx(path)

  def readLbx(path:String):Map[Int, Data] = {
    val file:File = new File(path)
    val reader:BufferedReader = new BufferedReader(new InputStreamReader(new FileInputStream(file)))
    reader.close()
    new TreeMap[Int, Data]()
  }
}
